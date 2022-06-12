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

func getCompanyByCode(companyCode string, companies []entity.Company) *entity.Company {
	for _, company := range companies {
		if company.Code == companyCode {
			return &company
		}
	}

	return nil
}

func getTaxByCode(taxCode string, taxes []entity.Tax) *entity.Tax {
	for _, tax := range taxes {
		if tax.Code == taxCode {
			return &tax
		}
	}

	return nil
}

func processClient(tx *sql.Tx, client *inputData.Client) (*entity.User, error) {
	userService := service.GetUserService(db.GetUserDAO(tx))
	return userService.UpsertUser(client)
}

func processCompanies(tx *sql.Tx, items []inputData.Item) ([]entity.Company, error) {
	companyService := service.GetCompanyService(tx)
	companies := make([]entity.Company, 0, len(items))

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
	taxService := service.GetTaxService(tx)
	taxes := make([]entity.Tax, 0, len(inputTaxes))

	for _, inputTax := range inputTaxes {
		tax, err := taxService.UpsertTax(&inputTax)
		if err != nil {
			return nil, err
		}
		taxes = append(taxes, *tax)
	}

	return taxes, nil
}

func processInvoice(tx *sql.Tx, user *entity.User, invoiceInput *inputData.Invoice) (*entity.Invoice, error) {
	invoiceService := service.GetInvoiceService(tx, user)
	return invoiceService.UpsertInvoice(invoiceInput)
}

func processInvoiceItems(
	tx *sql.Tx,
	invoice *entity.Invoice,
	companies []entity.Company,
	invoiceItemInputs []inputData.Item,
) error {
	invoice.Items = make([]entity.InvoiceItem, 0, len(invoiceItemInputs))
	for _, invoiceItemInput := range invoiceItemInputs {
		companyCode := invoiceItemInput.Company.Code
		company := getCompanyByCode(companyCode, companies)
		if company == nil {
			return fmt.Errorf("invalid company code %s", companyCode)
		}
		service := service.GetInvoiceItemService(tx, invoice, company)
		invoiceItem, err := service.UpsertInvoiceItem(&invoiceItemInput)
		if err != nil {
			return err
		}
		invoice.Items = append(invoice.Items, *invoiceItem)
	}

	return nil
}

func processInvoiceTaxes(
	tx *sql.Tx,
	invoice *entity.Invoice,
	taxes []entity.Tax,
	inputTaxes []inputData.Tax,
) error {
	invoice.Taxes = make([]entity.InvoiceTax, 0, len(taxes))

	for _, invoiceTaxInput := range inputTaxes {
		tax := getTaxByCode(invoiceTaxInput.Code, taxes)
		if tax == nil {
			return fmt.Errorf("invalid tax code %s", invoiceTaxInput.Code)
		}
		invoiceTaxService := service.GetInvoiceTaxService(tx, invoice, tax)
		invoiceTax, err := invoiceTaxService.UpsertInvoiceTax(&invoiceTaxInput)
		if err != nil {
			return err
		}
		invoice.Taxes = append(invoice.Taxes, *invoiceTax)
	}

	return nil
}

func main() {
	invoiceInput, err := reader.Read("./assets/2022_04_28_000125974.json")
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

	user, err := processClient(tx, &invoiceInput.Client)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	companies, err := processCompanies(tx, invoiceInput.Items[0:])
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	taxes, err := processTaxes(tx, invoiceInput.Taxes[0:])
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	invoice, err := processInvoice(tx, user, invoiceInput)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	err = processInvoiceItems(tx, invoice, companies, invoiceInput.Items[0:])
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	err = processInvoiceTaxes(tx, invoice, taxes, invoiceInput.Taxes)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return
	}

	tx.Commit()

	fmt.Println("---------- Summary ----------")

	invoiceItems := invoice.Items
	invoiceTaxes := invoice.Taxes

	fmt.Printf("Invoice.file  : %s\n", invoice.FileName)
	fmt.Printf("Invoice.Id .. : %d\n", invoice.Id)
	fmt.Printf("User.uuid ... : %s\n", user.UUID.String())
	fmt.Printf("User.id ..... : %d\n", user.Id)
	fmt.Printf("User.name ... : %s\n", user.UserName)
	fmt.Printf("Companies ... : %2d %-6s %-20s %3s %s\n", len(invoiceItems), "Code", "Name", "Qty", "Debit")
	for i, invoiceItem := range invoiceItems {
		company := invoiceItem.Company
		fmt.Printf("%14s: %2d %6s %-20s %3d %5t\n", "", i+1, company.Code, company.Name, invoiceItem.Qty, invoiceItem.Debit)
	}
	fmt.Printf("Taxes ....... : %d\n", len(taxes))
	for i, tax := range invoiceTaxes {
		fmt.Printf("%14s: %d %10s %7.2f\n", "", i+1, tax.Tax.Code, tax.Value)
	}
}
