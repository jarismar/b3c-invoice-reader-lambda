package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type InvoiceItemDAO struct {
	tx   *sql.Tx
	item *entity.InvoiceItem
}

func GetInvoiceItemDAO(tx *sql.Tx, item *entity.InvoiceItem) *InvoiceItemDAO {
	return &InvoiceItemDAO{
		tx:   tx,
		item: item,
	}
}

func (dao *InvoiceItemDAO) CreateItem() (*entity.InvoiceItem, error) {
	insertStmt := `INSERT INTO broker_invoice_item (
		cmp_id,
		biv_id,
		biv_market_date,
		bii_order,
		bii_qty,
		bii_price,
		bii_debit
	) VALUES (?,?,?,?,?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	item := dao.item

	res, err := stmt.Exec(
		item.Company.Id,
		item.InvoiceID,
		item.MarketDate,
		item.Order,
		item.Qty,
		item.Price,
		item.Debit,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	itemRec := &entity.InvoiceItem{
		Id:         lastId,
		Company:    item.Company,
		MarketDate: item.MarketDate,
		Qty:        item.Qty,
		Price:      item.Price,
		Debit:      item.Debit,
		Order:      item.Order,
	}

	log.Printf(
		"invoiceItemDAO.CreateItem: created new invoice item [%d, %d, %s, %d]",
		lastId,
		item.Order,
		item.Company.Code,
		item.Qty,
	)

	return itemRec, nil
}
