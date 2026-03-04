package stress_v1

// AssertDeterministic panics if Run(input) would not produce the same result when called twice.
// Intended for use in tests to ensure the pipeline is deterministic.
func AssertDeterministic(input StressInput) {
	r1 := Run(input)
	r2 := Run(input)
	if !stressOutcomeEqual(r1, r2) {
		panic("stress engine must be deterministic: same inputs produced different outputs")
	}
}

// AssertValidStressLevel panics if result.StressLevel is not one of the valid constants.
func AssertValidStressLevel(result StressOutcome) {
	switch result.StressLevel {
	case StressResilient, StressSensitive, StressFragile:
		return
	default:
		panic("invalid stress level: " + string(result.StressLevel))
	}
}

// AssertValidSurvivabilityBand panics if result.SurvivabilityBand is not valid.
func AssertValidSurvivabilityBand(result StressOutcome) {
	switch result.SurvivabilityBand {
	case SurvivabilityImmediate, SurvivabilityShort, SurvivabilityModerate, SurvivabilityStable:
		return
	default:
		panic("invalid survivability band: " + string(result.SurvivabilityBand))
	}
}

// AssertKeyDriversMax3 panics if more than 3 key drivers are returned.
func AssertKeyDriversMax3(result StressOutcome) {
	if len(result.KeyDrivers) > 3 {
		panic("too many key drivers: expected at most 3")
	}
}

func stressOutcomeEqual(a, b StressOutcome) bool {
	if a.StressLevel != b.StressLevel ||
		a.SurvivabilityBand != b.SurvivabilityBand ||
		a.EngineVersion != b.EngineVersion {
		return false
	}
	if !a.FlexibilityAfterShock.Equal(b.FlexibilityAfterShock) {
		return false
	}
	if len(a.FailedDebtCategories) != len(b.FailedDebtCategories) {
		return false
	}
	for i, c := range a.FailedDebtCategories {
		if c != b.FailedDebtCategories[i] {
			return false
		}
	}
	if len(a.KeyDrivers) != len(b.KeyDrivers) {
		return false
	}
	for i, d := range a.KeyDrivers {
		if d != b.KeyDrivers[i] {
			return false
		}
	}
	return true
}

