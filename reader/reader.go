package reader

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/jarismar/b3c-invoice-reader-lambda/inputData"
)

type FileMeta struct {
	FilePath string
	FileType string
}

func ReadAssets(path string) ([]FileMeta, error) {
	fileEntries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	fileList := make([]FileMeta, 0, len(fileEntries))

	for _, fileDesc := range fileEntries {
		if fileDesc.IsDir() {
			continue
		}

		filename := fileDesc.Name()
		if len(filename) < 21 {
			continue
		}

		var fileMetaEntry FileMeta

		fileMetaEntry.FilePath = (path + filename)

		filePrefix := filename[:8]
		if filePrefix == "earnings" {
			fileMetaEntry.FileType = "ear"
		} else {
			fileMetaEntry.FileType = "nc"
		}

		fileList = append(fileList, fileMetaEntry)
	}

	return fileList, nil
}

func ReadInvoice(fileName string) (*inputData.Invoice, error) {
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

func ReadEarnings(fileName string) (*inputData.Earning, error) {
	earningsFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer earningsFile.Close()

	jsonContent, err := ioutil.ReadAll(earningsFile)
	if err != nil {
		return nil, err
	}

	var earnings inputData.Earning

	json.Unmarshal(jsonContent, &earnings)

	return &earnings, nil
}
