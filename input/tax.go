package input

type Tax struct {
	Code   string  `json:"code"`
	Source string  `json:"source"`
	Value  float64 `json:"value"`
	Rate   float64 `json:"rate"`
}
