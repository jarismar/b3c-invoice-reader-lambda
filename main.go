package main

import (
	"log"
	"os"

	awsLambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/jarismar/b3c-invoice-reader-lambda/lambda"
	"github.com/jarismar/b3c-invoice-reader-lambda/local"
)

func main() {
	log.Print("main: starting")
	envName, envDef := os.LookupEnv("GO_ENV")

	if !envDef {
		log.Fatal("main: GO_ENV not defined")
		return
	}

	if envName == "DEV" {
		log.Print("main: DEV mode is on")
		_, err := local.Handler()
		if err != nil {
			log.Fatal(err)
		}

		log.Print("main: Done!")
		return
	}

	awsLambda.Start(lambda.Handler)
}
