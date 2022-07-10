package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

const qrySaleByItemOut = `SELECT
  bii_id_in
FROM broker_invoice_item_sale
WHERE bii_id_out = ?`

const insertSaleStmt = `INSERT INTO broker_invoice_item_sale (
  bii_id_in,
  bii_id_out
) VALUES (?,?)`

type SalesDAO struct {
	conn    *sql.Tx
	invoice *entity.Invoice
	itemOut *entity.InvoiceItem
}

func GetSalesDAO(conn *sql.Tx, invoice *entity.Invoice, itemOut *entity.InvoiceItem) *SalesDAO {
	return &SalesDAO{
		conn:    conn,
		invoice: invoice,
		itemOut: itemOut,
	}
}

func (dao *SalesDAO) FindByItemOut() (*entity.Trade, error) {
	stmt, err := dao.conn.Prepare(qrySaleByItemOut)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(dao.itemOut.Id)

	if err != nil {
		return nil, err
	}

	itemIdsIn := make([]int64, 0, 10)
	itemsIn := make([]entity.InvoiceItem, 0, 10)

	var itemId int64

	for rows.Next() {
		err = rows.Scan(
			&itemId,
		)

		if err != nil {
			return nil, err
		}

		itemIdsIn = append(itemIdsIn, itemId)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(itemIdsIn) == 0 {
		return nil, err
	}

	invoiceItemsDAO := GetInvoiceItemDAO(dao.conn, nil, nil)

	for _, itemId := range itemIdsIn {
		itemIn, err := invoiceItemsDAO.FindById(itemId)
		if err != nil {
			return nil, err
		}

		itemsIn = append(itemsIn, *itemIn)
	}

	var trade entity.Trade
	trade.Invoice = dao.invoice
	trade.InvoiceItemOut = dao.itemOut
	trade.InvoiceItemsIn = itemsIn

	return &trade, nil
}

func (dao *SalesDAO) CreateSale(invoiceItemIn *entity.InvoiceItem) error {
	stmt, err := dao.conn.Prepare(insertSaleStmt)
	if err != nil {
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		invoiceItemIn.Id,
		dao.itemOut.Id,
	)

	if err != nil {
		return err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return err
	}

	log.Printf("created sale [%d, %s]\n", lastId, dao.itemOut.Company.Code)

	return nil
}
