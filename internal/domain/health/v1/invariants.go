package health_v1

// AssertScoreInRange panics if result.Score is not in [0, 100].
func AssertScoreInRange(result HealthScoreResult) {
	if result.Score < 0 || result.Score > 100 {
		panic("health score must be in [0, 100]")
	}
}

// AssertValidStatus panics if result.Status is not one of the valid constants.
func AssertValidStatus(result HealthScoreResult) {
	switch result.Status {
	case StatusStable, StatusSensitive, StatusFragile:
		return
	default:
		panic("invalid health status: " + string(result.Status))
	}
}

// AssertDeterministic panics if Run(input) would not produce the same result when called twice.
func AssertDeterministic(input HealthScoreInput) {
	r1 := Run(input)
	r2 := Run(input)
	if r1.Score != r2.Score || r1.Status != r2.Status || r1.EngineVersion != r2.EngineVersion {
		panic("health score engine must be deterministic: same inputs produced different outputs")
	}
	if len(r1.Drivers) != len(r2.Drivers) {
		panic("health score drivers length must be deterministic")
	}
	for i := range r1.Drivers {
		if r1.Drivers[i] != r2.Drivers[i] {
			panic("health score drivers must be deterministic")
		}
	}
}
