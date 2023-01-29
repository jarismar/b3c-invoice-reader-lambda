package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type InvoiceDAO struct {
	tx      *sql.Tx
	invoice *entity.Invoice
}

func GetInvoiceDAO(tx *sql.Tx, invoice *entity.Invoice) *InvoiceDAO {
	return &InvoiceDAO{
		tx:      tx,
		invoice: invoice,
	}
}

func (dao *InvoiceDAO) IsNewInvoice() (bool, error) {
	filename := dao.invoice.FileName
	query := `SELECT biv_id FROM broker_invoice WHERE biv_filename = ?`
	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return false, err
	}

	defer stmt.Close()

	var invoiceID int64

	err = stmt.QueryRow(filename).Scan(
		&invoiceID,
	)

	if err == sql.ErrNoRows {
		return true, nil
	} else if err != nil {
		return false, err
	}

	log.Printf("invoiceDAO.IsNew: invoice already exists [%d, %s]", invoiceID, filename)

	return false, nil
}

func (dao *InvoiceDAO) CreateInvoice() (*entity.Invoice, error) {
	insertStmt := `INSERT INTO broker_invoice (
		usr_id,
		tgr_id,
		biv_filename,
		biv_number,
		biv_market_date,
		biv_billing_date,
		biv_raw_value,
		biv_net_value,
		biv_total_sold,
		biv_total_acquired
	  ) VALUES (?,?,?,?,?,?,?,?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	invoice := dao.invoice
	res, err := stmt.Exec(
		invoice.User.Id,
		invoice.TaxGroup.Id,
		invoice.FileName,
		invoice.Number,
		invoice.MarketDate,
		invoice.BillingDate,
		invoice.RawValue,
		invoice.NetValue,
		invoice.TotalSold,
		invoice.TotalAcquired,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	invoiceRec := &entity.Invoice{
		Id:            lastId,
		Number:        invoice.Number,
		User:          invoice.User,
		TaxGroup:      invoice.TaxGroup,
		FileName:      invoice.FileName,
		MarketDate:    invoice.MarketDate,
		BillingDate:   invoice.BillingDate,
		RawValue:      invoice.RawValue,
		NetValue:      invoice.NetValue,
		TotalSold:     invoice.TotalSold,
		TotalAcquired: invoice.TotalAcquired,
		Items:         invoice.Items,
	}

	log.Printf(
		"invoiceDAO.CreateInvoice: created new invoice [%d, %s]",
		invoiceRec.Id,
		invoiceRec.FileName,
	)

	return invoiceRec, nil
}
