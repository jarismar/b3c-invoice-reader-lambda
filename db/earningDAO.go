package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

const qryEarningByUUID = `SELECT
  ear_id,
  ear_type,
  ear_pay_date,
  ear_raw_value,
  ear_net_value
FROM earning
WHERE ear_uuid = ?`

const insertEarningStmt = `INSERT INTO earning (
  ear_uuid,
  usr_id,
  cmp_id,
  ear_type,
  ear_pay_date,
  ear_raw_value,
  ear_net_value
) VALUES (?,?,?,?,?,?,?)`

type EarningDAO struct {
	conn    *sql.Tx
	user    *entity.User
	company *entity.Company
}

func GetEarningDAO(conn *sql.Tx, user *entity.User, company *entity.Company) *EarningDAO {
	return &EarningDAO{
		conn:    conn,
		user:    user,
		company: company,
	}
}

func (dao *EarningDAO) FindByUUID(uuid uuid.UUID) (*entity.Earning, error) {
	stmt, err := dao.conn.Prepare(qryEarningByUUID)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var earning entity.Earning

	err = stmt.QueryRow(uuid).Scan(
		&earning.Id,
		&earning.Type,
		&earning.PayDate,
		&earning.RawValue,
		&earning.NetValue,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	earning.UUID = uuid
	earning.Company = dao.company
	earning.User = dao.user

	return &earning, nil
}

func (dao *EarningDAO) CreateEarning(earningInput *inputData.EarningEntry) (*entity.Earning, error) {
	stmt, err := dao.conn.Prepare(insertEarningStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	payDate, err := time.Parse(time.RFC3339, earningInput.PayDate)
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(
		earningInput.UUID,
		dao.user.Id,
		dao.company.Id,
		earningInput.Type,
		payDate,
		earningInput.RawValue,
		earningInput.NetValue,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	earning := entity.Earning{
		Id:       lastId,
		UUID:     earningInput.UUID,
		Type:     earningInput.Type,
		Company:  dao.company,
		User:     dao.user,
		PayDate:  payDate,
		RawValue: earningInput.RawValue,
		NetValue: earningInput.NetValue,
	}

	log.Printf("created earning record [%d, %s, %f]\n", earning.Id, earning.Company.Code, earning.NetValue)

	return &earning, nil
}
