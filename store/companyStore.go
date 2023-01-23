package store

import (
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type CompanyStore struct {
	cache map[string]*entity.Company
}

func GetCompanyStore() *CompanyStore {
	return &CompanyStore{
		cache: make(map[string]*entity.Company),
	}
}

func (store *CompanyStore) Has(company *entity.Company) bool {
	_, ok := store.cache[company.Code]
	return ok
}

func (store *CompanyStore) Put(company *entity.Company) *entity.Company {
	store.cache[company.Code] = company
	return company
}

func (store *CompanyStore) Get(company *entity.Company) *entity.Company {
	entry, ok := store.cache[company.Code]

	if !ok {
		log.Printf("store.CompanyStore.Get: WARNING: entry %s not found", company.Code)
	}

	return entry
}
