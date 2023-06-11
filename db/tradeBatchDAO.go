package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

type TradeBatchDAO struct {
	tx         *sql.Tx
	tradeBatch *entity.TradeBatch
}

func GetTradeBatchDAO(tx *sql.Tx, tradeBatch *entity.TradeBatch) *TradeBatchDAO {
	return &TradeBatchDAO{
		tx:         tx,
		tradeBatch: tradeBatch,
	}
}

func (dao *TradeBatchDAO) GetTradeBatch() (*entity.TradeBatch, error) {
	query := `SELECT
		trb_id,
		tgr_id,
		trb_shr_loss,
		trb_shr_results,
		trb_total_shr_tax,
		trb_total_shr_trade,
		trb_bdr_loss,
		trb_bdr_results,
		trb_total_bdr_tax,
		trb_etf_loss,
		trb_etf_results,
		trb_total_etf_tax
	FROM trade_batch
	WHERE usr_id = ?
	  AND trb_start_date = ?`

	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	tradeBatch := dao.tradeBatch
	var taxGroupId int64
	var tradeBatchRec entity.TradeBatch
	var shrData entity.TradeBatchData
	var bdrData entity.TradeBatchData
	var etfData entity.TradeBatchData

	err = stmt.QueryRow(
		tradeBatch.User.Id,
		tradeBatch.StartDate.Format(time.RFC3339),
	).Scan(
		&tradeBatchRec.Id,
		&taxGroupId,
		&shrData.AccLoss,
		&shrData.Results,
		&shrData.TotalTax,
		&shrData.TotalTrade,
		&bdrData.AccLoss,
		&bdrData.Results,
		&bdrData.TotalTax,
		&etfData.AccLoss,
		&etfData.Results,
		&etfData.TotalTax,
	)

	if err == sql.ErrNoRows {
		log.Printf(
			"TradeBatchDAO.GetTradeBatch: not found [usr = %d, startDate = %s]",
			tradeBatch.User.Id,
			tradeBatch.StartDate,
		)
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	log.Printf(
		"TradeBatchDAO.GetTradeBatch: found trade batch [%d, %s]",
		tradeBatchRec.Id,
		tradeBatch.StartDate.Format("2006-01-01"),
	)

	taxGroup := entity.TaxGroup{
		Id: taxGroupId,
	}

	taxGroupDAO := GetTaxGroupDAO(dao.tx, &taxGroup)
	taxGroupRec, err := taxGroupDAO.GetTaxGroup()

	if err != nil {
		return nil, err
	}

	tradeBatchRec.User = tradeBatch.User
	tradeBatchRec.TaxGroup = taxGroupRec
	tradeBatchRec.StartDate = tradeBatch.StartDate
	tradeBatchRec.Shr = &shrData
	tradeBatchRec.Bdr = &bdrData
	tradeBatchRec.Etf = &etfData

	return &tradeBatchRec, nil
}

func (dao *TradeBatchDAO) GetLastTradeBatch() (*entity.TradeBatch, error) {
	query := `SELECT
		trb_id,
		tgr_id,
		trb_start_date,
		trb_shr_loss,
		trb_shr_results,
		trb_total_shr_tax,
		trb_total_shr_trade,
		trb_bdr_loss,
		trb_bdr_results,
		trb_total_bdr_tax,
		trb_etf_loss,
		trb_etf_results,
		trb_total_etf_tax
	FROM trade_batch
	WHERE usr_id = ?
	ORDER BY trb_start_date DESC
	LIMIT 1`

	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	tradeBatch := dao.tradeBatch
	var taxGroupId int64
	var tradeBatchRec entity.TradeBatch
	var shrData entity.TradeBatchData
	var bdrData entity.TradeBatchData
	var etfData entity.TradeBatchData

	err = stmt.QueryRow(
		tradeBatch.User.Id,
	).Scan(
		&tradeBatchRec.Id,
		&taxGroupId,
		&tradeBatchRec.StartDate,
		&shrData.AccLoss,
		&shrData.Results,
		&shrData.TotalTax,
		&shrData.TotalTrade,
		&bdrData.AccLoss,
		&bdrData.Results,
		&bdrData.TotalTax,
		&etfData.AccLoss,
		&etfData.Results,
		&etfData.TotalTax,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	log.Printf(
		"TradeBatchDAO.GetLastTradeBatch: found last trade batch [%d, %s]",
		tradeBatchRec.Id,
		tradeBatch.StartDate.Format(time.RFC3339),
	)

	taxGroup := entity.TaxGroup{
		Id: taxGroupId,
	}

	taxGroupDAO := GetTaxGroupDAO(dao.tx, &taxGroup)
	taxGroupRec, err := taxGroupDAO.GetTaxGroup()

	if err != nil {
		return nil, err
	}

	tradeBatchRec.User = tradeBatch.User
	tradeBatchRec.TaxGroup = taxGroupRec
	tradeBatchRec.StartDate = tradeBatch.StartDate
	tradeBatchRec.Shr = &shrData
	tradeBatchRec.Bdr = &bdrData
	tradeBatchRec.Etf = &etfData

	return &tradeBatchRec, nil
}

func (dao *TradeBatchDAO) CreateTradeBatch() (*entity.TradeBatch, error) {
	insertStmt := `INSERT INTO trade_batch (
		tgr_id,
		usr_id,
		trb_start_date,
		trb_shr_loss,
		trb_shr_results,
		trb_total_shr_tax,
		trb_total_shr_trade,
		trb_bdr_loss,
		trb_bdr_results,
		trb_total_bdr_tax,
		trb_etf_loss,
		trb_etf_results,
		trb_total_etf_tax
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	tradeBatch := dao.tradeBatch

	res, err := stmt.Exec(
		tradeBatch.TaxGroup.Id,
		tradeBatch.User.Id,
		tradeBatch.StartDate,
		tradeBatch.Shr.AccLoss,
		tradeBatch.Shr.Results,
		tradeBatch.Shr.TotalTax,
		tradeBatch.Shr.TotalTrade,
		tradeBatch.Bdr.AccLoss,
		tradeBatch.Bdr.Results,
		tradeBatch.Bdr.TotalTax,
		tradeBatch.Etf.AccLoss,
		tradeBatch.Etf.Results,
		tradeBatch.Etf.TotalTax,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	tradeBatchRec := &entity.TradeBatch{
		Id:        lastId,
		User:      tradeBatch.User,
		TaxGroup:  tradeBatch.TaxGroup,
		StartDate: tradeBatch.StartDate,
		Shr:       tradeBatch.Shr,
		Bdr:       tradeBatch.Bdr,
		Etf:       tradeBatch.Etf,
	}

	log.Printf(
		"TradeBatchDAO.CreateTradeBatch: created trade batch [%d, %s]",
		tradeBatchRec.Id,
		tradeBatchRec.StartDate.Format(time.RFC3339),
	)

	return tradeBatchRec, nil
}

func (dao *TradeBatchDAO) UpdateTradeBatch() error {
	updateStmt := `UPDATE trade_batch SET
		trb_shr_loss = ?,
		trb_shr_results = ?,
		trb_total_shr_tax = ?,
		trb_total_shr_trade = ?,
		trb_bdr_loss = ?,
		trb_bdr_results = ?,
		trb_total_bdr_tax = ?,
		trb_etf_loss = ?,
		trb_etf_results = ?,
		trb_total_etf_tax = ?
	WHERE trb_id = ?`

	stmt, err := dao.tx.Prepare(updateStmt)

	if err != nil {
		return nil
	}

	defer stmt.Close()

	tradeBatch := dao.tradeBatch

	res, err := stmt.Exec(
		tradeBatch.Shr.AccLoss,
		tradeBatch.Shr.Results,
		tradeBatch.Shr.TotalTax,
		tradeBatch.Shr.TotalTrade,
		tradeBatch.Bdr.AccLoss,
		tradeBatch.Bdr.Results,
		tradeBatch.Bdr.TotalTax,
		tradeBatch.Etf.AccLoss,
		tradeBatch.Etf.Results,
		tradeBatch.Etf.TotalTax,
		tradeBatch.Id,
	)

	if err != nil {
		return err
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowCnt != 1 {
		details := fmt.Sprintf("expected 1 row, found %d rows", rowCnt)
		return utils.GetError("TradeBatchDAO.UpdateTradeBatch", "ERR_DB_001", details)
	}

	log.Printf(
		"TradeBatchDAO.UpdateTradeBatch: record updated [%d]",
		tradeBatch.Id,
	)

	return nil
}
