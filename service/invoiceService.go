package service

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/constants"
	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/input"
	"github.com/jarismar/b3c-invoice-reader-lambda/store"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

type InvoiceService struct {
	tx                *sql.Tx
	invoiceInput      *input.Invoice
	taxStore          *store.TaxStore
	companyStore      *store.CompanyStore
	companyBatchStore *store.CompanyBatchStore
	brokerTaxStore    *store.BrokerTaxStore
}

func GetInvoiceService(
	tx *sql.Tx,
	invoiceInput *input.Invoice,
	taxStore *store.TaxStore,
	companyStore *store.CompanyStore,
	companyBatchStore *store.CompanyBatchStore,
	brokerTaxStore *store.BrokerTaxStore,
) *InvoiceService {
	return &InvoiceService{
		tx:                tx,
		invoiceInput:      invoiceInput,
		taxStore:          taxStore,
		companyStore:      companyStore,
		companyBatchStore: companyBatchStore,
		brokerTaxStore:    brokerTaxStore,
	}
}

func (isvc *InvoiceService) getTaxBaseValue(inputTax *input.Tax) (float64, error) {
	invoice := isvc.invoiceInput
	taxTypes := constants.TaxTypes

	switch inputTax.Code {
	case taxTypes.SETFEE:
		return invoice.RawValue, nil
	case taxTypes.EMLFEE:
		return invoice.RawValue, nil
	case taxTypes.BRKFEE:
		return inputTax.Value, nil
	case taxTypes.ISSSPFEE:
		{
			brkFree, err := utils.GetInputTaxByCode(invoice.Taxes, taxTypes.BRKFEE)
			if err != nil {
				return 0, err
			}
			return brkFree.Value, nil
		}
	case taxTypes.IRRFFEE:
		return invoice.TotalSold, nil
	}

	error := fmt.Errorf("invoiceTaxService::GetInvoiceTaxBaseValue: Unknown tax code %s", inputTax.Code)
	return 0.0, error
}

func (isvc *InvoiceService) getTaxValue(baseValue float64, taxInput *input.Tax) (float64, error) {
	invoice := isvc.invoiceInput
	taxTypes := constants.TaxTypes
	taxRates := constants.TaxRates

	switch taxInput.Code {
	case taxTypes.SETFEE:
		return (baseValue * taxRates.SETFEE), nil
	case taxTypes.EMLFEE:
		return (taxInput.Value), nil
	case taxTypes.BRKFEE:
		return baseValue, nil
	case taxTypes.ISSSPFEE:
		return ((baseValue / 0.95) - baseValue), nil
	case taxTypes.IRRFFEE:
		return (invoice.TotalSold * taxRates.IRRFFEE), nil
	}

	error := fmt.Errorf("invoiceTaxService::GetInvoiceTaxValue: Unknown tax code %s", taxInput.Code)
	return 0, error
}

func (isvc *InvoiceService) getTaxRate(taxInput *input.Tax) float64 {
	if taxInput.Rate > 0 {
		return taxInput.Rate
	}

	taxTypes := constants.TaxTypes
	taxRates := constants.TaxRates

	switch taxInput.Code {
	case taxTypes.SETFEE:
		return taxRates.SETFEE
	case taxTypes.EMLFEE:
		return taxRates.EMLFEE
	case taxTypes.ISSSPFEE:
		return taxRates.ISSSPFEE
	case taxTypes.IRRFFEE:
		return taxRates.IRRFFEE
	default:
		return 0.0
	}
}

func (isvc *InvoiceService) getTaxGroup() (*entity.TaxGroup, error) {
	invoiceInput := isvc.invoiceInput

	groupId, err := utils.GetTaxGroupId(
		invoiceInput.MarketDate,
		constants.TaxGroupPrefix.INVOICE,
	)

	if err != nil {
		return nil, err
	}

	taxGroup := &entity.TaxGroup{
		Source:     entity.BIV,
		ExternalId: groupId,
	}

	inputTaxes := invoiceInput.Taxes
	taxInstances := make([]entity.TaxInstance, 0, len(inputTaxes))

	for _, taxInput := range inputTaxes {
		baseValue, err := isvc.getTaxBaseValue(&taxInput)

		if err != nil {
			return nil, err
		}

		taxValue, err := isvc.getTaxValue(baseValue, &taxInput)

		if err != nil {
			return nil, err
		}

		marketDate, err := utils.GetDateObject(invoiceInput.MarketDate)

		if err != nil {
			return nil, err
		}

		taxRate := isvc.getTaxRate(&taxInput)

		taxInstance := entity.TaxInstance{
			Tax: &entity.Tax{
				Code:   taxInput.Code,
				Source: taxInput.Source,
				Rate:   taxRate,
			},
			MarketDate: marketDate,
			TaxValue:   taxValue,
			BaseValue:  baseValue,
			TaxRate:    taxRate,
		}

		log.Printf(
			"invoiceService.getTaxGroup: found tax %s, tv = %.4f, bv = %.4f, tr = %f",
			taxInstance.Tax.Code,
			taxInstance.TaxValue,
			taxInstance.BaseValue,
			taxInstance.TaxRate,
		)

		taxInstances = append(taxInstances, taxInstance)
	}

	taxGroup.Taxes = taxInstances

	return taxGroup, nil
}

func (isvc *InvoiceService) upsertUser() (*entity.User, error) {
	invoiceInput := isvc.invoiceInput

	user := &entity.User{
		ExternalUUID: invoiceInput.Client.Id,
		UserName:     invoiceInput.Client.Name,
	}

	userService := GetUserService(isvc.tx, user)
	return userService.UpsertUser()
}

func (isvc *InvoiceService) getInvoiceItem(invoice *entity.Invoice, itemInput *input.Item) (*entity.InvoiceItem, error) {
	company := &entity.Company{
		Code: itemInput.Company.Code,
		Name: itemInput.Company.Name,
	}

	companyService := GetCompanyService(
		isvc.tx,
		company,
		isvc.companyStore,
	)

	companyRec, err := companyService.UpsertCompany()

	if err != nil {
		return nil, err
	}

	invoiceItem := &entity.InvoiceItem{
		Company:    companyRec,
		InvoiceID:  invoice.Id,
		MarketDate: invoice.MarketDate,
		Qty:        itemInput.Qty,
		Price:      itemInput.Price,
		Debit:      itemInput.Debit,
		Order:      itemInput.Order,
	}

	return invoiceItem, nil
}

func (isvc *InvoiceService) ProcessInvoice() (*entity.Invoice, error) {
	invoiceInput := isvc.invoiceInput

	invoice := &entity.Invoice{
		Number:        invoiceInput.InvoiceNum,
		FileName:      invoiceInput.FileName,
		TotalSold:     invoiceInput.TotalSold,
		TotalAcquired: invoiceInput.TotalAcquired,
		RawValue:      invoiceInput.RawValue,
		NetValue:      invoiceInput.NetValue,
	}

	invoiceDAO := db.GetInvoiceDAO(isvc.tx, invoice)
	isNew, err := invoiceDAO.IsNewInvoice()

	if err != nil {
		return nil, err
	}

	if !isNew {
		err = fmt.Errorf(
			"invoiceService.ProcessInvoice: error: Invoice %s already exists on DB",
			invoice.FileName,
		)
		return nil, err
	}

	// handle user
	userRec, err := isvc.upsertUser()

	invoice.User = userRec

	if err != nil {
		return nil, err
	}

	// handle invoice taxes
	taxGroup, err := isvc.getTaxGroup()

	if err != nil {
		return nil, err
	}

	taxGroupService := GetTaxGroupService(isvc.tx, taxGroup, isvc.taxStore)
	taxGroupRec, err := taxGroupService.CreateTaxGroup()

	if err != nil {
		return nil, err
	}

	// handle invoice
	marketDate, err := utils.GetDateObject(invoiceInput.MarketDate)

	if err != nil {
		return nil, err
	}

	billingDate, err := utils.GetDateObject(invoiceInput.BillingDate)

	if err != nil {
		return nil, err
	}

	invoice.User = userRec
	invoice.TaxGroup = taxGroupRec
	invoice.MarketDate = marketDate
	invoice.BillingDate = billingDate

	invoiceRec, err := invoiceDAO.CreateInvoice()

	if err != nil {
		return nil, err
	}

	//handle items
	var tradeBatch *entity.TradeBatch = nil

	items := make([]entity.InvoiceItem, 0, len(invoiceInput.Items))
	for _, item := range invoiceInput.Items {
		invoiceItem, err := isvc.getInvoiceItem(invoiceRec, &item)

		if err != nil {
			return nil, err
		}

		invoiceItemService := GetInvoiceItemService(isvc.tx, invoiceItem)
		itemRec, err := invoiceItemService.CreateItem()

		if err != nil {
			return nil, err
		}

		if item.Debit {
			// item batch
			itemBatchService := GetItemBatchService(
				isvc.tx,
				userRec,
				invoiceRec,
				itemRec,
				isvc.taxStore,
				isvc.brokerTaxStore,
				isvc.companyBatchStore,
			)

			itemBatchRec, err := itemBatchService.CreateItemBatch()

			if err != nil {
				return nil, err
			}

			itemRec.ItemBatch = itemBatchRec
		} else {
			// trade batch
			if tradeBatch == nil {
				tradeBatchService := GetTradeBatchService(
					isvc.tx,
					userRec,
					tradeBatch,
					isvc.taxStore,
				)

				tradeBatch, err = tradeBatchService.FindTradeBatch(invoiceRec.MarketDate)

				if err != nil {
					return nil, err
				}
			}

			tradeService := GetTradeService(
				isvc.tx,
				userRec,
				invoiceRec,
				itemRec,
				tradeBatch,
				isvc.taxStore,
				isvc.brokerTaxStore,
				isvc.companyBatchStore,
			)

			tradeRec, err := tradeService.ProcessTrade()

			if err != nil {
				return nil, err
			}

			tradeBatchService := GetTradeBatchService(
				isvc.tx,
				userRec,
				tradeBatch,
				isvc.taxStore,
			)

			tradeBatch = tradeBatchService.ProcessTrade(tradeRec)

			tradeRec.TradeBatch = tradeBatch
			itemRec.Trade = tradeRec
		}

		items = append(items, *itemRec)
	}

	if tradeBatch != nil {
		tradeBatchService := GetTradeBatchService(
			isvc.tx,
			userRec,
			tradeBatch,
			isvc.taxStore,
		)

		tradeBatch, err = tradeBatchService.SaveTradeBatch()

		if err != nil {
			return nil, err
		}

		for _, item := range items {
			if !item.Debit {
				item.Trade.TradeBatch = tradeBatch
			}
		}
	}

	invoiceRec.Items = items

	return invoiceRec, nil
}
