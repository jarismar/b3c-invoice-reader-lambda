package service

import (
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-service-entities/entity"
)

type UserService struct {
	userDAO *db.UserDAO
}

func GetUserService(userDAO *db.UserDAO) *UserService {
	return &UserService{
		userDAO: userDAO,
	}
}

func (userService *UserService) insert(cli *inputData.Client) (*entity.User, error) {
	log.Printf("creating user [%s, %s]\n", cli.Id, cli.Name)
	return userService.userDAO.CreateUser(cli)
}

func (userService *UserService) update(user *entity.User) (*entity.User, error) {
	log.Printf("updating user [%s, %s]\n", user.ExternalUUID, user.UserName)
	err := userService.userDAO.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (userService *UserService) UpsertUser(cli *inputData.Client) (*entity.User, error) {
	user, err := userService.userDAO.FindByExternalUUID(cli.Id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return userService.insert(cli)
	}

	if user.UserName != cli.Name {
		user.UserName = cli.Name
		return userService.update(user)
	}

	log.Printf("nothing to be done for user [%s, %s]\n", cli.Id, cli.Name)
	return user, nil
}
