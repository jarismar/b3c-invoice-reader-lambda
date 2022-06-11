package inputData

type Item struct {
	Company Company `json:"company"`
	Qty     int64   `json:"qty"`
	Price   float64 `json:"price"`
	Debit   bool    `json:"debit"`
}
