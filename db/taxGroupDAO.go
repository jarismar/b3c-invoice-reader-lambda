package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/jarismar/b3c-service-entities/entity"
)

type TaxGroupDAO struct {
	tx       *sql.Tx
	taxGroup *entity.TaxGroup
}

func GetTaxGroupDAO(tx *sql.Tx, taxGroup *entity.TaxGroup) *TaxGroupDAO {
	return &TaxGroupDAO{
		tx:       tx,
		taxGroup: taxGroup,
	}
}

func (dao *TaxGroupDAO) GetTaxGroup() (*entity.TaxGroup, error) {
	query := `SELECT
		tgr.tgr_source,
		tgr.tgr_external_id,
		tin.tin_id,
		tin.tax_id,
		tin.tin_market_date,
		tin.tin_tax_value,
		tin.tin_base_value,
		tin.tin_tax_rate
		tax.tax_code,
		tax.tax_source,
		tax.tax_rate
		FROM tax_group tgr
	INNER JOIN tax_instance tin ON tgr.tgr_id = tin.tgr_id
	INNER JOIN tax ON tin.tax_id = tax.tax_id
	WHERE tgr_id = ?`

	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(dao.taxGroup.Id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var taxGroupRec entity.TaxGroup

	var rowNum int = 0
	var taxGroupSource string
	var taxGroupExtId int64
	var taxInstId int64
	var taxId int64
	var taxInstMarketDate time.Time
	var taxInstValue float64
	var taxInstBaseValue float64
	var taxInstRate float64
	var taxCode string
	var taxSource string
	var taxRate float64

	taxGroupRec.Taxes = make([]entity.TaxInstance, 0, 10)

	for rows.Next() {
		err := rows.Scan(
			&taxGroupSource,
			&taxGroupExtId,
			&taxInstId,
			&taxId,
			&taxInstMarketDate,
			&taxInstValue,
			&taxInstBaseValue,
			&taxInstRate,
			&taxCode,
			&taxSource,
			&taxRate,
		)

		if err != nil {
			return nil, err
		}

		if rowNum == 0 {
			taxGroupRec.Id = dao.taxGroup.Id
			taxGroupRec.Source = taxGroupSource
			taxGroupRec.ExternalId = taxGroupExtId
		}

		taxInstace := entity.TaxInstance{
			Id: taxInstId,
			Tax: &entity.Tax{
				Id:     taxId,
				Code:   taxCode,
				Source: taxSource,
				Rate:   taxRate,
			},
			MarketDate: taxInstMarketDate,
			TaxValue:   taxInstValue,
			BaseValue:  taxInstBaseValue,
			TaxRate:    taxInstRate,
		}

		taxGroupRec.Taxes = append(taxGroupRec.Taxes, taxInstace)
		rowNum = rowNum + 1
	}

	return &taxGroupRec, nil
}

func (dao *TaxGroupDAO) CreateTaxGroup() (*entity.TaxGroup, error) {
	insertStmt := `INSERT INTO tax_group (
		tgr_source,
		tgr_external_id
	) VALUES (?,?)
	`
	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		dao.taxGroup.Source,
		dao.taxGroup.ExternalId,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	taxGroup := &entity.TaxGroup{
		Id:         lastId,
		Source:     dao.taxGroup.Source,
		ExternalId: dao.taxGroup.ExternalId,
	}

	log.Printf(
		"taxGroupDAO.createTaxGroup: created new tax group [%d, %s, %d]",
		taxGroup.Id,
		taxGroup.Source,
		taxGroup.ExternalId,
	)

	return taxGroup, nil
}
