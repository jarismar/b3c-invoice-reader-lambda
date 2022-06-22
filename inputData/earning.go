package inputData

import (
	"github.com/google/uuid"
	"github.com/jarismar/b3c-service-entities/entity"
)

type EarningEntry struct {
	UUID     uuid.UUID          `json:"uuid"`
	Type     entity.EarningType `json:"type"`
	Company  Company            `json:"company"`
	Taxes    []Tax              `json:"taxes"`
	PayDate  string             `json:"payDate"`
	RawValue float64            `json:"rawValue"`
	NetValue float64            `json:"netValue"`
}

type Earning struct {
	Client Client         `json:"client"`
	Items  []EarningEntry `json:"earnings"`
}
