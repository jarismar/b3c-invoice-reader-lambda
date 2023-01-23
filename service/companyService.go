package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/store"
	"github.com/jarismar/b3c-service-entities/entity"
)

type CompanyService struct {
	tx           *sql.Tx
	company      *entity.Company
	companyStore *store.CompanyStore
}

func GetCompanyService(
	tx *sql.Tx,
	company *entity.Company,
	companyStore *store.CompanyStore,
) *CompanyService {
	return &CompanyService{
		tx:           tx,
		company:      company,
		companyStore: companyStore,
	}
}

func (csvc *CompanyService) UpsertCompany() (*entity.Company, error) {
	companyStore := csvc.companyStore
	company := csvc.company

	if companyStore.Has(company) {
		companyRec := companyStore.Get(csvc.company)

		log.Printf(
			"companyService.CreateCompany: returning company %s from cache",
			companyRec.Code,
		)

		return companyRec, nil
	}

	companyDAO := db.GetCompanyDAO(csvc.tx, company)
	companyRec, err := companyDAO.GetCompany()

	if err != nil {
		return nil, err
	}

	if companyRec == nil {
		companyRec, err = companyDAO.CreateCompany()

		if err != nil {
			return nil, err
		}
	}

	companyStore.Put(companyRec)
	return companyRec, nil
}
