package db

import (
	"database/sql"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func GetConnection() (*sql.DB, error) {
	timeout, _ := time.ParseDuration("30s")
	config := mysql.Config{
		User:                 os.Getenv("MYSQL_DB_USER"),
		Passwd:               os.Getenv("MYSQL_DB_PASSWD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("MYSQL_DB_ADDR"),
		DBName:               os.Getenv("MYSQL_DB_SCHEMA"),
		Collation:            "utf8_general_ci",
		Timeout:              timeout,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	return sql.Open("mysql", config.FormatDSN())
}
