package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

type CompanyBatchDAO struct {
	tx           *sql.Tx
	companyBatch *entity.CompanyBatch
}

func GetCompanyBatchDAO(tx *sql.Tx, companyBatch *entity.CompanyBatch) *CompanyBatchDAO {
	return &CompanyBatchDAO{
		tx:           tx,
		companyBatch: companyBatch,
	}
}

func (dao *CompanyBatchDAO) GetCompanyBatch() (*entity.CompanyBatch, error) {
	query := `SELECT
		cbt_id,
		cbt_qty,
		cbt_avg_price,
		cbt_total_price
	FROM company_batch
	WHERE cmp_id = ?
		AND usr_id = ?
		AND cbt_qty > 0`

	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	company := dao.companyBatch.Company
	user := dao.companyBatch.User
	var companyBatchRec entity.CompanyBatch

	err = stmt.QueryRow(
		company.Id,
		user.Id,
	).Scan(
		&companyBatchRec.Id,
		&companyBatchRec.Qty,
		&companyBatchRec.AvgPrice,
		&companyBatchRec.TotalPrice,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	companyBatchRec.Company = company
	companyBatchRec.User = user

	log.Printf(
		"CompanyBatchDAO.GetCompanyBatch: found company batch [%d, %s, %d]",
		companyBatchRec.Id,
		companyBatchRec.Company.Code,
		companyBatchRec.Qty,
	)

	return &companyBatchRec, nil
}

func (dao *CompanyBatchDAO) UpdateCompanyBatch() (*entity.CompanyBatch, error) {
	updateStmt := `UPDATE company_batch SET
		cbt_qty = ?,
		cbt_avg_price = ?,
		cbt_total_price = ?
	WHERE cbt_id = ?`

	stmt, err := dao.tx.Prepare(updateStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	companyBatch := dao.companyBatch

	res, err := stmt.Exec(
		companyBatch.Qty,
		companyBatch.AvgPrice,
		companyBatch.TotalPrice,
		companyBatch.Id,
	)

	if err != nil {
		return nil, err
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowCnt != 1 {
		details := fmt.Sprintf("expected 1 row, found %d rows", rowCnt)
		return nil, utils.GetError("CompanyBatchDAO.UpdateCompanyBatch", "ERR_DB_001", details)
	}

	log.Printf(
		"CompanyBatchDAO.UpdateCompanyBatch: updated company batch [%d, %s, qty = %d, avg = %.4f, tot = %.4f]\n",
		companyBatch.Id,
		companyBatch.Company.Code,
		companyBatch.Qty,
		companyBatch.AvgPrice,
		companyBatch.TotalPrice,
	)

	return companyBatch, nil
}

func (dao *CompanyBatchDAO) CreateCompanyBatch() (*entity.CompanyBatch, error) {
	insertStmt := `INSERT INTO company_batch (
		cmp_id,
		usr_id,
		cbt_start_date,
		cbt_qty,
		cbt_avg_price,
		cbt_total_price
	) VALUES (?,?,?,?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	companyBatch := dao.companyBatch
	company := companyBatch.Company
	user := companyBatch.User

	res, err := stmt.Exec(
		company.Id,
		user.Id,
		companyBatch.StartDate,
		companyBatch.Qty,
		companyBatch.AvgPrice,
		companyBatch.TotalPrice,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	companyBatchRec := entity.CompanyBatch{
		Id:         lastId,
		User:       user,
		Company:    company,
		StartDate:  companyBatch.StartDate,
		Qty:        companyBatch.Qty,
		AvgPrice:   companyBatch.AvgPrice,
		TotalPrice: companyBatch.TotalPrice,
	}

	log.Printf(
		"CompanyBatchDAO.CreateCompanyBatch: created new company batch [%d, %s, %d]",
		companyBatchRec.Id,
		companyBatchRec.Company.Code,
		companyBatchRec.Qty,
	)

	return &companyBatchRec, nil
}
