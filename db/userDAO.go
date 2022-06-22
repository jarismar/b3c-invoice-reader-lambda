package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

const qryUserByExternalUUID = `SELECT
  usr_id,
  usr_uuid,
  usr_ext_uuid,
  usr_name
FROM user
WHERE usr_ext_uuid = ?`

const insertUserStmt = "INSERT INTO user (usr_uuid, usr_ext_uuid, usr_name) VALUES (?,?,?)"

const updateUserStmt = "UPDATE user SET usr_name = ? WHERE usr_id = ?"

type UserDAO struct {
	conn *sql.Tx
}

func GetUserDAO(conn *sql.Tx) *UserDAO {
	return &UserDAO{
		conn: conn,
	}
}

func (dao *UserDAO) FindByExternalUUID(uuid string) (*entity.User, error) {
	stmt, err := dao.conn.Prepare(qryUserByExternalUUID)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var user entity.User

	err = stmt.QueryRow(uuid).Scan(
		&user.Id,
		&user.UUID,
		&user.ExternalUUID,
		&user.UserName,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (dao *UserDAO) CreateUser(cli *inputData.Client) (*entity.User, error) {
	stmt, err := dao.conn.Prepare(insertUserStmt)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	userUUID := uuid.New()
	res, err := stmt.Exec(userUUID.String(), cli.Id, cli.Name)
	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	user := entity.User{
		Id:           lastId,
		UUID:         userUUID,
		ExternalUUID: cli.Id,
		UserName:     cli.Name,
	}

	log.Printf("created user [%d, %s, %s, %s]\n", user.Id, userUUID.String(), cli.Id, cli.Name)

	return &user, nil
}

func (dao *UserDAO) UpdateUser(user *entity.User) error {
	stmt, err := dao.conn.Prepare(updateUserStmt)
	if err != nil {
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(user.UserName, user.Id)
	if err != nil {
		return err
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowCnt != 1 {
		details := fmt.Sprintf("expected 1 found %d rows", rowCnt)
		return utils.GetError("UserDAO::UpdateUser", "ERR_DB_001", details)
	}

	log.Printf("updated user [%d, %s, %s] \n", user.Id, user.ExternalUUID, user.UserName)

	return nil
}
