package api

type Currency string

type ConvertRequest struct {
	Amount float64  `json:"amount"`
	From   Currency `json:"from"`
	To     Currency `json:"to"`
}

type ConvertResponse struct {
	Amount   float64  `json:"amount"`
	From     Currency `json:"from"`
	To       Currency `json:"to"`
	Result   float64  `json:"result"`
	Original float64  `json:"original"`
}

type Rate struct {
	From Currency `json:"from"`
	To   Currency `json:"to"`
	Rate float64  `json:"rate"`
}

type SetRatesRequest struct {
	Rates []Rate `json:"rates"`
}

type SetRatesResponse struct {
	Success bool `json:"success"`
}

type CurrenciesResponse struct {
	Currencies []Currency `json:"currencies"`
}

type RatesResponse struct {
	Rates []Rate `json:"rates"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
