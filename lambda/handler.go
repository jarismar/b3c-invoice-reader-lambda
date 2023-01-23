package lambda

import (
	"context"
	"fmt"
	"log"

	"github.com/jarismar/b3c-invoice-reader-lambda/reader"
)

type Request struct {
	Filename string `json:"filename"`
}

func validateRequest(req *Request) error {
	if req.Filename == "" {
		err := fmt.Errorf("Lambda.Handler: Invalid request: Invalid filename")
		return err
	}

	return nil
}

func Handler(ctx context.Context, req Request) (bool, error) {
	if err := validateRequest(&req); err != nil {
		log.Fatal(err.Error())
		return false, err
	}

	log.Printf("lambda.Handler: Handling file %s", req.Filename)

	invoiceInput, err := reader.S3FileReader(req.Filename)
	if err != nil {
		return false, err
	}

	log.Printf("lambda.Handler: Skipping file: %s", invoiceInput.FileName)

	return false, nil
}
