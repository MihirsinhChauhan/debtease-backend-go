package health_v1

import "github.com/shopspring/decimal"

// ComputeEMIRatio returns Total_Monthly_EMI / Monthly_Income (0–1).
func ComputeEMIRatio(input HealthScoreInput) decimal.Decimal {
	return ratio(input.TotalMonthlyEMI, input.MonthlyIncome)
}

// ComputeShockBufferMonths returns (Savings + Investments) / monthly_burn,
// where monthly_burn = fixed_expenses + variable_expenses + total_monthly_emi (liquidity runway).
func ComputeShockBufferMonths(input HealthScoreInput) decimal.Decimal {
	liquid := input.SavingsCash.Add(input.InvestmentsValue)
	burn := input.FixedExpenses.Add(input.VariableExpenses).Add(input.TotalMonthlyEMI)
	return ratio(liquid, burn)
}

// DebtCount returns the number of active debts (fragmentation).
func DebtCount(input HealthScoreInput) int {
	return len(input.Debts)
}

// ComputeUnsecuredRatio returns Unsecured_EMI / Total_EMI (0–1) from debt portfolio.
func ComputeUnsecuredRatio(input HealthScoreInput) decimal.Decimal {
	var total, unsecured decimal.Decimal
	for _, d := range input.Debts {
		total = total.Add(d.EMI)
		if !d.Secured {
			unsecured = unsecured.Add(d.EMI)
		}
	}
	if total.IsZero() {
		return decimalZero
	}
	return ratio(unsecured, total)
}

// PenaltyEMIRatio returns the penalty (0, 10, or 25) for EMI ratio band.
func PenaltyEMIRatio(emiRatio decimal.Decimal, p HealthPolicyThresholds) int {
	if emiRatio.LessThan(p.EMIRatioLowBound) {
		return 0
	}
	if emiRatio.LessThanOrEqual(p.EMIRatioHighBound) {
		return p.PenaltyEMIRatioMid
	}
	return p.PenaltyEMIRatioHigh
}

// PenaltyShockBuffer returns the penalty (0, 5, 15, or 30) for shock buffer band.
func PenaltyShockBuffer(months decimal.Decimal, p HealthPolicyThresholds) int {
	if months.GreaterThanOrEqual(p.ShockBufferMonths6) {
		return 0
	}
	if months.GreaterThanOrEqual(p.ShockBufferMonths3) {
		return p.PenaltyShockBufferMid
	}
	if months.GreaterThanOrEqual(p.ShockBufferMonths1) {
		return p.PenaltyShockBufferLow
	}
	return p.PenaltyShockBufferNone
}

// PenaltyFragmentation returns the penalty (0, 5, or 15) for debt count.
func PenaltyFragmentation(debtCount int, p HealthPolicyThresholds) int {
	if debtCount <= p.FragmentationMax2 {
		return 0
	}
	if debtCount <= p.FragmentationMax4 {
		return p.PenaltyFragmentationMid
	}
	return p.PenaltyFragmentationHigh
}

// PenaltyUnsecured returns the penalty (0, 5, or 10) for unsecured ratio band.
func PenaltyUnsecured(unsecuredRatio decimal.Decimal, p HealthPolicyThresholds) int {
	if unsecuredRatio.LessThan(p.UnsecuredRatioMid) {
		return 0
	}
	if unsecuredRatio.LessThanOrEqual(p.UnsecuredRatioHigh) {
		return p.PenaltyUnsecuredMid
	}
	return p.PenaltyUnsecuredHigh
}
