package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-service-entities/entity"
)

type SaleService struct {
	tx      *sql.Tx
	dao     *db.SalesDAO
	invoice *entity.Invoice
	itemOut *entity.InvoiceItem
}

func GetSaleService(tx *sql.Tx, invoice *entity.Invoice, itemOut *entity.InvoiceItem) *SaleService {
	return &SaleService{
		tx:      tx,
		dao:     db.GetSalesDAO(tx, invoice, itemOut),
		invoice: invoice,
		itemOut: itemOut,
	}
}

func (saleService *SaleService) insert() (*entity.Trade, error) {
	itemOut := saleService.itemOut

	log.Printf("creating trade [%s, %d]\n", itemOut.Company.Name, itemOut.Qty)

	invoiceItemDAO := db.GetInvoiceItemDAO(saleService.tx, saleService.invoice, nil, 0)
	itemInList, err := invoiceItemDAO.SearchInvoiceItemForSale(itemOut)
	if err != nil {
		return nil, err
	}

	var itemBalance, qtySold int64

	pendingQty := itemOut.Qty
	for _, itemIn := range itemInList {
		if itemIn.Balance >= pendingQty {
			itemBalance = itemIn.Balance - pendingQty
			qtySold = pendingQty
		} else {
			itemBalance = 0
			qtySold = itemIn.Balance
		}

		pendingQty = (pendingQty - qtySold)
		invoiceItemDAO.UpdateBalance(&itemIn, itemBalance)
		saleService.dao.CreateSale(&itemIn)

		if pendingQty <= 0 {
			break
		}
	}

	log.Printf("created new trade [%s, %d]\n", itemOut.Company.Name, itemOut.Qty)

	return &entity.Trade{
		Invoice:        saleService.invoice,
		InvoiceItemOut: saleService.itemOut,
		InvoiceItemsIn: itemInList,
	}, nil
}

func (saleService *SaleService) UpsertSale() (*entity.Trade, error) {
	trade, err := saleService.dao.FindByItemOut()
	if err != nil {
		return nil, err
	}

	if trade == nil {
		return saleService.insert()
	}

	itemOut := saleService.itemOut
	log.Printf("nothing to be done for trade [%s, %d]\n", itemOut.Company.Name, itemOut.Qty)

	return trade, nil
}
