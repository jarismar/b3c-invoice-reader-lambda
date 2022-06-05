package inputData

type Item struct {
	Company Company `json:"company"`
	Qty     string  `json:"qty"`
	Price   float32 `json:"price"`
	Sell    bool    `json:"sell"`
}
