package utils

import (
	"fmt"
)

var errorMessagesByCode = map[string]string{
	"ERR_DB_001": "wrong number of affected rows",
}

func GetError(location string, code string) error {
	errorMsg, ok := errorMessagesByCode[code]
	if ok {
		return fmt.Errorf("%s - %s %s", location, code, errorMsg)
	}

	return fmt.Errorf("%s - %s", location, code)
}
