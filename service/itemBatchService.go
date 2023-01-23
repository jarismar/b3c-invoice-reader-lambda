package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/constants"
	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/store"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

type ItemBatchService struct {
	tx                *sql.Tx
	user              *entity.User
	invoice           *entity.Invoice
	invoiceItem       *entity.InvoiceItem
	taxStore          *store.TaxStore
	brokerTaxStore    *store.BrokerTaxStore
	companyBatchStore *store.CompanyBatchStore
}

func GetItemBatchService(
	tx *sql.Tx,
	user *entity.User,
	invoice *entity.Invoice,
	invoiceItem *entity.InvoiceItem,
	taxStore *store.TaxStore,
	brokerTaxStore *store.BrokerTaxStore,
	companyBatchStore *store.CompanyBatchStore,
) *ItemBatchService {
	return &ItemBatchService{
		tx:                tx,
		user:              user,
		invoice:           invoice,
		invoiceItem:       invoiceItem,
		taxStore:          taxStore,
		brokerTaxStore:    brokerTaxStore,
		companyBatchStore: companyBatchStore,
	}
}

func (ibsvc *ItemBatchService) updateCompanyBatch(
	currentCB *entity.CompanyBatch,
	persistedCB *entity.CompanyBatch,
) (*entity.CompanyBatch, error) {
	newQty := currentCB.Qty + persistedCB.Qty
	newTotalPrice := currentCB.TotalPrice + persistedCB.TotalPrice
	newAvgPrice := newTotalPrice / float64(newQty)

	newCompanyBatch := &entity.CompanyBatch{
		Id:         persistedCB.Id,
		User:       persistedCB.User,
		Company:    persistedCB.Company,
		StartDate:  persistedCB.StartDate,
		Qty:        newQty,
		AvgPrice:   newAvgPrice,
		TotalPrice: newTotalPrice,
	}

	companyBatchDAO := db.GetCompanyBatchDAO(
		ibsvc.tx,
		newCompanyBatch,
	)

	newCompanyBatchRec, err := companyBatchDAO.UpdateCompanyBatch()

	if err != nil {
		return nil, err
	}

	return newCompanyBatchRec, nil
}

func (ibsvc *ItemBatchService) getCompanyBatch(taxGroup *entity.TaxGroup) (*entity.CompanyBatch, error) {
	invoice := ibsvc.invoice
	invoiceItem := ibsvc.invoiceItem
	user := ibsvc.user
	store := ibsvc.companyBatchStore

	rawPrice := invoiceItem.Price * float64(invoiceItem.Qty)
	totalTax := utils.GetTotalTax(taxGroup)
	totalPrice := rawPrice + totalTax

	companyBatch := &entity.CompanyBatch{
		User:       user,
		Company:    invoiceItem.Company,
		StartDate:  invoice.MarketDate,
		Qty:        invoiceItem.Qty,
		AvgPrice:   totalPrice / float64(invoiceItem.Qty),
		TotalPrice: totalPrice,
	}

	if store.Has(companyBatch) {
		companyBatchRec := store.Get(companyBatch)

		log.Printf(
			"ItemBatchService.getCompanyBatch: found on cache [%d, %s, qty = %d, avg = %.4f]",
			companyBatchRec.Id,
			companyBatchRec.Company.Code,
			companyBatchRec.Qty,
			companyBatchRec.AvgPrice,
		)

		newCompanyBatchRec, err := ibsvc.updateCompanyBatch(
			companyBatch,
			companyBatchRec,
		)

		if err != nil {
			return nil, err
		}

		store.Put(newCompanyBatchRec)

		return newCompanyBatchRec, nil
	}

	companyBatchDAO := db.GetCompanyBatchDAO(
		ibsvc.tx,
		companyBatch,
	)

	companyBatchRec, err := companyBatchDAO.GetCompanyBatch()

	if err != nil {
		return nil, err
	}

	if companyBatchRec == nil {
		newCompanyBatchRec, err := companyBatchDAO.CreateCompanyBatch()

		if err != nil {
			return nil, err
		}

		store.Put(newCompanyBatchRec)
		return newCompanyBatchRec, nil
	}

	newCompanyBatchRec, err := ibsvc.updateCompanyBatch(
		companyBatch,
		companyBatchRec,
	)

	if err != nil {
		return nil, err
	}

	store.Put(newCompanyBatchRec)

	return newCompanyBatchRec, nil
}

func (ibsvc *ItemBatchService) getTaxValue(
	invoiceTaxInstance *entity.TaxInstance,
	itemPriceRate float64,
) float64 {
	taxCode := invoiceTaxInstance.Tax.Code
	taxTypes := constants.TaxTypes
	taxRates := constants.TaxRates
	item := ibsvc.invoiceItem

	if taxCode == taxTypes.BRKFEE {
		if ibsvc.brokerTaxStore.Has(item, taxCode) {
			return 0
		}

		if invoiceTaxInstance.TaxValue == 0 {
			return 0
		}

		ibsvc.brokerTaxStore.Put(item, taxCode)

		return taxRates.BRKFEE
	}

	if taxCode == taxTypes.ISSSPFEE {
		if ibsvc.brokerTaxStore.Has(item, taxCode) {
			return 0
		}

		if invoiceTaxInstance.TaxValue == 0 {
			return 0
		}

		ibsvc.brokerTaxStore.Put(item, taxCode)

		// t = (b / (1 - i)) - c
		return (taxRates.BRKFEE / (1 - taxRates.ISSSPFEE)) - taxRates.BRKFEE
	}

	if taxCode == taxTypes.IRRFFEE {
		return 0.0
	}

	return invoiceTaxInstance.TaxValue * itemPriceRate
}

func (ibsvc *ItemBatchService) getTaxGroup() (*entity.TaxGroup, error) {
	invoice := ibsvc.invoice
	invoiceItem := ibsvc.invoiceItem

	groupId, err := utils.GetTaxGroupIdFromTime(
		invoice.MarketDate,
		constants.TaxGroupPrefix.ITEM_BATCH,
	)

	if err != nil {
		return nil, err
	}

	taxGroup := &entity.TaxGroup{
		Source:     entity.ITB,
		ExternalId: groupId,
	}

	itemTotalPrice := invoiceItem.Price * float64(invoiceItem.Qty)
	itemPriceRate := itemTotalPrice / invoice.RawValue
	baseValue := 0.0
	invoiceTaxInstances := invoice.TaxGroup.Taxes
	taxInstances := make([]entity.TaxInstance, 0, len(invoiceTaxInstances))

	for _, invoiceTaxInstance := range invoiceTaxInstances {
		taxValue := ibsvc.getTaxValue(&invoiceTaxInstance, itemPriceRate)
		taxInstance := entity.TaxInstance{
			Tax:        invoiceTaxInstance.Tax,
			MarketDate: invoiceTaxInstance.MarketDate,
			BaseValue:  baseValue,
			TaxValue:   taxValue,
			TaxRate:    invoiceTaxInstance.TaxRate,
		}

		log.Printf(
			"invoiceBatchService.getTaxGroup: found tax %s, tv = %.4f, bv = %.4f, tr = %f",
			taxInstance.Tax.Code,
			taxInstance.TaxValue,
			taxInstance.BaseValue,
			taxInstance.TaxRate,
		)

		taxInstances = append(taxInstances, taxInstance)
	}

	taxGroup.Taxes = taxInstances

	taxGroupService := GetTaxGroupService(ibsvc.tx, taxGroup, ibsvc.taxStore)

	return taxGroupService.CreateTaxGroup()
}

func (ibsvc *ItemBatchService) getItemBatch() (*entity.ItemBatch, error) {
	item := ibsvc.invoiceItem

	taxGroup, err := ibsvc.getTaxGroup()

	if err != nil {
		return nil, err
	}

	companyBatch, err := ibsvc.getCompanyBatch(taxGroup)

	if err != nil {
		return nil, err
	}

	totalTaxes := utils.GetTotalTax(taxGroup)
	rawPrice := item.Price * float64(item.Qty)
	avgPrice := (rawPrice + totalTaxes) / float64(item.Qty)

	itemBatch := &entity.ItemBatch{
		Item:         item,
		TaxGroup:     taxGroup,
		CompanyBatch: companyBatch,
		Qty:          item.Qty,
		AvgPrice:     avgPrice,
		RawPrice:     rawPrice,
		TotalTaxes:   totalTaxes,
	}

	return itemBatch, nil
}

func (ibsvc *ItemBatchService) CreateItemBatch() (*entity.ItemBatch, error) {
	itemBatch, err := ibsvc.getItemBatch()

	if err != nil {
		return nil, err
	}

	itemBatchDAO := db.GetItemBatchDAO(ibsvc.tx, itemBatch)
	return itemBatchDAO.CreateItemBatch()
}
