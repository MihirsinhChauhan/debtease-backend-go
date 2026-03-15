package health_v1

// Run executes the health score engine with default policy.
func Run(input HealthScoreInput) HealthScoreResult {
	return RunWithPolicy(input, DefaultHealthPolicy())
}

// RunWithPolicy executes the health score engine with explicit policy.
func RunWithPolicy(input HealthScoreInput, policy HealthPolicyThresholds) HealthScoreResult {
	emiRatio := ComputeEMIRatio(input)
	shockBufferMonths := ComputeShockBufferMonths(input)
	debtCount := DebtCount(input)
	unsecuredRatio := ComputeUnsecuredRatio(input)

	// Zero income: treat EMI ratio as max penalty (doc edge case).
	emiPenalty := 0
	if input.MonthlyIncome.IsZero() && input.TotalMonthlyEMI.GreaterThan(decimalZero) {
		emiPenalty = policy.PenaltyEMIRatioHigh
	} else {
		emiPenalty = PenaltyEMIRatio(emiRatio, policy)
	}

	shockPenalty := PenaltyShockBuffer(shockBufferMonths, policy)
	fragPenalty := PenaltyFragmentation(debtCount, policy)
	unsecPenalty := PenaltyUnsecured(unsecuredRatio, policy)

	totalPenalty := emiPenalty + shockPenalty + fragPenalty + unsecPenalty
	score := baseScore - totalPenalty
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	status := scoreToStatus(score)
	drivers := buildDrivers(emiPenalty, shockPenalty, fragPenalty, unsecPenalty, policy)

	return HealthScoreResult{
		Score:         score,
		Status:        status,
		Drivers:       drivers,
		EngineVersion: EngineVersion,
	}
}

func scoreToStatus(score int) HealthStatus {
	switch {
	case score >= 80:
		return StatusStable
	case score >= 60:
		return StatusSensitive
	default:
		return StatusFragile
	}
}

func buildDrivers(emiPenalty, shockPenalty, fragPenalty, unsecPenalty int, policy HealthPolicyThresholds) []string {
	var d []string
	if emiPenalty > 0 {
		d = append(d, "EMI load is high")
	}
	if shockPenalty > 0 {
		d = append(d, "Low financial buffer")
	}
	if fragPenalty > 0 {
		d = append(d, "Multiple short-term debts")
	}
	if unsecPenalty > 0 {
		d = append(d, "High unsecured exposure")
	}
	return d
}
