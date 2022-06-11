package utils

import (
	"fmt"
)

var errorMessagesByCode = map[string]string{
	"ERR_SYS_001": "input data validation error",
	"ERR_DB_001":  "wrong number of affected rows",
}

func GetError(location string, code string, details string) error {
	errorMsg, ok := errorMessagesByCode[code]
	if ok {
		return fmt.Errorf("%s - %s %s %s", location, code, errorMsg, details)
	}

	return fmt.Errorf("%s - %s %s", location, code, details)
}
