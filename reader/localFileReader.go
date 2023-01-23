package reader

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jarismar/b3c-invoice-reader-lambda/input"
)

func LocalFileReader(fileName string) (*input.Invoice, error) {
	baseName := filepath.Base(fileName)
	filenamePattern := regexp.MustCompile(`^\d{4}_\d{2}_\d{2}_\d{9}\.json$`)

	if !filenamePattern.MatchString(baseName) {
		log.Printf("reader.LocalFileReader: invalid file name: %s", baseName)
		err := fmt.Errorf("invalid file name: %s", baseName)
		return nil, err
	}

	invoiceFile, err := os.Open(fileName)

	if err != nil {
		log.Printf("reader.LocalFileReader: error opening file: %s", fileName)
		return nil, err
	}

	defer invoiceFile.Close()

	jsonContent, err := io.ReadAll(invoiceFile)

	if err != nil {
		log.Printf("reader.LocalFileReader: error reading file: %s", fileName)
		return nil, err
	}

	var invoice input.Invoice

	json.Unmarshal(jsonContent, &invoice)

	log.Printf("reader.LocalFileReader: success loading: %s", fileName)

	return &invoice, nil
}
