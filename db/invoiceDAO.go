package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

const qryInvoiceByFilename = `SELECT
  biv_id,
  usr_id,
  biv_filename,
  biv_number,
  biv_market_date,
  biv_billing_date,
  biv_raw_value,
  biv_net_value
FROM broker_invoice
WHERE biv_filename = ?`

const insertInvoiceStmt = `INSERT INTO broker_invoice (
  usr_id,
  biv_filename,
  biv_number,
  biv_market_date,
  biv_billing_date,
  biv_raw_value,
  biv_net_value
) VALUES (?,?,?,?,?,?,?)`

type InvoiceDAO struct {
	conn *sql.Tx
	user *entity.User
}

func GetInvoiceDAO(conn *sql.Tx, user *entity.User) *InvoiceDAO {
	return &InvoiceDAO{
		conn: conn,
		user: user,
	}
}

func (dao *InvoiceDAO) FindByFileName(filename string) (*entity.Invoice, error) {
	stmt, err := dao.conn.Prepare(qryInvoiceByFilename)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var invoice entity.Invoice
	var userId int64

	qryErr := stmt.QueryRow(filename).Scan(
		&invoice.Id,
		&userId,
		&invoice.FileName,
		&invoice.Number,
		&invoice.MarketDate,
		&invoice.BillingDate,
		&invoice.RawValue,
		&invoice.NetValue,
	)

	if qryErr == sql.ErrNoRows {
		return nil, nil
	} else if qryErr != nil {
		return nil, qryErr
	}

	if userId != dao.user.Id {
		utils.GetError("InvoiceDAO::FindByFileName", "ERR_SYS_001", "user id mismatch")
	}

	invoice.User = dao.user

	return &invoice, nil
}

func (dao *InvoiceDAO) CreateInvoice(invoiceInput *inputData.Invoice) (*entity.Invoice, error) {
	stmt, err := dao.conn.Prepare(insertInvoiceStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	marketDate, err := time.Parse(time.RFC3339, invoiceInput.MarketDate)
	if err != nil {
		return nil, err
	}

	billingDate, err := time.Parse(time.RFC3339, invoiceInput.BillingDate)
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(
		dao.user.Id,
		invoiceInput.FileName,
		invoiceInput.InvoiceNum,
		marketDate,
		billingDate,
		invoiceInput.RawValue,
		invoiceInput.NetValue,
	)
	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	invoice := entity.Invoice{
		Id:          lastId,
		Number:      invoiceInput.InvoiceNum,
		User:        dao.user,
		FileName:    invoiceInput.FileName,
		MarketDate:  marketDate,
		BillingDate: billingDate,
		RawValue:    invoiceInput.RawValue,
		NetValue:    invoiceInput.NetValue,
	}

	log.Printf("created invoice [%d, %s]\n", invoice.Id, invoice.FileName)

	return &invoice, nil
}
