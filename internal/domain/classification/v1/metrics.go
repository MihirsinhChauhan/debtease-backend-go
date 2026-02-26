package classification_v1

import "github.com/shopspring/decimal"

var half = decimal.RequireFromString("0.5")

// ComputeEMIRatio returns Total_Monthly_EMI / Monthly_Net_Income.
func ComputeEMIRatio(input ClassificationInput) decimal.Decimal {
	totalEMI := totalEMI(input.Debts)
	return ratio(totalEMI, input.Context.MonthlyNetIncome)
}

// ComputeFragmentationIndex returns Active_Debts + BNPL_Count*0.5.
func ComputeFragmentationIndex(input ClassificationInput) decimal.Decimal {
	n := len(input.Debts)
	if n == 0 {
		return decimalZero
	}
	activeDebts := decimal.NewFromInt(int64(n))
	bnplCount := decimal.NewFromInt(int64(countBNPL(input.Debts)))
	return activeDebts.Add(bnplCount.Mul(half))
}

// ComputeRigidityScore returns SUM(EMI * Remaining_Months) / Monthly_Income.
func ComputeRigidityScore(input ClassificationInput) decimal.Decimal {
	var sum decimal.Decimal
	for _, d := range input.Debts {
		rem := decimal.NewFromInt(int64(d.RemainingMonths))
		sum = sum.Add(d.EMI.Mul(rem))
	}
	return ratio(sum, input.Context.MonthlyNetIncome)
}

// ComputeShockBufferMonths returns (Income - Fixed_Obligations - EMIs) / EMIs.
func ComputeShockBufferMonths(input ClassificationInput) decimal.Decimal {
	totalEMI := totalEMI(input.Debts)
	numerator := input.Context.MonthlyNetIncome.Sub(input.Context.FixedObligations).Sub(totalEMI)
	return ratio(numerator, totalEMI)
}

// ComputeUnsecuredRatio returns Unsecured_EMI / Total_EMI.
func ComputeUnsecuredRatio(input ClassificationInput) decimal.Decimal {
	totalEMI := totalEMI(input.Debts)
	if totalEMI.IsZero() {
		return decimalZero
	}
	var unsecured decimal.Decimal
	for _, d := range input.Debts {
		if !d.Secured {
			unsecured = unsecured.Add(d.EMI)
		}
	}
	return ratio(unsecured, totalEMI)
}

// ComputeAllMetrics computes all five context metrics.
func ComputeAllMetrics(input ClassificationInput) ContextMetrics {
	return ContextMetrics{
		EMIRatio:           ComputeEMIRatio(input),
		FragmentationIndex: ComputeFragmentationIndex(input),
		RigidityScore:      ComputeRigidityScore(input),
		ShockBufferMonths:  ComputeShockBufferMonths(input),
		UnsecuredRatio:     ComputeUnsecuredRatio(input),
	}
}

func totalEMI(debts []DebtEntry) decimal.Decimal {
	var sum decimal.Decimal
	for _, d := range debts {
		sum = sum.Add(d.EMI)
	}
	return sum
}

func countBNPL(debts []DebtEntry) int {
	var n int
	for _, d := range debts {
		if d.InstrumentType == InstrumentBNPL {
			n++
		}
	}
	return n
}
