package local

import (
	"fmt"
	"log"
	"os"

	"github.com/jarismar/b3c-invoice-reader-lambda/db"
	"github.com/jarismar/b3c-invoice-reader-lambda/reader"
	"github.com/jarismar/b3c-invoice-reader-lambda/report"
	"github.com/jarismar/b3c-invoice-reader-lambda/service"
	"github.com/jarismar/b3c-invoice-reader-lambda/store"
)

func Handler() (bool, error) {
	if len(os.Args) != 2 {
		err := fmt.Errorf("local.Handler: error: missing filename on arg1 <invoice>")
		return false, err
	}

	fileNameStr := os.Args[1]
	log.Printf("local.Hander: processing file %s", fileNameStr)

	invoiceInput, err := reader.LocalFileReader(fileNameStr)
	if err != nil {
		return false, err
	}

	log.Print("local.Handler: going to process invoice: ", invoiceInput.FileName)

	conn, err := db.GetConnection()
	if err != nil {
		return false, err
	}

	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		return false, err
	}

	invoiceService := service.GetInvoiceService(
		tx,
		invoiceInput,
		store.GetTaxStore(),
		store.GetCompanyStore(),
		store.GetCompanyBatchStore(),
		store.GetBrokerTaxStore(),
	)

	invoiceRec, err := invoiceService.ProcessInvoice()

	// TODO self checking

	report := report.GetConsoleReport(invoiceRec)
	report.Run()

	if err != nil {
		tx.Rollback()
		return false, err
	}

	tx.Commit()
	// tx.Rollback()

	log.Printf("local.Handler: done processing file: %s", fileNameStr)

	return true, nil
}
