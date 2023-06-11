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

func (tbsvc *TradeBatchService) getNewTradeData(lastTradeData *entity.TradeBatchData) *entity.TradeBatchData {
	var tradeData entity.TradeBatchData

	if lastTradeData == nil {
		return &tradeData
	}

	la := lastTradeData.AccLoss
	lr := lastTradeData.Results
	lt := lastTradeData.TotalTax

	accLoss := la + lr - lt

	if accLoss > 0 {
		accLoss = 0
	}

	tradeData.AccLoss = accLoss
	tradeData.Results = 0.0
	tradeData.TotalTax = 0.0
	tradeData.TotalTrade = 0.0

	return &tradeData
}

func (tbsvc *TradeBatchService) adjustTradeBatchTaxes() *entity.TradeBatch {
	tradeBatch := tbsvc.tradeBatch

	var shrIRFee float64
	var bdrIRFee float64
	var etfIRFee float64
	var irFeeBaseValue float64 = 0.0

	shrTradeData := tradeBatch.Shr
	bdrTradeData := tradeBatch.Bdr
	etfTradeData := tradeBatch.Etf

	totalAccLoss := shrTradeData.AccLoss + bdrTradeData.AccLoss + etfTradeData.AccLoss
	totalResults := shrTradeData.Results + bdrTradeData.Results + etfTradeData.Results
	totalTaxes := shrTradeData.TotalTax + bdrTradeData.TotalTax + etfTradeData.TotalTax
	irExemptByLoss := (totalResults - totalTaxes) <= totalAccLoss

	// SHR
	shrIRExcemptByLimit := (shrTradeData.TotalTrade <= constants.TaxRates.IR_EXPT_LIMIT)
	shrCurrentResults := shrTradeData.Results - shrTradeData.TotalTax
	shrIRExcemptByResuls := shrCurrentResults <= 0

	if shrIRExcemptByLimit || shrIRExcemptByResuls || irExemptByLoss {
		shrIRFee = 0.0
	} else {
		shrIRFee = shrCurrentResults * constants.TaxRates.IRFEE
		irFeeBaseValue = irFeeBaseValue + shrCurrentResults
	}

	// BDR
	bdrCurrentResults := bdrTradeData.Results - bdrTradeData.TotalTax
	bdrIRExcempByResults := bdrCurrentResults <= 0
	if bdrIRExcempByResults || irExemptByLoss {
		bdrIRFee = 0.0
	} else {
		bdrIRFee = bdrCurrentResults * constants.TaxRates.IRFEE
		irFeeBaseValue = irFeeBaseValue + bdrCurrentResults
	}

	// ETF
	etfCurrentResults := etfTradeData.Results - etfTradeData.TotalTax
	etfIRExcempByResults := etfCurrentResults <= 0
	if etfIRExcempByResults || irExemptByLoss {
		etfIRFee = 0.0
	} else {
		etfIRFee = etfCurrentResults * constants.TaxRates.IRFEE
		irFeeBaseValue = irFeeBaseValue + etfCurrentResults
	}

	taxGroup := tradeBatch.TaxGroup

	// find IRFEE from tax group - if not exists then create
	irFeeInstance := taxGroup.GetTaxInstanceByCode(constants.TaxTypes.IRFEE)

	irFeeInstance.TaxValue = (shrIRFee + bdrIRFee + etfIRFee)
	irFeeInstance.BaseValue = irFeeBaseValue

	taxGroup.Taxes[0] = *irFeeInstance // TODO taxGroup.Taxes needs to use ptrs

	log.Printf(
		"TradeBatchService.adjustTradeBatchTaxes: id = %d",
		irFeeInstance.Id,
	)

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

	var lastShrData, lastBdrData, lastEftData *entity.TradeBatchData

	if lastTradeBatch != nil {
		lastShrData = lastTradeBatch.Shr
		lastBdrData = lastTradeBatch.Bdr
		lastEftData = lastTradeBatch.Etf
	}

	newShrData := tbsvc.getNewTradeData(lastShrData)
	newBdrData := tbsvc.getNewTradeData(lastBdrData)
	newEtfData := tbsvc.getNewTradeData(lastEftData)

	user := tbsvc.user

	tradeBatch := &entity.TradeBatch{
		User:      user,
		StartDate: utils.ToFirstDayOfMonth(marketDate),
		TaxGroup:  taxGroup,
		Shr:       newShrData,
		Bdr:       newBdrData,
		Etf:       newEtfData,
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
	company := trade.Item.Company

	shrTradeData := tradeBatch.Shr
	bdrTradeData := tradeBatch.Bdr

	if company.BDR {
		bdrTradeData.Results = bdrTradeData.Results + trade.RawResults
		bdrTradeData.TotalTax = bdrTradeData.TotalTax + trade.TotalTax
	} else if company.ETF {
		// TODO implement support for ETF
	} else {
		itemTrade := trade.Item.Price * float64(trade.Item.Qty)

		shrTradeData.Results = shrTradeData.Results + trade.RawResults
		shrTradeData.TotalTax = shrTradeData.TotalTax + trade.TotalTax
		shrTradeData.TotalTrade = shrTradeData.TotalTrade + itemTrade
	}

	log.Printf(
		"TradeBatchService.ProcessTrade: id = %d",
		tradeBatch.Id,
	)

	return tbsvc.adjustTradeBatchTaxes()
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
