# b3c-invoice-reader-lambda

Read invoice, validate and populate database with all entities from the invoice.

## AWS Environment

Files are read from this S3 Bucket `b3c-data/invoice/<userID>/<yyyy_mm>/`
Files are set to auto-archive in 90s after they are created

## Dependencies

1. [google/uuid](https://github.com/google/uuid)
2. [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
3. [jarismar/b3c-service-entities](https://github.com/jarismar/b3c-service-entities)

## References

1. [AWS Golang SDK v2](https://github.com/aws/aws-sdk-go-v2)
2. [AWS Lambda - Go function handler](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html)
3. [AWS Lambda - Database Proxy](https://docs.aws.amazon.com/lambda/latest/dg/configuration-database.html)
4. [How To Write Unit Tests in Go](https://www.digitalocean.com/community/tutorials/how-to-write-unit-tests-in-go-using-go-test-and-the-testing-package)
