package reader

import (
	"fmt"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/input"
)

func S3FileReader(fileName string) (*input.Invoice, error) {
	log.Fatal("S3FileReader: error: Not implemented")
	err := fmt.Errorf("S3FileReader: error: Not implemented")
	return nil, err
}
