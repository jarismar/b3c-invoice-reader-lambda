package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
	"github.com/jarismar/b3c-invoice-reader-lambda/reader"
	"github.com/jarismar/b3c-invoice-reader-lambda/service"
	"github.com/jarismar/b3c-service-entities/entity"
)

// TODO move to utils / company
func getCompanyByCode(companyCode string, companies []entity.Company) *entity.Company {
	for _, company := range companies {
		if company.Code == companyCode {
			return &company
		}
	}

	return nil
}

// TODO move to utils / tax
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

func getInputCompaniesFromInvoice(items []inputData.Item) []inputData.Company {
	inputCompanies := make([]inputData.Company, 0, len(items))
	for _, item := range items {
		inputCompanies = append(inputCompanies, item.Company)
	}
	return inputCompanies
}

func getInputCompaniesFromEarnings(earnings []inputData.EarningEntry) []inputData.Company {
	inputCompanies := make([]inputData.Company, 0, len(earnings))
	for _, earning := range earnings {
		inputCompanies = append(inputCompanies, earning.Company)
	}
	return inputCompanies
}

func processCompanies(tx *sql.Tx, inputCompanies []inputData.Company) ([]entity.Company, error) {
	companyService := service.GetCompanyService(tx)
	companies := make([]entity.Company, 0, len(inputCompanies))

	for _, inputCompany := range inputCompanies {
		company, err := companyService.UpsertCompany(&inputCompany)
		if err != nil {
			return nil, err
		}
		companies = append(companies, *company)
	}

	return companies, nil
}

func getInputTaxesFromEarnings(earnings []inputData.EarningEntry) []inputData.Tax {
	foundTaxes := make(map[string]bool)
	inputTaxes := make([]inputData.Tax, 0, len(earnings))

	for _, earning := range earnings {
		for _, tax := range earning.Taxes {
			if _, ok := foundTaxes[tax.Code]; ok {
				continue
			}

			inputTaxes = append(inputTaxes, tax)
			foundTaxes[tax.Code] = true
		}
	}

	return inputTaxes
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

func processEarnings(
	tx *sql.Tx,
	inputEarnings []inputData.EarningEntry,
	user *entity.User,
	taxes []entity.Tax,
	companies []entity.Company,
) ([]entity.Earning, error) {
	earnings := make([]entity.Earning, 0, len(inputEarnings))

	for _, inputEarning := range inputEarnings {
		company := getCompanyByCode(inputEarning.Company.Code, companies)
		earningService := service.GetEarningService(tx, user, company)
		earning, err := earningService.UpsertEarning(&inputEarning)
		if err != nil {
			return nil, err
		}

		earningTaxes := make([]entity.EarningTax, 0, len(inputEarning.Taxes))
		for _, inputTax := range inputEarning.Taxes {
			tax := getTaxByCode(inputTax.Code, taxes)
			earningTaxService := service.GetEarningTaxService(tx, earning, tax)
			earningTax, err := earningTaxService.UpsertEarningTax(&inputTax)
			if err != nil {
				return nil, err
			}
			earningTaxes = append(earningTaxes, *earningTax)
		}

		earning.Taxes = earningTaxes
		earnings = append(earnings, *earning)
	}

	return earnings, nil
}

func processSales(
	tx *sql.Tx,
	invoice *entity.Invoice,
) ([]entity.Trade, error) {
	trades := make([]entity.Trade, 0, len(invoice.Items))
	for _, item := range invoice.Items {
		if !item.Debit {
			saleService := service.GetSaleService(tx, invoice, &item)
			trade, err := saleService.UpsertSale()

			if err != nil {
				return nil, err
			}

			trades = append(trades, *trade)
		}
	}

	return trades, nil
}

func processInvoiceFile(filePath string) error {
	invoiceInput, err := reader.ReadInvoice(filePath)
	if err != nil {
		return err
	}

	conn, err := db.GetConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	user, err := processClient(tx, &invoiceInput.Client)
	if err != nil {
		tx.Rollback()
		return err
	}

	inputCompanies := getInputCompaniesFromInvoice(invoiceInput.Items)
	companies, err := processCompanies(tx, inputCompanies)
	if err != nil {
		tx.Rollback()
		return err
	}

	taxes, err := processTaxes(tx, invoiceInput.Taxes[0:])
	if err != nil {
		tx.Rollback()
		return err
	}

	invoice, err := processInvoice(tx, user, invoiceInput)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = processInvoiceItems(tx, invoice, companies, invoiceInput.Items[0:])
	if err != nil {
		tx.Rollback()
		return err
	}

	err = processInvoiceTaxes(tx, invoice, taxes, invoiceInput.Taxes)
	if err != nil {
		tx.Rollback()
		return err
	}

	trades, err := processSales(tx, invoice)
	if err != nil {
		tx.Rollback()
		return err
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
	fmt.Printf("Companies ... : %-2d %-6s %-20s %3s %s\n", len(invoiceItems), "Code", "Name", "Qty", "Debit")
	for i, invoiceItem := range invoiceItems {
		company := invoiceItem.Company
		fmt.Printf("%14s: %-2d %-6s %-20s %3d %5t\n", "", i+1, company.Code, company.Name, invoiceItem.Qty, invoiceItem.Debit)
	}
	fmt.Printf("Taxes ....... : %-2d\n", len(taxes))
	for i, tax := range invoiceTaxes {
		fmt.Printf("%14s: %-2d %-10s %7.2f\n", "", i+1, tax.Tax.Code, tax.Value)
	}
	fmt.Printf("Trades ...... : %-2d %-6s %-20s %3s\n", len(trades), "Code", "Name", "Qty")
	for i, trade := range trades {
		itemOut := trade.InvoiceItemOut
		fmt.Printf("%14s: %-2d %-6s %-20s %3d\n", "", i+1, itemOut.Company.Code, itemOut.Company.Name, itemOut.Qty)
	}

	return nil
}

func processEarningsFile(filePath string) error {
	inputEarningsRecord, err := reader.ReadEarnings(filePath)
	if err != nil {
		return err
	}

	conn, err := db.GetConnection()
	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	inputEarnings := inputEarningsRecord.Items

	user, err := processClient(tx, &inputEarningsRecord.Client)
	if err != nil {
		tx.Rollback()
		return err
	}

	inputCompanies := getInputCompaniesFromEarnings(inputEarnings)
	companies, err := processCompanies(tx, inputCompanies)
	if err != nil {
		tx.Rollback()
		return err
	}

	inputTaxes := getInputTaxesFromEarnings(inputEarnings)
	taxes, err := processTaxes(tx, inputTaxes)
	if err != nil {
		tx.Rollback()
		return err
	}

	earnings, err := processEarnings(tx, inputEarnings, user, taxes, companies)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	fmt.Println("---------- Summary ----------")
	fmt.Printf("Earnings.file  : %s\n", filePath)
	fmt.Printf("User.uuid ... : %s\n", user.UUID.String())
	fmt.Printf("User.id ..... : %d\n", user.Id)
	fmt.Printf("User.name ... : %s\n", user.UserName)

	fmt.Printf("Earnings .... : %2d %-6s %7s %7s\n", len(earnings), "Code", "Raw", "Net")
	for i, earning := range earnings {
		fmt.Printf("%14s: %2d %-6s %7.2f %7.2f\n", "", i, earning.Company.Code, earning.RawValue, earning.NetValue)
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: program <assetsDir>")
		return
	}

	assetsDir := os.Args[1]
	fmt.Printf("starting on dir %s\n", assetsDir)
	files, err := reader.ReadAssets(assetsDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, fileMeta := range files {
		fmt.Printf("processing %s (%s)\n", fileMeta.FilePath, fileMeta.FileType)
		var err error
		if fileMeta.FileType == "ear" {
			err = processEarningsFile(fileMeta.FilePath)
		} else {
			err = processInvoiceFile(fileMeta.FilePath)
		}

		if err != nil {
			fmt.Println(err)
		}
	}
}
