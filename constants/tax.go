package constants

type TaxSourcesEnum struct {
	EARNING     string
	INVOICE     string
	ITEM_BATCH  string
	TRADE_BATCH string
	TRADE       string
}

type TaxRatesEnum struct {
	SETFEE        float64
	EMLFEE        float64
	ISSSPFEE      float64 // t = c / 0,95 - c
	IRRFFEE       float64
	IR_EXPT_LIMIT float64
	IRFEE         float64
	BRKFEE        float64
}

type TaxTypesEnum struct {
	SETFEE   string
	EMLFEE   string
	ISSSPFEE string
	IRRFFEE  string
	IRFEE    string
	BRKFEE   string
}

type TaxGroupPrefixEnum struct {
	EARNING     string
	INVOICE     string
	ITEM_BATCH  string
	TRADE       string
	TRADE_BATCH string
}

var TaxSources = TaxSourcesEnum{
	EARNING:     "EAR",
	INVOICE:     "BIV",
	ITEM_BATCH:  "ITB",
	TRADE_BATCH: "TDB",
	TRADE:       "TRD",
}

var TaxTypes = TaxTypesEnum{
	SETFEE:   "SETFEE",
	EMLFEE:   "EMLFEE",
	ISSSPFEE: "ISSSPFEE",
	IRRFFEE:  "IRRFFEE",
	IRFEE:    "IRFEEE",
	BRKFEE:   "BRKFEE",
}

var TaxRates = TaxRatesEnum{
	SETFEE:        0.00025,
	EMLFEE:        0.00005,
	ISSSPFEE:      0.05, // ~ 0.0522449
	IRRFFEE:       0.00005,
	IR_EXPT_LIMIT: 20000,
	IRFEE:         0.15,
	BRKFEE:        4.9,
}

var TaxGroupPrefix = TaxGroupPrefixEnum{
	EARNING:     "1",
	INVOICE:     "2",
	ITEM_BATCH:  "3",
	TRADE:       "4",
	TRADE_BATCH: "5",
}
