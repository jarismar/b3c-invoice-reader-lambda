package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type CompanyService struct {
	companyDAO *db.CompanyDAO
}

func GetCompanyService(tx *sql.Tx) *CompanyService {
	return &CompanyService{
		companyDAO: db.GetCompanyDAO(tx),
	}
}

func (companyService *CompanyService) insert(inputCompany *inputData.Company) (*entity.Company, error) {
	log.Printf("creating company [%s, %s]\n", inputCompany.Code, inputCompany.Name)
	return companyService.companyDAO.CreateCompany(inputCompany)
}

func (companyService *CompanyService) update(company *entity.Company) (*entity.Company, error) {
	log.Printf("updating company [%d, %s, %s]\n", company.Id, company.Code, company.Name)
	err := companyService.companyDAO.UpdateCompany(company)
	return company, err
}

func (companyService *CompanyService) UpsertCompany(inputCompany *inputData.Company) (*entity.Company, error) {
	company, err := companyService.companyDAO.FindByCode(inputCompany.Code)
	if err != nil {
		return nil, err
	}

	if company == nil {
		return companyService.insert(inputCompany)
	}

	if company.Name != inputCompany.Name {
		company.Name = inputCompany.Name
		return companyService.update(company)
	}

	log.Printf("nothing to be done for company [%d, %s]\n", company.Id, company.Code)
	return company, nil
}
