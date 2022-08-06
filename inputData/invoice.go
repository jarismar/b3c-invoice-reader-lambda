package inputData

type Invoice struct {
	Market      string    `json:"market"`
	InvoiceNum  int64     `json:"invoiceNum"`
	FileName    string    `json:"filename"`
	MarketDate  string    `json:"marketDate"`
	BillingDate string    `json:"billingDate"`
	AgentId     string    `json:"agentId"`
	RawValue    float64   `json:"rawValue"`
	NetValue    float64   `json:"netValue"`
	Client      Client    `json:"client"`
	Items       []Item    `json:"items"`
	Taxes       []Tax     `json:"taxes"`
	Earnings    []Earning `json:"earnings,omitempty"`
}
