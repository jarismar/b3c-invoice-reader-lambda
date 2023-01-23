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
		trb_acc_loss,
		trb_current_results,
		trb_total_trade,
		trb_total_tax
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

	err = stmt.QueryRow(
		tradeBatch.User.Id,
		tradeBatch.StartDate,
	).Scan(
		&tradeBatchRec.Id,
		&taxGroupId,
		&tradeBatchRec.AccLoss,
		&tradeBatchRec.CurrentResults,
		&tradeBatchRec.TotalTrade,
		&tradeBatchRec.TotalTax,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	log.Printf(
		"TradeBatchDAO.GetTradeBatch: found trade batch [%d, %s, %.4f]",
		tradeBatchRec.Id,
		tradeBatch.StartDate.Format("2006-01-01"),
		tradeBatch.CurrentResults,
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

	return &tradeBatchRec, nil
}

func (dao *TradeBatchDAO) GetLastTradeBatch() (*entity.TradeBatch, error) {
	query := `SELECT
		trb_id,
		tgr_id,
		trb_start_date,
		trb_acc_loss,
		trb_current_results,
		trb_total_trade,
		trb_total_tax
	FROM trade_batch
	WHERE usr_id = ?
	ORDER BY trb_start_date DESC
	LIMI 1`

	stmt, err := dao.tx.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	tradeBatch := dao.tradeBatch
	var taxGroupId int64
	var tradeBatchRec entity.TradeBatch

	err = stmt.QueryRow(
		tradeBatch.User.Id,
	).Scan(
		&tradeBatchRec.Id,
		&taxGroupId,
		&tradeBatchRec.AccLoss,
		&tradeBatchRec.CurrentResults,
		&tradeBatchRec.TotalTrade,
		&tradeBatchRec.TotalTax,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	log.Printf(
		"TradeBatchDAO.GetLastTradeBatch: found trade batch [%d, %s, %.4f]",
		tradeBatchRec.Id,
		tradeBatch.StartDate.Format(time.RFC3339),
		tradeBatch.CurrentResults,
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

	return &tradeBatchRec, nil
}

func (dao *TradeBatchDAO) CreateTradeBatch() (*entity.TradeBatch, error) {
	insertStmt := `INSERT INTO trade_batch (
		tgr_id,
		usr_id,
		trb_start_date,
		trb_acc_loss,
		trb_current_results,
		trb_total_trade,
		trb_total_tax
	) VALUES (?,?,?,?,?,?)`

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
		tradeBatch.AccLoss,
		tradeBatch.CurrentResults,
		tradeBatch.TotalTrade,
		tradeBatch.TotalTax,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	tradeBatchRec := &entity.TradeBatch{
		Id:             lastId,
		User:           tradeBatch.User,
		TaxGroup:       tradeBatch.TaxGroup,
		StartDate:      tradeBatch.StartDate,
		AccLoss:        tradeBatch.AccLoss,
		CurrentResults: tradeBatch.CurrentResults,
		TotalTrade:     tradeBatch.TotalTrade,
		TotalTax:       tradeBatch.TotalTax,
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
		trb_acc_loss = ?,
		trb_current_results = ?,
		trb_total_trade = ?,
		trb_total_tax = ?
	WHERE trb_id = ?`

	stmt, err := dao.tx.Prepare(updateStmt)

	if err != nil {
		return nil
	}

	defer stmt.Close()

	tradeBatch := dao.tradeBatch

	res, err := stmt.Exec(
		tradeBatch.AccLoss,
		tradeBatch.CurrentResults,
		tradeBatch.TotalTrade,
		tradeBatch.TotalTax,
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
		"TradeBatchDAO.UpdateTradeBatch: record updated [%d, loss = %f, cr = %f, tt = %f]",
		tradeBatch.Id,
		tradeBatch.AccLoss,
		tradeBatch.CurrentResults,
		tradeBatch.TotalTrade,
	)

	return nil
}
