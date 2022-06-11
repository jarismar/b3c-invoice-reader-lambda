package main

import (
	"database/sql"
	"fmt"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-invoice-reader-lambda/reader"
	"github.com/jarismar/b3c-invoice-reader-lambda/service"
	"github.com/jarismar/b3c-service-entities/entity"
)

func processClient(tx *sql.Tx, client *inputData.Client) (*entity.User, error) {
	userService := service.GetUserService(db.GetUserDAO(tx))
	user, err := userService.UpsertUser(client)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func processCompanies(tx *sql.Tx, items []inputData.Item) ([]entity.Company, error) {
	companies := make([]entity.Company, 0, len(items))
	companyService := service.GetCompanyService(tx)

	for _, item := range items {
		company, err := companyService.UpsertCompany(&item.Company)
		if err != nil {
			return nil, err
		}
		companies = append(companies, *company)
	}

	return companies, nil
}

func processTaxes(tx *sql.Tx, inputTaxes []inputData.Tax) ([]entity.Tax, error) {
	taxes := make([]entity.Tax, 0, len(inputTaxes))
	taxService := service.GetTaxService(tx)

	for _, inputTax := range inputTaxes {
		tax, err := taxService.UpsertTax(&inputTax)
		if err != nil {
			return nil, err
		}
		taxes = append(taxes, *tax)
	}

	return taxes, nil
}

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

	user, err := processClient(tx, &invoice.Client)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	companies, err := processCompanies(tx, invoice.Items[0:])
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	taxes, err := processTaxes(tx, invoice.Taxes[0:])
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	tx.Commit()

	fmt.Println("---------- Summary ----------")
	fmt.Printf("Invoice ..... : %s\n", invoice.FileName)
	fmt.Printf("User.uuid ... : %s\n", user.UUID.String())
	fmt.Printf("User.id ..... : %d\n", user.Id)
	fmt.Printf("User.name ... : %s\n", user.UserName)
	fmt.Printf("Companies ... : %d\n", len(companies))
	for _, company := range companies {
		fmt.Printf("%14s: %d, %s, %s\n", "", company.Id, company.Code, company.Name)
	}
	fmt.Printf("Taxes ....... : %d\n", len(taxes))
	for _, tax := range taxes {
		fmt.Printf("%14s: %d, %s, %s\n", "", tax.Id, tax.Code, tax.Source)
	}
}
