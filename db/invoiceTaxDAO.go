package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

const qryInvoiceTaxByInvoiceAndTax = `SELECT
  bit_id,
  bit_value
FROM broker_invoice_tax
WHERE biv_id = ? AND tax_id = ?`

const insertInvoiceTaxStmt = `INSERT INTO broker_invoice_tax (
  biv_id,
  tax_id,
  bit_value
) VALUES (?,?,?)`

type InvoiceTaxDAO struct {
	conn    *sql.Tx
	invoice *entity.Invoice
	tax     *entity.Tax
}

func GetInvoiceTaxDAO(conn *sql.Tx, invoice *entity.Invoice, tax *entity.Tax) *InvoiceTaxDAO {
	return &InvoiceTaxDAO{
		conn:    conn,
		invoice: invoice,
		tax:     tax,
	}
}

func (dao *InvoiceTaxDAO) FindByInvoiceAndTax() (*entity.InvoiceTax, error) {
	stmt, err := dao.conn.Prepare(qryInvoiceTaxByInvoiceAndTax)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var invoiceTax entity.InvoiceTax

	err = stmt.QueryRow(dao.invoice.Id, dao.tax.Id).Scan(
		&invoiceTax.Id,
		&invoiceTax.Value,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	invoiceTax.Tax = dao.tax

	return &invoiceTax, nil
}

func (dao *InvoiceTaxDAO) CreateInvoiceTax(invoiceTaxInput *inputData.Tax) (*entity.InvoiceTax, error) {
	stmt, err := dao.conn.Prepare(insertInvoiceTaxStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		dao.invoice.Id,
		dao.tax.Id,
		invoiceTaxInput.Value,
	)
	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	invoiceTax := entity.InvoiceTax{
		Id:    lastId,
		Tax:   dao.tax,
		Value: invoiceTaxInput.Value,
	}

	log.Printf("created invoiceTax [%d, %s, %f]\n", invoiceTax.Id, dao.tax.Code, invoiceTax.Value)

	return &invoiceTax, nil
}
