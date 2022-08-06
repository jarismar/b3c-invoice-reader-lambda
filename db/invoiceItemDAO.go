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
WHERE biv_id = ? AND cmp_id = ? AND bii_order = ?`

const qryInvoiceItemById = `SELECT
  cmp_id,
  biv_id,
  bii_qty,
  bii_qty_balance,
  bii_price,
  bii_debit,
  bii_order
FROM broker_invoice_item
WHERE bii_id = ?`

const searchInvoiceItemForSale = `SELECT
  bii.bii_id,
  bii.bii_qty,
  bii.bii_qty_balance,
  bii.bii_price,
  bii.bii_order
FROM broker_invoice_item bii
INNER JOIN broker_invoice biv ON bii.biv_id = biv.biv_id
WHERE cmp_id = ?
  AND bii_qty_balance > 0
  AND bii_debit = 1
ORDER BY biv.biv_market_date DESC, bii.bii_order DESC`

const insertInvoiceItemStmt = `INSERT INTO broker_invoice_item (
	biv_id,
	cmp_id,
	bii_qty,
	bii_qty_balance,
	bii_price,
	bii_debit,
	bii_order
) VALUES (?,?,?,?,?,?,?)`

const updateItemBalanceStmt = `UPDATE broker_invoice_item SET
  bii_qty_balance = ?
WHERE bii_id = ?`

type InvoiceItemDAO struct {
	conn    *sql.Tx
	invoice *entity.Invoice
	company *entity.Company
	order   int64
}

func GetInvoiceItemDAO(
	conn *sql.Tx,
	invoice *entity.Invoice,
	company *entity.Company,
	order int64,
) *InvoiceItemDAO {
	return &InvoiceItemDAO{
		conn:    conn,
		invoice: invoice,
		company: company,
		order:   order,
	}
}

func (dao *InvoiceItemDAO) FindById(id int64) (*entity.InvoiceItem, error) {
	stmt, err := dao.conn.Prepare(qryInvoiceItemById)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var item entity.InvoiceItem
	var companyId int64
	var invoiceId int64
	var debit []uint8

	err = stmt.QueryRow(id).Scan(
		&companyId,
		&invoiceId,
		&item.Qty,
		&item.Balance,
		&item.Price,
		&debit,
		&item.Order,
	)

	// TODO load company

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	item.Id = id
	item.Debit = (debit[0] == 1)

	return &item, nil
}

func (dao *InvoiceItemDAO) SearchInvoiceItemForSale(itemOut *entity.InvoiceItem) ([]entity.InvoiceItem, error) {
	stmt, err := dao.conn.Prepare(searchInvoiceItemForSale)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(itemOut.Company.Id)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	itemInvoiceList := make([]entity.InvoiceItem, 0, 10)
	for rows.Next() {
		var item entity.InvoiceItem
		err = rows.Scan(
			&item.Id,
			&item.Qty,
			&item.Balance,
			&item.Price,
			&item.Order,
		)

		if err != nil {
			return nil, err
		}

		item.Debit = false
		item.Company = itemOut.Company

		itemInvoiceList = append(itemInvoiceList, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return itemInvoiceList, nil
}

func (dao *InvoiceItemDAO) FindByInvoiceAndCompanyAndOrder() (*entity.InvoiceItem, error) {
	stmt, err := dao.conn.Prepare(queryInvoiceItemByInvoiceAndCompany)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var invoiceItem entity.InvoiceItem
	var debit []uint8

	err = stmt.QueryRow(
		dao.invoice.Id,
		dao.company.Id,
		dao.order,
	).Scan(
		&invoiceItem.Id,
		&invoiceItem.Qty,
		&invoiceItem.Balance,
		&invoiceItem.Price,
		&debit,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	invoiceItem.Order = dao.order
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
		invoiceItemInput.Order,
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
		Order:   invoiceItemInput.Order,
	}

	log.Printf("created invoice item [%d, %d, %s, %d]\n", lastId, invoiceItem.Order, dao.company.Code, invoiceItem.Qty)

	return &invoiceItem, nil
}

func (dao *InvoiceItemDAO) UpdateBalance(invoiceItem *entity.InvoiceItem, balance int64) error {
	stmt, err := dao.conn.Prepare(updateItemBalanceStmt)
	if err != nil {
		return nil
	}

	defer stmt.Close()

	_, err = stmt.Exec(balance, invoiceItem.Id)

	if err == nil {
		log.Printf("updated item balance [%d, %s, %d -> %d]\n", invoiceItem.Id, invoiceItem.Company.Code, invoiceItem.Balance, balance)
	}

	return err
}
