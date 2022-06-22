package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type EarningTaxService struct {
	earningTaxDAO *db.EarningTaxDAO
	earning       *entity.Earning
	tax           *entity.Tax
}

func GetEarningTaxService(tx *sql.Tx, earning *entity.Earning, tax *entity.Tax) *EarningTaxService {
	return &EarningTaxService{
		earningTaxDAO: db.GetEarningTaxDAO(tx, earning, tax),
		earning:       earning,
		tax:           tax,
	}
}

func (earningTaxService *EarningTaxService) insert(taxInput *inputData.Tax) (*entity.EarningTax, error) {
	log.Printf("creating earningTax [%s, %f]\n", taxInput.Code, taxInput.Value)
	return earningTaxService.earningTaxDAO.CreateEarningTax(taxInput)
}

func (earningTaxService *EarningTaxService) UpsertEarningTax(taxInput *inputData.Tax) (*entity.EarningTax, error) {
	earningTax, err := earningTaxService.earningTaxDAO.FindEarningTax()
	if err != nil {
		return nil, err
	}

	if earningTax == nil {
		return earningTaxService.insert(taxInput)
	}

	// TODO implement support for udpate

	log.Printf("nothing to be done for earning tax [%s, %f]\n", earningTax.Tax.Code, earningTax.Value)

	return earningTax, nil
}
