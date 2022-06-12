package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

const queryInvoiceItemByInvoiceAndCompany = `SELECT
  bii_id,
  bii_qty,
  bii_qty_balance,
  bii_price,
  bii_debit
FROM broker_invoice_item
WHERE biv_id = ? AND cmp_id = ?`

const insertInvoiceItemStmt = `INSERT INTO broker_invoice_item (
	biv_id,
	cmp_id,
	bii_qty,
	bii_qty_balance,
	bii_price,
	bii_debit
) VALUES (?,?,?,?,?,?)`

type InvoiceItemDAO struct {
	conn    *sql.Tx
	invoice *entity.Invoice
	company *entity.Company
}

func GetInvoiceItemDAO(conn *sql.Tx, invoice *entity.Invoice, company *entity.Company) *InvoiceItemDAO {
	return &InvoiceItemDAO{
		conn:    conn,
		invoice: invoice,
		company: company,
	}
}

func (dao *InvoiceItemDAO) FindByInvoiceAndCompany() (*entity.InvoiceItem, error) {
	stmt, err := dao.conn.Prepare(queryInvoiceItemByInvoiceAndCompany)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var invoiceItem entity.InvoiceItem
	var debit []uint8

	qryErr := stmt.QueryRow(dao.invoice.Id, dao.company.Id).Scan(
		&invoiceItem.Id,
		&invoiceItem.Qty,
		&invoiceItem.Balance,
		&invoiceItem.Price,
		&debit,
	)

	if qryErr == sql.ErrNoRows {
		return nil, nil
	} else if qryErr != nil {
		return nil, qryErr
	}

	invoiceItem.Company = dao.company
	invoiceItem.Debit = (debit[0] == 1)

	return &invoiceItem, nil
}

func (dao *InvoiceItemDAO) CreateInvoiceItem(invoiceItemInput *inputData.Item) (*entity.InvoiceItem, error) {
	stmt, err := dao.conn.Prepare(insertInvoiceItemStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		dao.invoice.Id,
		dao.company.Id,
		invoiceItemInput.Qty,
		invoiceItemInput.Qty,
		invoiceItemInput.Price,
		invoiceItemInput.Debit,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	invoiceItem := entity.InvoiceItem{
		Id:      lastId,
		Company: dao.company,
		Qty:     invoiceItemInput.Qty,
		Balance: invoiceItemInput.Qty,
		Price:   invoiceItemInput.Price,
		Debit:   invoiceItemInput.Debit,
	}

	log.Printf("created invoice item [%d, %s, %d]\n", lastId, dao.company.Code, invoiceItem.Qty)

	return &invoiceItem, nil
}
