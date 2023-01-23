package store

import "github.com/jarismar/b3c-service-entities/entity"

type BrokerTaxStore struct {
	cache map[string]bool
}

func GetBrokerTaxStore() *BrokerTaxStore {
	return &BrokerTaxStore{
		cache: make(map[string]bool),
	}
}

func (store *BrokerTaxStore) getItemId(item *entity.InvoiceItem, taxCode string) string {
	var prefix string

	if item.Debit {
		prefix = "D-"
	} else {
		prefix = "C-"
	}

	return prefix + taxCode + "-" + item.Company.Code
}

func (store *BrokerTaxStore) Has(item *entity.InvoiceItem, taxCode string) bool {
	key := store.getItemId(item, taxCode)
	_, ok := store.cache[key]
	return ok
}

func (store *BrokerTaxStore) Put(item *entity.InvoiceItem, taxCode string) {
	key := store.getItemId(item, taxCode)
	store.cache[key] = true
}
