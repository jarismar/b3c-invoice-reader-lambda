package db

import (
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
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

func FindUserByExternalUUID(conn *sql.Tx, uuid string) (*entity.User, error) {
	stmt, err := conn.Prepare(qryUserByExternalUUID)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var user entity.User

	qryErr := stmt.QueryRow(uuid).Scan(
		&user.UserID,
		&user.UserUUID,
		&user.UserExternalUUID,
		&user.UserName,
	)

	if qryErr == sql.ErrNoRows {
		return nil, nil
	}

	return &user, nil
}

func CreateUser(conn *sql.Tx, cli *inputData.Client) (*entity.User, error) {
	stmt, err := conn.Prepare(insertUserStmt)
	if err != nil {
		return nil, err
	}

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
		UserID:           lastId,
		UserUUID:         userUUID,
		UserExternalUUID: cli.Id,
		UserName:         cli.Name,
	}

	log.Printf("created new user (%s) from %s %s \n", userUUID.String(), cli.Id, cli.Name)

	return &user, nil
}

func UpdateUser(conn *sql.Tx, usr *entity.User) error {
	stmt, err := conn.Prepare(updateUserStmt)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(usr.UserName, usr.UserID)
	if err != nil {
		return err
	}

	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowCnt != 1 {
		return errors.New("userDAO::UpdateUser - too many affected rows")
	}

	return nil
}
