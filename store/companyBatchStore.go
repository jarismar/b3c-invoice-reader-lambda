package store

import (
	"log"

	"github.com/jarismar/b3c-service-entities/entity"
)

type CompanyBatchStore struct {
	cache map[string]*entity.CompanyBatch
}

func GetCompanyBatchStore() *CompanyBatchStore {
	return &CompanyBatchStore{
		cache: make(map[string]*entity.CompanyBatch),
	}
}

func (store *CompanyBatchStore) Has(cb *entity.CompanyBatch) bool {
	_, ok := store.cache[cb.Company.Code]
	return ok
}

func (store *CompanyBatchStore) Put(cb *entity.CompanyBatch) *entity.CompanyBatch {
	store.cache[cb.Company.Code] = cb
	return cb
}

func (store *CompanyBatchStore) Get(cb *entity.CompanyBatch) *entity.CompanyBatch {
	cbrec, ok := store.cache[cb.Company.Code]

	if !ok {
		log.Printf("store.CompanyBatchStore.Get: WARNING: entry %s not found", cb.Company.Code)
	}

	return cbrec
}
