package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type CompanyDAO struct {
	tx      *sql.Tx
	company *entity.Company
}

func GetCompanyDAO(tx *sql.Tx, company *entity.Company) *CompanyDAO {
	return &CompanyDAO{
		tx:      tx,
		company: company,
	}
}

func (dao *CompanyDAO) GetCompany() (*entity.Company, error) {
	query := `SELECT
		cmp_id,
		cmp_code,
		cmp_name
  	FROM company
  	WHERE cmp_code = ?`

	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var companyRec entity.Company

	err = stmt.QueryRow(dao.company.Code).Scan(
		&companyRec.Id,
		&companyRec.Code,
		&companyRec.Name,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	log.Printf(
		"companyDAO.CreateCompany: found company [%d, %s, %s]",
		companyRec.Id,
		companyRec.Code,
		companyRec.Name,
	)

	return &companyRec, nil
}

func (dao *CompanyDAO) CreateCompany() (*entity.Company, error) {
	insertStmt := `INSERT INTO company (
		cmp_code,
		cmp_name
	) VALUES (?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	company := dao.company

	res, err := stmt.Exec(
		company.Code,
		company.Name,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	companyRec := &entity.Company{
		Id:   lastId,
		Code: company.Code,
		Name: company.Name,
	}

	log.Printf(
		"companyDAO.CreateCompany: created new company [%d, %s, %s]",
		companyRec.Id,
		companyRec.Code,
		companyRec.Name,
	)

	return companyRec, nil
}
