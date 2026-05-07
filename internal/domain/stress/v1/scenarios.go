package stress_v1

import (
	"github.com/shopspring/decimal"
)

// ApplyInterestRateShock returns EMI' for each debt after applying
// a simplified interest rate shock, as per v1 policy:
//   - BNPL / 0% EMI: no rate shock (β = 0)
//   - Secured loans: treated as fixed-rate here (β = 0)
//   - Other unsecured loans: mild increase (β = UnsecuredRateShockBeta)
func ApplyInterestRateShock(
	debts []StressDebtEntry,
	betaUnsecured decimal.Decimal,
) []decimal.Decimal {
	shocked := make([]decimal.Decimal, len(debts))
	for i, d := range debts {
		var beta decimal.Decimal
		switch {
		case d.InstrumentType == "bnpl":
			beta = decimalZero
		case d.Secured:
			beta = decimalZero
		default:
			beta = betaUnsecured
		}
		factor := one.Add(beta)
		shocked[i] = d.EMI.Mul(factor)
	}
	return shocked
}

