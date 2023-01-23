package store

import (
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type TaxStore struct {
	cache map[string]*entity.Tax
}

func GetTaxStore() *TaxStore {
	return &TaxStore{
		cache: make(map[string]*entity.Tax),
	}
}

func (store *TaxStore) Has(tax *entity.Tax) bool {
	_, ok := store.cache[tax.Code]
	return ok
}

func (store *TaxStore) Put(tax *entity.Tax) *entity.Tax {
	store.cache[tax.Code] = tax
	return tax
}

func (store *TaxStore) Get(tax *entity.Tax) *entity.Tax {
	entry, ok := store.cache[tax.Code]

	if !ok {
		log.Printf("store.TaxStore.Get: WARNING: entry %s not found", tax.Code)
	}

	return entry
}
