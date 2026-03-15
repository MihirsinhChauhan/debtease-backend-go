package health_v1

import "github.com/shopspring/decimal"

var decimalZero = decimal.Zero

// ratio returns num/denom; returns decimalZero if denom is zero (safe division).
func ratio(num, denom decimal.Decimal) decimal.Decimal {
	if denom.IsZero() {
		return decimalZero
	}
	return num.Div(denom)
}
