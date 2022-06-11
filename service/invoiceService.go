package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type InvoiceService struct {
	invoiceDAO *db.InvoiceDAO
	user       *entity.User
}

func GetInvoiceService(conn *sql.Tx, user *entity.User) *InvoiceService {
	return &InvoiceService{
		invoiceDAO: db.GetInvoiceDAO(conn, user),
		user:       user,
	}
}

func (invoiceService *InvoiceService) insert(invoiceInput *inputData.Invoice) (*entity.Invoice, error) {
	log.Printf("creating invoice [%s]\n", invoiceInput.FileName)
	return invoiceService.invoiceDAO.CreateInvoice(invoiceInput)
}

func (invoiceService *InvoiceService) UpsertInvoice(invoiceInput *inputData.Invoice) (*entity.Invoice, error) {
	invoice, err := invoiceService.invoiceDAO.FindByFileName(invoiceInput.FileName)
	if err != nil {
		return nil, err
	}

	if invoice == nil {
		return invoiceService.invoiceDAO.CreateInvoice(invoiceInput)
	}

	// TODO implement support for invoice update

	log.Printf("nothing to be done for invoice [%s]\n", invoice.FileName)
	return invoice, nil
}
