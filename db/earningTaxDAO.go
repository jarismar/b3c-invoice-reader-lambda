package db

import (
	"database/sql"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

const qryEarningTaxByEarningAndTax = `SELECT
  eat_id,
  eat_value
FROM earning_tax
WHERE ear_id = ? AND tax_id = ?`

const insertEarningTaxStmt = `INSERT INTO earning_tax (
	ear_id,
	tax_id,
	eat_value
  ) VALUES (?,?,?)`

type EarningTaxDAO struct {
	conn    *sql.Tx
	earning *entity.Earning
	tax     *entity.Tax
}

func GetEarningTaxDAO(conn *sql.Tx, earning *entity.Earning, tax *entity.Tax) *EarningTaxDAO {
	return &EarningTaxDAO{
		conn:    conn,
		earning: earning,
		tax:     tax,
	}
}

func (dao *EarningTaxDAO) FindEarningTax() (*entity.EarningTax, error) {
	stmt, err := dao.conn.Prepare(qryEarningTaxByEarningAndTax)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var earningTax entity.EarningTax

	err = stmt.QueryRow(dao.earning.Id, dao.tax.Id).Scan(
		&earningTax.Id,
		&earningTax.Value,
	)

	if err != nil {
		return nil, err
	}

	earningTax.Tax = dao.tax

	return &earningTax, nil
}

func (dao *EarningTaxDAO) CreateEarningTax(taxInput *inputData.Tax) (*entity.EarningTax, error) {
	stmt, err := dao.conn.Prepare(insertEarningTaxStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		dao.earning.Id,
		dao.tax.Id,
		taxInput.Value,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	earningTax := entity.EarningTax{
		Id:    lastId,
		Tax:   dao.tax,
		Value: taxInput.Value,
	}

	return &earningTax, nil
}
