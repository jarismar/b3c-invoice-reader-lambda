package db

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/jarismar/b3c-service-entities/entity"
)

type UserDAO struct {
	conn *sql.Tx
	user *entity.User
}

func GetUserDAO(conn *sql.Tx, user *entity.User) *UserDAO {
	return &UserDAO{
		conn: conn,
		user: user,
	}
}

func (dao *UserDAO) GetUser() (*entity.User, error) {
	query := `SELECT
		usr_id,
		usr_uuid,
		usr_ext_uuid,
		usr_name
	FROM user
	WHERE usr_ext_uuid = ?`

	stmt, err := dao.conn.Prepare(query)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	var user entity.User

	err = stmt.QueryRow(dao.user.ExternalUUID).Scan(
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

func (dao *UserDAO) CreateUser() (*entity.User, error) {
	insertStmt := `INSERT INTO user (
		usr_uuid,
		usr_ext_uuid,
		usr_name
	) VALUES (?,?,?)`

	stmt, err := dao.conn.Prepare(insertStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	userUUID := uuid.New()
	user := dao.user

	res, err := stmt.Exec(
		userUUID.String(),
		user.ExternalUUID,
		user.UserName,
	)

	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	userRec := entity.User{
		Id:           lastId,
		UUID:         userUUID,
		ExternalUUID: user.ExternalUUID,
		UserName:     user.UserName,
	}

	log.Printf("userDAO.CreateUser: created user [%d, %s]", userRec.Id, userRec.UserName)

	return &userRec, nil
}
