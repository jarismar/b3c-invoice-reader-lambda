package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type InvoiceItemService struct {
	invoiceItemDAO *db.InvoiceItemDAO
	invoice        *entity.Invoice
	company        *entity.Company
	order          int64
}

func GetInvoiceItemService(
	tx *sql.Tx,
	invoice *entity.Invoice,
	company *entity.Company,
	order int64,
) *InvoiceItemService {
	return &InvoiceItemService{
		invoiceItemDAO: db.GetInvoiceItemDAO(tx, invoice, company, order),
		invoice:        invoice,
		company:        company,
		order:          order,
	}
}

func (invoiceItemService *InvoiceItemService) insert(invoiceItemInput *inputData.Item) (*entity.InvoiceItem, error) {
	company := invoiceItemService.company
	log.Printf("creating invoiceItem [%s, %d]\n", company.Code, invoiceItemInput.Qty)
	return invoiceItemService.invoiceItemDAO.CreateInvoiceItem(invoiceItemInput)
}

func (invoiceItemService *InvoiceItemService) UpsertInvoiceItem(invoiceItemInput *inputData.Item) (*entity.InvoiceItem, error) {
	invoiceItem, err := invoiceItemService.invoiceItemDAO.FindByInvoiceAndCompanyAndOrder()
	if err != nil {
		return nil, err
	}

	if invoiceItem == nil {
		return invoiceItemService.insert(invoiceItemInput)
	}

	// TODO implement support for invoiceItem update

	log.Printf("nothing to be done for invoiceItem [%s, %d, %d]\n", invoiceItem.Company.Code, invoiceItem.Qty, invoiceItem.Order)
	return invoiceItem, nil
}
