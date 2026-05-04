package stress_v1

import "github.com/shopspring/decimal"

var (
	alphaFull  = decimal.NewFromInt(1)
	alphaShock10 = decimal.RequireFromString("0.9")
	alphaShock20 = decimal.RequireFromString("0.8")
)

// FlexibilityBuffersForIncomeShocks returns monthly flexibility (currency units) at 100%, 90%, and 80%
// of monthly income, using the same definition as ComputeFlexibilityBuffer.
func FlexibilityBuffersForIncomeShocks(
	monthlyIncome, fixedObligations, totalEMI decimal.Decimal,
) (current, shock10, shock20 decimal.Decimal) {
	current, _ = ComputeFlexibilityBuffer(monthlyIncome, fixedObligations, totalEMI)
	shock10, _ = ComputeFlexibilityBuffer(monthlyIncome.Mul(alphaShock10), fixedObligations, totalEMI)
	shock20, _ = ComputeFlexibilityBuffer(monthlyIncome.Mul(alphaShock20), fixedObligations, totalEMI)
	return current, shock10, shock20
}

// WorstSurvivabilityMonthsAcrossIncomeShocks returns the minimum buffer/EMI months across
// full income, −10%, and −20% shocks (most fragile scenario for survivability banding).
func WorstSurvivabilityMonthsAcrossIncomeShocks(
	monthlyIncome, fixedObligations, totalEMI decimal.Decimal,
) decimal.Decimal {
	alphas := []decimal.Decimal{alphaFull, alphaShock10, alphaShock20}
	var worst decimal.Decimal
	first := true
	for _, a := range alphas {
		inc := monthlyIncome.Mul(a)
		_, months := ComputeFlexibilityBuffer(inc, fixedObligations, totalEMI)
		if first || months.LessThan(worst) {
			worst = months
			first = false
		}
	}
	return worst
}
