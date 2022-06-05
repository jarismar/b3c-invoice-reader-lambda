package main

import (
	"fmt"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/reader"
	"github.com/jarismar/b3c-invoice-reader-lambda/service"
)

func main() {
	invoice, err := reader.Read("./assets/2022_04_28_000125974.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := db.GetConnection()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}

	user, err := service.UpsertUser(tx, &invoice.Client)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	tx.Commit()

	fmt.Println("User.uuid ... : ", user.UserUUID.String())
	fmt.Println("User.id ..... : ", user.UserID)
	fmt.Println("User.name ... : ", user.UserName)
}
