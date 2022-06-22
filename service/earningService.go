package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type EarningService struct {
	earningDAO *db.EarningDAO
	user       *entity.User
	company    *entity.Company
}

func GetEarningService(tx *sql.Tx, user *entity.User, company *entity.Company) *EarningService {
	return &EarningService{
		earningDAO: db.GetEarningDAO(tx, user, company),
		user:       user,
		company:    company,
	}
}

func (earningService *EarningService) insert(earningInput *inputData.EarningEntry) (*entity.Earning, error) {
	log.Printf("creating earning [%s %f]\n", earningService.company.Code, earningInput.NetValue)
	return earningService.earningDAO.CreateEarning(earningInput)
}

func (earningService *EarningService) UpsertEarning(earningInput *inputData.EarningEntry) (*entity.Earning, error) {
	earning, err := earningService.earningDAO.FindByUUID(earningInput.UUID)
	if err != nil {
		return nil, err
	}

	if earning == nil {
		return earningService.insert(earningInput)
	}

	// TODO implement support for earning update

	log.Printf("nothing to be done for earning [%d, %s %f]\n", earning.Id, earning.Company.Code, earning.NetValue)
	return earning, nil
}
