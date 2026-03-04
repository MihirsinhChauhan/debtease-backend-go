package stress_v1

import "github.com/shopspring/decimal"

// ComputeFlexibilityBuffer returns:
//   buffer  = income - fixedExpenses - totalEMI
//   months  = buffer / totalEMI (capped via safe ratio helper)
func ComputeFlexibilityBuffer(
	income, fixedExpenses, totalEMI decimal.Decimal,
) (buffer decimal.Decimal, months decimal.Decimal) {
	buffer = income.Sub(fixedExpenses).Sub(totalEMI)
	months = ratio(buffer, totalEMI)
	return buffer, months
}

// ComputeSurvivabilityBand maps buffer months into a coarse band using policy thresholds.
func ComputeSurvivabilityBand(months decimal.Decimal, policy StressPolicyThresholds) SurvivabilityBand {
	if months.LessThan(policy.ImmediateStressMaxMonths) {
		return SurvivabilityImmediate
	}
	if months.LessThan(policy.ShortBandMaxMonths) {
		return SurvivabilityShort
	}
	if months.LessThan(policy.ModerateBandMaxMonths) {
		return SurvivabilityModerate
	}
	return SurvivabilityStable
}

// ClassifyStressLevel assigns a qualitative stress level based on buffer and months.
func ClassifyStressLevel(
	buffer, months decimal.Decimal,
	policy StressPolicyThresholds,
) StressLevel {
	if buffer.LessThan(decimalZero) {
		return StressFragile
	}
	if months.LessThan(policy.ResilientMinMonths) {
		return StressSensitive
	}
	return StressResilient
}

// ComputeKeyDrivers returns up to three deterministic driver strings explaining the outcome.
// unsecuredRatio and shortTermRatio are shares (0–1) of total shocked EMI.
func ComputeKeyDrivers(
	buffer, months decimal.Decimal,
	unsecuredRatio, shortTermRatio decimal.Decimal,
	policy StressPolicyThresholds,
) []string {
	var drivers []string

	if buffer.LessThan(decimalZero) {
		drivers = append(drivers, "Low flexibility buffer")
	} else if months.LessThan(policy.ResilientMinMonths) {
		drivers = append(drivers, "Thin flexibility buffer")
	}

	if unsecuredRatio.GreaterThan(policy.UnsecuredExposureHigh) {
		drivers = append(drivers, "Unsecured exposure")
	}

	if shortTermRatio.GreaterThan(policy.ShortTermRigidityHigh) {
		drivers = append(drivers, "High short-term rigidity")
	}

	if len(drivers) > 3 {
		return drivers[:3]
	}
	return drivers
}

