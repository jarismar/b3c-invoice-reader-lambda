# b3c-invoice-reader-lambda

Read invoice, validate and populate database with all entities from the invoice.

## Environment

Files are read from this S3 Bucket `b3c-data/invoice`

## Dependencies

1. [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

## References

1. [AWS Golang SDK v2](https://github.com/aws/aws-sdk-go-v2)
