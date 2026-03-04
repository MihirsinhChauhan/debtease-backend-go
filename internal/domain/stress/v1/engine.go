package stress_v1

import "github.com/shopspring/decimal"

// Run executes the stress engine with default policy thresholds.
func Run(input StressInput) StressOutcome {
	return RunWithPolicy(input, DefaultStressPolicy())
}

// RunWithPolicy executes the stress engine with explicit policy thresholds.
func RunWithPolicy(input StressInput, policy StressPolicyThresholds) StressOutcome {
	// No debts -> resilient by definition for this domain.
	if len(input.Debts) == 0 {
		buffer := input.MonthlyNetIncome.Sub(input.FixedObligations)
		return StressOutcome{
			StressLevel:           StressResilient,
			FlexibilityAfterShock: buffer,
			SurvivabilityBand:     SurvivabilityStable,
			FailedDebtCategories:  nil,
			KeyDrivers:            nil,
			EngineVersion:         EngineVersion,
		}
	}

	// Interest rate shock on EMIs.
	shockedEMIs := ApplyInterestRateShock(input.Debts, policy.UnsecuredRateShockBeta)

	var totalEMI decimal.Decimal
	for _, e := range shockedEMIs {
		totalEMI = totalEMI.Add(e)
	}

	// Track worst (most fragile) scenario over income shocks.
	worstLevel := StressResilient
	worstBuffer := decimalZero
	worstMonths := decimalZero

	severity := func(l StressLevel) int {
		switch l {
		case StressFragile:
			return 2
		case StressSensitive:
			return 1
		default:
			return 0
		}
	}

	for _, alpha := range policy.IncomeShockAlphas {
		stressedIncome := input.MonthlyNetIncome.Mul(alpha)
		buffer, months := ComputeFlexibilityBuffer(stressedIncome, input.FixedObligations, totalEMI)
		level := ClassifyStressLevel(buffer, months, policy)
		if severity(level) > severity(worstLevel) {
			worstLevel = level
			worstBuffer = buffer
			worstMonths = months
		}
	}

	// If all scenarios stayed at the initial zero values, compute once with no shock.
	if worstLevel == StressResilient && worstBuffer.Equal(decimalZero) && worstMonths.Equal(decimalZero) {
		buffer, months := ComputeFlexibilityBuffer(input.MonthlyNetIncome, input.FixedObligations, totalEMI)
		worstBuffer = buffer
		worstMonths = months
		worstLevel = ClassifyStressLevel(buffer, months, policy)
	}

	survivability := ComputeSurvivabilityBand(worstMonths, policy)

	// Failure ordering uses the worst-case buffer after shocks.
	failedCategories := ComputeFailureOrder(worstBuffer, shockedEMIs, input.Debts)

	// Ratios for key drivers.
	var unsecuredEMI, shortTermUnsecuredEMI decimal.Decimal
	for i, d := range input.Debts {
		cat := categoryForDebt(d)
		switch cat {
		case FailureShortTermUnsecured:
			shortTermUnsecuredEMI = shortTermUnsecuredEMI.Add(shockedEMIs[i])
			unsecuredEMI = unsecuredEMI.Add(shockedEMIs[i])
		case FailureLongTermUnsecured:
			unsecuredEMI = unsecuredEMI.Add(shockedEMIs[i])
		}
	}

	unsecuredRatio := ratio(unsecuredEMI, totalEMI)
	shortTermRatio := ratio(shortTermUnsecuredEMI, totalEMI)

	keyDrivers := ComputeKeyDrivers(worstBuffer, worstMonths, unsecuredRatio, shortTermRatio, policy)

	return StressOutcome{
		StressLevel:           worstLevel,
		FlexibilityAfterShock: worstBuffer,
		SurvivabilityBand:     survivability,
		FailedDebtCategories:  failedCategories,
		KeyDrivers:            keyDrivers,
		EngineVersion:         EngineVersion,
	}
}

