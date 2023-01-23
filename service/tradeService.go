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

type TradeService struct {
	tx                *sql.Tx
	user              *entity.User
	invoice           *entity.Invoice
	invoiceItem       *entity.InvoiceItem
	tradeBatch        *entity.TradeBatch
	taxStore          *store.TaxStore
	brokerTaxStore    *store.BrokerTaxStore
	companyBatchStore *store.CompanyBatchStore
}

func GetTradeService(
	tx *sql.Tx,
	user *entity.User,
	invoice *entity.Invoice,
	invoiceItem *entity.InvoiceItem,
	tradeBatch *entity.TradeBatch,
	taxStore *store.TaxStore,
	brokerTaxStore *store.BrokerTaxStore,
	companyBatchStore *store.CompanyBatchStore,
) *TradeService {
	return &TradeService{
		tx:                tx,
		user:              user,
		invoice:           invoice,
		invoiceItem:       invoiceItem,
		tradeBatch:        tradeBatch,
		taxStore:          taxStore,
		brokerTaxStore:    brokerTaxStore,
		companyBatchStore: companyBatchStore,
	}
}

func (tsvc *TradeService) getTaxValue(
	invoiceTaxInstance *entity.TaxInstance,
	itemPriceRate float64,
) float64 {
	taxCode := invoiceTaxInstance.Tax.Code
	taxTypes := constants.TaxTypes
	taxRates := constants.TaxRates
	item := tsvc.invoiceItem

	if taxCode == taxTypes.BRKFEE {
		if tsvc.brokerTaxStore.Has(item, taxCode) {
			return 0
		}

		if invoiceTaxInstance.TaxValue == 0 {
			return 0
		}

		tsvc.brokerTaxStore.Put(item, taxCode)

		return taxRates.BRKFEE
	}

	if taxCode == taxTypes.ISSSPFEE {
		if tsvc.brokerTaxStore.Has(item, taxCode) {
			return 0
		}

		if invoiceTaxInstance.TaxValue == 0 {
			return 0
		}

		tsvc.brokerTaxStore.Put(item, taxCode)

		// t = (b / (1 - i)) - c
		return (taxRates.BRKFEE / (1 - taxRates.ISSSPFEE)) - taxRates.BRKFEE
	}

	if taxCode == taxTypes.IRRFFEE {
		irrfFeeBaseValue := item.Price * float64(item.Qty)
		return irrfFeeBaseValue * taxRates.IRRFFEE
	}

	return invoiceTaxInstance.TaxValue * itemPriceRate
}

func (tsvc *TradeService) getTaxGroup() (*entity.TaxGroup, error) {
	invoice := tsvc.invoice
	invoiceItem := tsvc.invoiceItem

	groupId, err := utils.GetTaxGroupIdFromTime(
		invoice.MarketDate,
		constants.TaxGroupPrefix.TRADE,
	)

	if err != nil {
		return nil, err
	}

	taxGroup := &entity.TaxGroup{
		Source:     entity.TRD,
		ExternalId: groupId,
	}

	itemTotalPrice := invoiceItem.Price * float64(invoiceItem.Qty)
	itemPriceRate := itemTotalPrice / invoice.RawValue
	baseValue := 0.0
	invoiceTaxInstances := invoice.TaxGroup.Taxes
	taxInstances := make([]entity.TaxInstance, 0, len(invoiceTaxInstances))

	for _, invoiceTaxInstance := range invoiceTaxInstances {
		taxValue := tsvc.getTaxValue(&invoiceTaxInstance, itemPriceRate)
		taxInstance := entity.TaxInstance{
			Tax:        invoiceTaxInstance.Tax,
			MarketDate: invoiceTaxInstance.MarketDate,
			BaseValue:  baseValue,
			TaxValue:   taxValue,
			TaxRate:    invoiceTaxInstance.TaxRate,
		}

		log.Printf(
			"TradeService.getTaxGroup: found tax %s, tv = %.4f, bv = %.4f, tr = %f",
			taxInstance.Tax.Code,
			taxInstance.TaxValue,
			taxInstance.BaseValue,
			taxInstance.TaxRate,
		)

		taxInstances = append(taxInstances, taxInstance)
	}

	taxGroup.Taxes = taxInstances

	taxGroupService := GetTaxGroupService(tsvc.tx, taxGroup, tsvc.taxStore)

	return taxGroupService.CreateTaxGroup()
}

func (tsvc *TradeService) getCompanyBatch() (*entity.CompanyBatch, error) {
	user := tsvc.user
	invoiceItem := tsvc.invoiceItem
	store := tsvc.companyBatchStore

	companyBatch := &entity.CompanyBatch{
		User:    user,
		Company: invoiceItem.Company,
	}

	if store.Has(companyBatch) {
		companyBatchRec := store.Get(companyBatch)

		log.Printf(
			"TradeService.getCompanyBatch: found on cache [%d, %s, qty = %d, avg = %.4f]",
			companyBatchRec.Id,
			companyBatchRec.Company.Code,
			companyBatchRec.Qty,
			companyBatchRec.AvgPrice,
		)

		return companyBatchRec, nil
	}

	companyBatchDAO := db.GetCompanyBatchDAO(
		tsvc.tx,
		companyBatch,
	)

	companyBatchRec, err := companyBatchDAO.GetCompanyBatch()

	if err != nil {
		return nil, err
	}

	store.Put(companyBatchRec)

	return companyBatchRec, nil
}

func (tsvc *TradeService) adjustCompanyBatch() (*entity.CompanyBatch, error) {
	invoiceItem := tsvc.invoiceItem
	store := tsvc.companyBatchStore

	companyBatch, err := tsvc.getCompanyBatch()

	if err != nil {
		return nil, err
	}

	totalSold := companyBatch.AvgPrice * float64(invoiceItem.Qty)
	companyBatch.Qty = companyBatch.Qty - invoiceItem.Qty
	companyBatch.TotalPrice = companyBatch.TotalPrice - totalSold

	companyBatchDAO := db.GetCompanyBatchDAO(
		tsvc.tx,
		companyBatch,
	)

	companyBatchRec, err := companyBatchDAO.UpdateCompanyBatch()

	if err != nil {
		return nil, err
	}

	store.Put(companyBatchRec)

	return companyBatchRec, nil
}

func (tsvc *TradeService) ProcessTrade() (*entity.Trade, error) {
	invoiceItem := tsvc.invoiceItem

	taxGroup, err := tsvc.getTaxGroup()

	if err != nil {
		return nil, err
	}

	companyBatch, err := tsvc.adjustCompanyBatch()

	if err != nil {
		return nil, err
	}

	totalTax := taxGroup.GetTotalTax()
	aqPrice := companyBatch.AvgPrice * float64(invoiceItem.Qty)
	slPrice := invoiceItem.Price * float64(invoiceItem.Qty)
	rawResults := slPrice - aqPrice
	avgPrice := (slPrice + totalTax) / float64(invoiceItem.Qty)

	trade := &entity.Trade{
		TaxGroup:     taxGroup,
		Item:         tsvc.invoiceItem,
		TradeBatch:   tsvc.tradeBatch,
		CompanyBatch: companyBatch,
		MarketDate:   invoiceItem.MarketDate,
		Qty:          invoiceItem.Qty,
		AvgPrice:     avgPrice,
		RawResults:   rawResults,
		TotalTax:     totalTax,
	}

	tradeDAO := db.GetTradeDAO(tsvc.tx, trade)

	return tradeDAO.CreateTrade()
}
