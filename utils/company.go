package utils

import (
	"strings"

	"github.com/jarismar/b3c-service-entities/entity"
)

func IsBDR(cmp *entity.Company) bool {
	return strings.HasSuffix(cmp.Code, "34") || strings.HasSuffix(cmp.Name, "DRN")
}
