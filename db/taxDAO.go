package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

const qryTaxByCode = `SELECT
  tax_id,
  tax_code,
  tax_source
FROM tax
WHERE tax_code = ?`

const insertTaxStmt = "INSERT INTO tax (tax_code, tax_source) VALUES (?,?)"

const updateTaxStmt = "UPDATE tax SET tax_source = ? WHERE tax_id = ?"

type TaxDAO struct {
	conn *sql.Tx
}

func GetTaxDAO(conn *sql.Tx) *TaxDAO {
	return &TaxDAO{
		conn: conn,
	}
}

func (dao *TaxDAO) FindByCode(code string) (*entity.Tax, error) {
	stmt, err := dao.conn.Prepare(qryTaxByCode)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var tax entity.Tax

	qryErr := stmt.QueryRow(code).Scan(
		&tax.Id,
		&tax.Code,
		&tax.Source,
	)

	if qryErr == sql.ErrNoRows {
		return nil, nil
	} else if qryErr != nil {
		return nil, qryErr
	}

	return &tax, nil
}

func (dao *TaxDAO) CreateTax(taxInput *inputData.Tax) (*entity.Tax, error) {
	stmt, err := dao.conn.Prepare(insertTaxStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(taxInput.Code, taxInput.Source)
	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	tax := entity.Tax{
		Id:     lastId,
		Code:   taxInput.Code,
		Source: taxInput.Source,
	}

	log.Printf("created tax [%d, %s, %s]\n", tax.Id, tax.Code, tax.Source)

	return &tax, nil
}

func (dao *TaxDAO) UpdateTax(tax *entity.Tax) error {
	stmt, err := dao.conn.Prepare(updateTaxStmt)
	if err != nil {
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(tax.Source, tax.Id)
	if err != nil {
		return err
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowCnt != 1 {
		details := fmt.Sprintf("expected 1 found %d rows", rowCnt)
		return utils.GetError("TaxDAO::UpdateTax", "ERR_DB_001", details)
	}

	log.Printf("updated tax [%d, %s, %s] \n", tax.Id, tax.Code, tax.Source)

	return nil
}
