package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

type TaxInstanceDAO struct {
	tx          *sql.Tx
	taxInstance *entity.TaxInstance
}

func GetTaxInstanceDAO(tx *sql.Tx, taxInstance *entity.TaxInstance) *TaxInstanceDAO {
	return &TaxInstanceDAO{
		tx:          tx,
		taxInstance: taxInstance,
	}
}

func (dao *TaxInstanceDAO) FindTaxInstance() (*entity.TaxInstance, error) {
	query := `SELECT
		tin.tin_id,
		tin.tin_market_date,
		tin.tin_tax_value,
		tin.tin_base_value,
		tin.tin_tax_rate,
		tax.tax_id,
		tax.tax_source,
		tax.tax_rate
	FROM tax_instance tin
	INNER JOIN tax ON tin.tax_id = tax.tax_id
	WHERE tin.tgr_id = ? AND
		tax.tax_code = ?`

	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return nil, err
	}

	taxGroupId := dao.taxInstance.TaxGroupId
	taxCode := dao.taxInstance.Tax.Code
	taxInstanceRec := entity.TaxInstance{}
	taxRec := entity.Tax{}

	err = stmt.QueryRow(
		taxGroupId,
		taxCode,
	).Scan(
		&taxInstanceRec.Id,
		&taxInstanceRec.MarketDate,
		&taxInstanceRec.TaxValue,
		&taxInstanceRec.BaseValue,
		&taxInstanceRec.TaxRate,
		&taxRec.Id,
		&taxRec.Source,
		&taxRec.Rate,
	)

	if err != nil {
		return nil, err
	}

	taxRec.Code = taxCode
	taxInstanceRec.TaxGroupId = dao.taxInstance.TaxGroupId
	taxInstanceRec.Tax = &taxRec

	return &taxInstanceRec, nil
}

func (dao *TaxInstanceDAO) CreateTaxInstance() (*entity.TaxInstance, error) {
	insertStmt := `INSERT INTO tax_instance (
	  tgr_id,
	  tax_id,
	  tin_market_date,
	  tin_tax_value,
	  tin_base_value,
	  tin_tax_rate
	) VALUES (?,?,?,?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	taxInstance := dao.taxInstance
	tax := taxInstance.Tax

	if tax == nil {
		err := utils.GetError("dao.TaxInstanceDAO.CreateTaxInstance", "", "tax is null")
		return nil, err
	}

	res, err := stmt.Exec(
		taxInstance.TaxGroupId,
		tax.Id,
		taxInstance.MarketDate,
		taxInstance.TaxValue,
		taxInstance.BaseValue,
		taxInstance.TaxRate,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	taxInstanceRec := &entity.TaxInstance{
		Id:         lastId,
		TaxGroupId: taxInstance.TaxGroupId,
		Tax:        tax,
		MarketDate: taxInstance.MarketDate,
		TaxValue:   taxInstance.TaxValue,
		BaseValue:  taxInstance.BaseValue,
		TaxRate:    taxInstance.TaxRate,
	}

	log.Printf(
		"created tax instance [%d, g = %d, %s, tv = %.4f, bv = %.4f, tr = %f]\n",
		taxInstanceRec.Id,
		taxInstanceRec.TaxGroupId,
		taxInstanceRec.Tax.Code,
		taxInstanceRec.TaxValue,
		taxInstanceRec.BaseValue,
		taxInstanceRec.TaxValue,
	)

	return taxInstanceRec, nil
}

func (dao *TaxInstanceDAO) UpdateTaxInstance() error {
	updateStmt := `UPDATE tax_instance SET
		tin_tax_value = ?,
		tin_base_value = ?,
		tin_tax_rate = ?
	WHERE tin_id = ?`

	stmt, err := dao.tx.Prepare(updateStmt)

	if err != nil {
		return err
	}

	defer stmt.Close()

	taxInstance := dao.taxInstance

	res, err := stmt.Exec(
		taxInstance.TaxValue,
		taxInstance.BaseValue,
		taxInstance.TaxRate,
		taxInstance.Id,
	)

	if err != nil {
		return err
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowCnt != 1 {
		log.Printf(
			"TaxInstanceDAO.UpdateTaxInstance: error updating tax instance [%d, tv = %f, bv = %f]",
			taxInstance.Id,
			taxInstance.TaxValue,
			taxInstance.BaseValue,
		)
		details := fmt.Sprintf("expected 1 row, found %d rows", rowCnt)
		return utils.GetError("TaxInstanceDAO.UpdateTaxInstance", "ERR_DB_001", details)
	}

	log.Printf(
		"TaxInstanceDAO.UpdateTaxInstance: record updated [%d, %s, tv = %f, bv = %f]",
		taxInstance.Id,
		taxInstance.Tax.Code,
		taxInstance.TaxValue,
		taxInstance.BaseValue,
	)

	return nil
}
