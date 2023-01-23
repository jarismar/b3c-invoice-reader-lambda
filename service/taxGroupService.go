package service

import (
	"database/sql"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/store"
	"github.com/jarismar/b3c-service-entities/entity"
)

type TaxGroupService struct {
	tx       *sql.Tx
	taxGroup *entity.TaxGroup
	taxStore *store.TaxStore
}

func GetTaxGroupService(
	tx *sql.Tx,
	taxGroup *entity.TaxGroup,
	taxStore *store.TaxStore,
) *TaxGroupService {
	return &TaxGroupService{
		tx:       tx,
		taxGroup: taxGroup,
		taxStore: taxStore,
	}
}

func (tgsvc *TaxGroupService) CreateTaxGroup() (*entity.TaxGroup, error) {
	taxGroupDAO := db.GetTaxGroupDAO(
		tgsvc.tx,
		tgsvc.taxGroup,
	)

	taxGroupRec, err := taxGroupDAO.CreateTaxGroup()

	if err != nil {
		return nil, err
	}

	taxInstances := tgsvc.taxGroup.Taxes
	taxInstanceRecs := make([]entity.TaxInstance, 0, len(taxInstances))

	var taxRec *entity.Tax

	for _, taxInstance := range taxInstances {
		if tgsvc.taxStore.Has(taxInstance.Tax) {
			taxRec = tgsvc.taxStore.Get(taxInstance.Tax)
		} else {
			taxDAO := db.GetTaxDAO(tgsvc.tx, taxInstance.Tax)
			taxRec, err = taxDAO.CreateTax()
			if err != nil {
				return nil, err
			}
			tgsvc.taxStore.Put(taxRec)
		}

		taxInstanceDAO := db.GetTaxInstanceDAO(tgsvc.tx, &entity.TaxInstance{
			TaxGroupId: taxGroupRec.Id,
			Tax:        taxRec,
			MarketDate: taxInstance.MarketDate,
			TaxValue:   taxInstance.TaxValue,
			BaseValue:  taxInstance.BaseValue,
			TaxRate:    taxInstance.TaxRate,
		})

		taxInstanceRec, err := taxInstanceDAO.CreateTaxInstance()

		if err != nil {
			return nil, err
		}

		taxInstanceRec.TaxGroupId = taxGroupRec.Id

		taxInstanceRecs = append(taxInstanceRecs, *taxInstanceRec)
	}

	taxGroupRec.Taxes = taxInstanceRecs

	return taxGroupRec, nil
}

func (tgsvc *TaxGroupService) UpdateTaxGroup() error {
	taxInstances := tgsvc.taxGroup.Taxes

	for _, taxInstance := range taxInstances {
		taxInstanceDAO := db.GetTaxInstanceDAO(tgsvc.tx, &taxInstance)
		err := taxInstanceDAO.UpdateTaxInstance()

		if err != nil {
			return err
		}
	}

	return nil
}
