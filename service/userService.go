package service

import (
	"database/sql"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-service-entities/entity"
)

type UserService struct {
	tx   *sql.Tx
	user *entity.User
}

func GetUserService(tx *sql.Tx, user *entity.User) *UserService {
	return &UserService{
		tx:   tx,
		user: user,
	}
}

func (usvc *UserService) LoadUser() (*entity.User, error) {
	userDAO := db.GetUserDAO(usvc.tx, usvc.user)
	return userDAO.GetUser()
}

func (usvc *UserService) UpsertUser() (*entity.User, error) {
	userRec, err := usvc.LoadUser()

	if err != nil {
		return nil, err
	}

	if userRec != nil {
		log.Printf(
			"userService.CreateUser: found user %s on db",
			usvc.user.ExternalUUID,
		)

		return userRec, nil
	}

	userDAO := db.GetUserDAO(usvc.tx, usvc.user)
	return userDAO.CreateUser()
}
