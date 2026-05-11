package currency

import "fmt"

type Currency string

const (
	USD Currency = "USD"
	CNY Currency = "CNY"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
	JPY Currency = "JPY"
	BTC Currency = "BTC"
)

var currencyDecimals = map[Currency]int{
	USD: 2,
	CNY: 2,
	EUR: 2,
	GBP: 2,
	JPY: 0,
	BTC: 8,
}

func GetDecimals(currency Currency) (int, error) {
	if d, ok := currencyDecimals[currency]; ok {
		return d, nil
	}
	return 0, fmt.Errorf("unknown currency: %s", currency)
}

func RegisterCurrency(currency Currency, decimals int) {
	currencyDecimals[currency] = decimals
}

type Rate struct {
	From Currency `json:"from"`
	To   Currency `json:"to"`
	Rate float64  `json:"rate"`
}
