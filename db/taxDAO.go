package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type TaxDAO struct {
	tx  *sql.Tx
	tax *entity.Tax
}

func GetTaxDAO(tx *sql.Tx, tax *entity.Tax) *TaxDAO {
	return &TaxDAO{
		tx:  tx,
		tax: tax,
	}
}

func (dao *TaxDAO) CreateTax() (*entity.Tax, error) {
	insertStmt := `INSERT INTO tax (
		tax_code,
		tax_source,
		tax_rate
	) VALUES (?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		dao.tax.Code,
		dao.tax.Source,
		dao.tax.Rate,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	tax := &entity.Tax{
		Id:     lastId,
		Code:   dao.tax.Code,
		Source: dao.tax.Source,
		Rate:   dao.tax.Rate,
	}

	log.Printf(
		"dao.taxDAO: created tax [%d, %s, %s, %f]\n",
		tax.Id,
		tax.Code,
		tax.Source,
		tax.Rate,
	)

	return tax, nil
}
