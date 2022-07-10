package utils

import "github.com/jarismar/b3c-service-entities/entity"

func FindItemByInvoiceAndId(invoice *entity.Invoice, id int64) *entity.InvoiceItem {
	for _, item := range invoice.Items {
		if item.Id == id {
			return &item
		}
	}

	return nil
}
