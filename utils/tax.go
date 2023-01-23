package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jarismar/b3c-invoice-reader-lambda/input"
	"github.com/jarismar/b3c-service-entities/entity"
)

func GetTaxGroupId(t string, prefix string) (int64, error) {
	dt := strings.Split(t, "T")
	ymd := strings.Split(dt[0], "-")
	res := prefix
	for _, ymdv := range ymd {
		res = res + ymdv
	}
	resi, err := strconv.Atoi(res)
	return int64(resi), err
}

func GetTaxGroupIdFromTime(t time.Time, prefix string) (int64, error) {
	ts := t.Format(time.RFC3339)
	return GetTaxGroupId(ts, prefix)
}

func GetInputTaxByCode(taxes []input.Tax, taxCode string) (*input.Tax, error) {
	for _, tax := range taxes {
		if tax.Code == taxCode {
			return &tax, nil
		}
	}

	error := fmt.Errorf("taxService::GetTaxByCode: Unknown tax code %s", taxCode)
	return nil, error
}

func GetTotalTax(taxGroup *entity.TaxGroup) float64 {
	total := 0.0

	for _, tax := range taxGroup.Taxes {
		total = total + tax.TaxValue
	}

	return total
}

func GetTaxValueByGroup(taxGroup *entity.TaxGroup, taxCode string) float64 {
	taxInstance := taxGroup.GetTaxInstanceByCode(taxCode)

	if taxInstance == nil {
		return 0
	} else {
		return taxInstance.TaxValue
	}
}
