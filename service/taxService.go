package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type TaxService struct {
	taxDAO *db.TaxDAO
}

func GetTaxService(conn *sql.Tx) *TaxService {
	return &TaxService{
		taxDAO: db.GetTaxDAO(conn),
	}
}

func (taxService *TaxService) insert(taxInput *inputData.Tax) (*entity.Tax, error) {
	log.Printf("creating new tax [%s, %s]\n", taxInput.Code, taxInput.Source)
	return taxService.taxDAO.CreateTax(taxInput)
}

func (taxService *TaxService) update(tax *entity.Tax) (*entity.Tax, error) {
	log.Printf("updating tax [%d, %s, %s]\n", tax.Id, tax.Code, tax.Source)
	err := taxService.taxDAO.UpdateTax(tax)
	if err != nil {
		return nil, err
	}
	return tax, nil
}

func (taxService *TaxService) UpsertTax(taxInput *inputData.Tax) (*entity.Tax, error) {
	tax, err := taxService.taxDAO.FindByCode(taxInput.Code)
	if err != nil {
		return nil, err
	}

	if tax == nil {
		return taxService.insert(taxInput)
	}

	if tax.Source != taxInput.Source {
		tax.Source = taxInput.Source
		return taxService.update(tax)
	}

	log.Printf("nothing to be done for tax [%d, %s, %s]\n", tax.Id, tax.Code, tax.Source)
	return tax, nil
}
