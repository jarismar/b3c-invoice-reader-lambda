package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type InvoiceTaxService struct {
	invoiceTaxDAO db.InvoiceTaxDAO
	invoice       *entity.Invoice
	tax           *entity.Tax
}

func GetInvoiceTaxService(tx *sql.Tx, invoice *entity.Invoice, tax *entity.Tax) *InvoiceTaxService {
	return &InvoiceTaxService{
		invoiceTaxDAO: *db.GetInvoiceTaxDAO(tx, invoice, tax),
		invoice:       invoice,
		tax:           tax,
	}
}

func (invoiceTaxService *InvoiceTaxService) insert(invoiceTaxInput *inputData.Tax) (*entity.InvoiceTax, error) {
	log.Printf("creating invoiceTax [%s, %f]\n", invoiceTaxInput.Code, invoiceTaxInput.Value)
	return invoiceTaxService.invoiceTaxDAO.CreateInvoiceTax(invoiceTaxInput)
}

func (invoiceTaxService *InvoiceTaxService) UpsertInvoiceTax(invoiceTaxInput *inputData.Tax) (*entity.InvoiceTax, error) {
	invoiceTax, err := invoiceTaxService.invoiceTaxDAO.FindByInvoiceAndTax()
	if err != nil {
		return nil, err
	}

	if invoiceTax == nil {
		return invoiceTaxService.insert(invoiceTaxInput)
	}

	// TODO implement support for invoice tax update

	log.Printf("nothing to be done for invoice tax [%s, %.2f]\n", invoiceTax.Tax.Code, invoiceTax.Value)

	return invoiceTax, nil
}
