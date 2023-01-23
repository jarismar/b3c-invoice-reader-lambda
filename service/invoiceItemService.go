package service

import (
	"database/sql"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-service-entities/entity"
)

type InvoiceItemService struct {
	tx   *sql.Tx
	item *entity.InvoiceItem
}

func GetInvoiceItemService(tx *sql.Tx, item *entity.InvoiceItem) *InvoiceItemService {
	return &InvoiceItemService{
		tx:   tx,
		item: item,
	}
}

func (iisvc *InvoiceItemService) CreateItem() (*entity.InvoiceItem, error) {
	invoiceItemDAO := db.GetInvoiceItemDAO(iisvc.tx, iisvc.item)
	return invoiceItemDAO.CreateItem()
}
