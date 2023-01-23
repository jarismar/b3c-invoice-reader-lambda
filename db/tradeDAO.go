package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type TradeDAO struct {
	tx    *sql.Tx
	trade *entity.Trade
}

func GetTradeDAO(tx *sql.Tx, trade *entity.Trade) *TradeDAO {
	return &TradeDAO{
		tx:    tx,
		trade: trade,
	}
}

func (dao *TradeDAO) CreateTrade() (*entity.Trade, error) {
	insertStmt := `INSERT INTO trade (
		cbt_id,
		trb_id,
		bii_id,
		tgr_id,
		biv_market_date,
		trd_qty,
		cbt_avg_price,
		trd_avg_price,
		trd_raw_results,
		trd_raw_price,
		trd_total_tax
	) VALUES (?,?,?,?,?,?,?,?,?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	trade := dao.trade

	res, err := stmt.Exec(
		trade.CompanyBatch.Id,
		trade.TradeBatch.Id,
		trade.Item.Id,
		trade.TaxGroup.Id,
		trade.MarketDate,
		trade.Qty,
		trade.CompanyBatch.AvgPrice,
		trade.AvgPrice,
		trade.RawResults,
		trade.RawPrice,
		trade.TotalTax,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	tradeRec := &entity.Trade{
		Id:           lastId,
		TaxGroup:     trade.TaxGroup,
		Item:         trade.Item,
		TradeBatch:   trade.TradeBatch,
		CompanyBatch: trade.CompanyBatch,
		MarketDate:   trade.MarketDate,
		Qty:          trade.Qty,
		AvgPrice:     trade.AvgPrice,
		RawResults:   trade.RawResults,
		RawPrice:     trade.RawPrice,
		TotalTax:     trade.TotalTax,
	}

	log.Printf(
		"TradeDAO.CreateTrade: created new trade record [%d, %s, q = %d, res = %f]",
		tradeRec.Id,
		tradeRec.Item.Company.Code,
		tradeRec.Qty,
		tradeRec.RawResults,
	)

	return tradeRec, nil
}
