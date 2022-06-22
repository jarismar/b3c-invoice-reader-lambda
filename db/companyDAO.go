package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

const qryCompanyByCode = `SELECT
  cmp_id,
  cmp_code,
  cmp_name
FROM company
WHERE cmp_code = ?
`
const insertCompanyStmt = "INSERT INTO company (cmp_code, cmp_name) VALUES (?,?)"

const updateCompanyStmt = "UPDATE company SET cmp_name = ? WHERE cmp_id = ?"

type CompanyDAO struct {
	conn *sql.Tx
}

func GetCompanyDAO(conn *sql.Tx) *CompanyDAO {
	return &CompanyDAO{
		conn: conn,
	}
}

func (dao *CompanyDAO) FindByCode(code string) (*entity.Company, error) {
	stmt, err := dao.conn.Prepare(qryCompanyByCode)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var company entity.Company

	err = stmt.QueryRow(code).Scan(
		&company.Id,
		&company.Code,
		&company.Name,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &company, nil
}

func (dao *CompanyDAO) CreateCompany(inputCompany *inputData.Company) (*entity.Company, error) {
	stmt, err := dao.conn.Prepare(insertCompanyStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(inputCompany.Code, inputCompany.Name)
	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	company := entity.Company{
		Id:   lastId,
		Code: inputCompany.Code,
		Name: inputCompany.Name,
	}

	log.Printf("created company [%03d, %s, %s]\n", company.Id, company.Code, company.Name)
	return &company, nil
}

func (dao *CompanyDAO) UpdateCompany(company *entity.Company) error {
	stmt, err := dao.conn.Prepare(updateCompanyStmt)
	if err != nil {
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(company.Name, company.Name)
	if err != nil {
		return err
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowCnt != 1 {
		details := fmt.Sprintf("expected 1 found %d rows", rowCnt)
		return utils.GetError("CompanyDAO::UpdateCompany", "ERR_DB_001", details)
	}

	log.Printf("updated company [%03d, %s, %s]\n", company.Id, company.Code, company.Name)
	return nil
}
