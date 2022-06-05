package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

func insert(conn *sql.Tx, cli *inputData.Client) (*entity.User, error) {
	log.Printf("creating new user %s - %s\n", cli.Id, cli.Name)
	return db.CreateUser(conn, cli)
}

func update(conn *sql.Tx, user *entity.User) (*entity.User, error) {
	log.Printf("updating user %s - %s\n", user.UserExternalUUID, user.UserName)
	err := db.UpdateUser(conn, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func UpsertUser(conn *sql.Tx, cli *inputData.Client) (*entity.User, error) {
	user, err := db.FindUserByExternalUUID(conn, cli.Id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return insert(conn, cli)
	}

	if user.UserName != cli.Name {
		return update(conn, user)
	}

	log.Printf("nothing to be done for user %s - %s\n", cli.Id, cli.Name)
	return user, nil
}
