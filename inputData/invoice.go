package inputData

type Invoice struct {
	Market      string  `json:"market"`
	InvoiceNum  int     `json:"inviceNum"`
	MarketDate  string  `json:"marketDate"`
	BillingDate string  `json:"billingDate"`
	AgentId     string  `json:"agentId"`
	RawValue    float32 `json:"rawValue"`
	NetValue    float32 `json:"netValue"`
	Client      Client  `json:"client"`
	Items       []Item  `json:"items"`
	Taxes       []Tax   `json:"taxes"`
}
