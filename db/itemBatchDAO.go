package db

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type ItemBatchDAO struct {
	tx        *sql.Tx
	itemBatch *entity.ItemBatch
}

func GetItemBatchDAO(tx *sql.Tx, itemBatch *entity.ItemBatch) *ItemBatchDAO {
	return &ItemBatchDAO{
		tx:        tx,
		itemBatch: itemBatch,
	}
}

func (dao *ItemBatchDAO) CreateItemBatch() (*entity.ItemBatch, error) {
	insertStmt := `INSERT into item_batch (
		cbt_id,
		bii_id,
		tgr_id,
		itb_qty,
		itb_avg_price,
		itb_raw_price,
		itb_total_tax
	) VALUES (?,?,?,?,?,?,?)`

	stmt, err := dao.tx.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	itemBatch := dao.itemBatch

	res, err := stmt.Exec(
		itemBatch.CompanyBatch.Id,
		itemBatch.Item.Id,
		itemBatch.TaxGroup.Id,
		itemBatch.Qty,
		itemBatch.AvgPrice,
		itemBatch.RawPrice,
		itemBatch.TotalTaxes,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	itemBatchRec := &entity.ItemBatch{
		Id:           lastId,
		Item:         itemBatch.Item,
		TaxGroup:     itemBatch.TaxGroup,
		CompanyBatch: itemBatch.CompanyBatch,
		Qty:          itemBatch.Qty,
		AvgPrice:     itemBatch.AvgPrice,
		RawPrice:     itemBatch.RawPrice,
		TotalTaxes:   itemBatch.TotalTaxes,
	}

	log.Printf(
		"ItemBatchDAO.CreateItemBatch: created invoice item batch [%d, %s, qty = %d, tax = %.4f, raw = %.4f, avg = %.4f]",
		itemBatchRec.Id,
		itemBatchRec.CompanyBatch.Company.Code,
		itemBatchRec.Qty,
		itemBatchRec.TotalTaxes,
		itemBatchRec.RawPrice,
		itemBatchRec.AvgPrice,
	)

	return itemBatchRec, nil
}
