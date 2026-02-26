package classification_v1

// AssertDeterministic panics if Classify(input) would not produce the same result when called twice.
// Call from tests to ensure the pipeline is deterministic.
func AssertDeterministic(input ClassificationInput) {
	r1 := Classify(input)
	r2 := Classify(input)
	if !classificationResultEqual(r1, r2) {
		panic("classification must be deterministic: same inputs produced different outputs")
	}
}

// AssertValidTreatment panics if result.Treatment is not one of the four valid categories.
func AssertValidTreatment(result ClassificationResult) {
	switch result.Treatment {
	case SafeToHold, NeedsMonitoring, AvoidAdding, ActivelyReduce:
		return
	default:
		panic("invalid treatment category: " + string(result.Treatment))
	}
}

// AssertMetricsNonNegative panics if EMI ratio, fragmentation index, or unsecured ratio are negative.
func AssertMetricsNonNegative(m ContextMetrics) {
	if m.EMIRatio.LessThan(decimalZero) {
		panic("EMI ratio cannot be negative")
	}
	if m.FragmentationIndex.LessThan(decimalZero) {
		panic("fragmentation index cannot be negative")
	}
	if m.UnsecuredRatio.LessThan(decimalZero) {
		panic("unsecured ratio cannot be negative")
	}
}

// AssertFlagsMatchTreatment panics if flags and treatment are inconsistent
// (e.g. SAFE_TO_HOLD must not have HIGH_CASHFLOW_RIGIDITY, SHORT_TERM_DEBT_STACKING, or LOW_SHOCK_BUFFER).
func AssertFlagsMatchTreatment(result ClassificationResult) {
	for _, f := range result.Flags {
		switch f {
		case FlagHighCashflowRigidity:
			if result.Treatment == SafeToHold {
				panic("HIGH_CASHFLOW_RIGIDITY inconsistent with SAFE_TO_HOLD")
			}
		case FlagShortTermDebtStacking, FlagLowShockBuffer:
			if result.Treatment != AvoidAdding && result.Treatment != ActivelyReduce {
				panic("SHORT_TERM_DEBT_STACKING / LOW_SHOCK_BUFFER require AVOID_ADDING or ACTIVELY_REDUCE")
			}
		case FlagUnsecuredOverexposure:
			// Allowed with any treatment including SafeToHold (structural warning only)
		}
	}
}

func classificationResultEqual(a, b ClassificationResult) bool {
	if a.Treatment != b.Treatment || a.Confidence != b.Confidence || a.EngineVersion != b.EngineVersion {
		return false
	}
	if len(a.Flags) != len(b.Flags) {
		return false
	}
	flagsA := make(map[Flag]struct{})
	for _, f := range a.Flags {
		flagsA[f] = struct{}{}
	}
	for _, f := range b.Flags {
		if _, ok := flagsA[f]; !ok {
			return false
		}
	}
	if !a.Metrics.EMIRatio.Equal(b.Metrics.EMIRatio) ||
		!a.Metrics.FragmentationIndex.Equal(b.Metrics.FragmentationIndex) ||
		!a.Metrics.RigidityScore.Equal(b.Metrics.RigidityScore) ||
		!a.Metrics.ShockBufferMonths.Equal(b.Metrics.ShockBufferMonths) ||
		!a.Metrics.UnsecuredRatio.Equal(b.Metrics.UnsecuredRatio) {
		return false
	}
	return true
}
