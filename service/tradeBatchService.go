package service

import (
	"database/sql"
	"log"
	"time"

	"github.com/jarismar/b3c-invoice-reader-lambda/constants"
	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/store"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

type TradeBatchService struct {
	tx         *sql.Tx
	user       *entity.User
	tradeBatch *entity.TradeBatch
	taxStore   *store.TaxStore
}

func GetTradeBatchService(
	tx *sql.Tx,
	user *entity.User,
	tradeBatch *entity.TradeBatch,
	taxStore *store.TaxStore,
) *TradeBatchService {
	return &TradeBatchService{
		tx:         tx,
		user:       user,
		tradeBatch: tradeBatch,
		taxStore:   taxStore,
	}
}

func (tbsvc *TradeBatchService) getTaxGroup(marketDate time.Time) (*entity.TaxGroup, error) {
	groupId, err := utils.GetTaxGroupIdFromTime(
		marketDate,
		constants.TaxGroupPrefix.TRADE_BATCH,
	)

	if err != nil {
		return nil, err
	}

	taxes := make([]entity.TaxInstance, 0)
	taxGroup := &entity.TaxGroup{
		Source:     entity.TRB,
		ExternalId: groupId,
	}

	irFeeTaxInstance := entity.TaxInstance{
		MarketDate: marketDate,
		TaxValue:   0,
		BaseValue:  0,
		TaxRate:    constants.TaxRates.IRFEE,
		Tax: &entity.Tax{
			Code:   constants.TaxTypes.IRFEE,
			Source: constants.TaxSources.TRADE_BATCH,
			Rate:   constants.TaxRates.IRFEE,
		},
	}

	taxes = append(taxes, irFeeTaxInstance)
	taxGroup.Taxes = taxes

	taxGroupService := GetTaxGroupService(tbsvc.tx, taxGroup, tbsvc.taxStore)

	return taxGroupService.CreateTaxGroup()
}

func (tbsvc *TradeBatchService) getNewAccLoss(lastTradeBatch *entity.TradeBatch) float64 {
	if lastTradeBatch == nil {
		return 0.0
	}

	lAccLoss := lastTradeBatch.AccLoss
	lResults := lastTradeBatch.CurrentResults
	accLoss := lAccLoss + lResults

	if accLoss > 0 {
		return 0.0
	}

	return accLoss
}

func (tbsvc *TradeBatchService) adjustTradeBatchTaxes(tradeBatch *entity.TradeBatch) *entity.TradeBatch {
	var irFee float64

	taxGroup := tradeBatch.TaxGroup
	currentResults := tradeBatch.CurrentResults - tradeBatch.TotalTax
	irexcemptByLimit := (tradeBatch.TotalTrade <= constants.TaxRates.IR_EXPT_LIMIT)
	irexcemptByResuls := currentResults <= 0
	irexcemptByLoss := currentResults <= tradeBatch.AccLoss

	if irexcemptByLimit || irexcemptByResuls || irexcemptByLoss {
		irFee = 0
	} else {
		irFee = currentResults * constants.TaxRates.IRFEE
	}

	// find IRFEE from tax group - if not exists then create
	irFeeInstance := taxGroup.GetTaxInstanceByCode(constants.TaxTypes.IRFEE)

	taxBaseValue := currentResults - tradeBatch.AccLoss

	if taxBaseValue < 0 {
		taxBaseValue = 0
	}

	irFeeInstance.TaxValue = irFee
	irFeeInstance.BaseValue = taxBaseValue

	return tradeBatch
}

func (tbsvc *TradeBatchService) createTradeBatch(
	marketDate time.Time,
	lastTradeBatch *entity.TradeBatch,
) (*entity.TradeBatch, error) {
	taxGroup, err := tbsvc.getTaxGroup(marketDate)

	if err != nil {
		return nil, err
	}

	user := tbsvc.user
	accLoss := tbsvc.getNewAccLoss(lastTradeBatch)

	tradeBatch := &entity.TradeBatch{
		User:           user,
		StartDate:      utils.ToFirstDayOfMonth(marketDate),
		TaxGroup:       taxGroup,
		AccLoss:        accLoss,
		CurrentResults: 0.0,
		TotalTrade:     0.0,
		TotalTax:       0.0,
	}

	tradeBatchDAO := db.GetTradeBatchDAO(
		tbsvc.tx,
		tradeBatch,
	)

	return tradeBatchDAO.CreateTradeBatch()
}

func (tbsvc *TradeBatchService) FindTradeBatch(marketDate time.Time) (*entity.TradeBatch, error) {
	user := tbsvc.user

	tradeBatch := entity.TradeBatch{
		User:      user,
		StartDate: utils.ToFirstDayOfMonth(marketDate),
	}

	tradeBatchDAO := db.GetTradeBatchDAO(
		tbsvc.tx,
		&tradeBatch,
	)

	tradeBatchRec, err := tradeBatchDAO.GetTradeBatch()

	if err != nil {
		return nil, err
	}

	if tradeBatchRec == nil {
		lastTradeBatch, err := tradeBatchDAO.GetLastTradeBatch()

		if err != nil {
			return nil, err
		}

		return tbsvc.createTradeBatch(marketDate, lastTradeBatch)
	}

	log.Printf(
		"TradeBatchService.FindTradeBatch: returning tradeBatch [%d, %s]",
		tradeBatchRec.Id,
		tradeBatchRec.StartDate.Format(time.RFC3339),
	)

	return tradeBatchRec, nil
}

func (tbsvc *TradeBatchService) ProcessTrade(trade *entity.Trade) *entity.TradeBatch {
	tradeBatch := tbsvc.tradeBatch

	tradeBatch.CurrentResults = tradeBatch.CurrentResults + trade.RawResults
	tradeBatch.TotalTrade = tradeBatch.TotalTrade + trade.RawResults
	tradeBatch.TotalTax = tradeBatch.TotalTax + trade.TotalTax

	return tbsvc.adjustTradeBatchTaxes(tradeBatch)
}

func (tbsvc *TradeBatchService) SaveTradeBatch() (*entity.TradeBatch, error) {
	taxGroupService := GetTaxGroupService(
		tbsvc.tx,
		tbsvc.tradeBatch.TaxGroup,
		tbsvc.taxStore,
	)

	err := taxGroupService.UpdateTaxGroup()

	if err != nil {
		return nil, err
	}

	tradeBatchDAO := db.GetTradeBatchDAO(
		tbsvc.tx,
		tbsvc.tradeBatch,
	)

	err = tradeBatchDAO.UpdateTradeBatch()

	if err != nil {
		return nil, err
	}

	return tbsvc.tradeBatch, nil
}
