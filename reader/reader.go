package reader

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
)

func Read(fileName string) (*inputData.Invoice, error) {
	invoiceFile, err := os.Open(fileName)

	if err != nil {
		return nil, err
	}

	defer invoiceFile.Close()

	jsonContent, err := ioutil.ReadAll(invoiceFile)

	if err != nil {
		return nil, err
	}

	var invoice inputData.Invoice

	json.Unmarshal(jsonContent, &invoice)

	return &invoice, nil
}
