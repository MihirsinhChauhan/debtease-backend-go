package classification_v1

// Classify runs the full classification pipeline and returns the treatment category,
// flags, confidence, and context metrics. Same inputs always produce same output.
func Classify(input ClassificationInput) ClassificationResult {
	return ClassifyWithPolicy(input, DefaultPolicyThresholds())
}

// ClassifyWithPolicy runs the pipeline with explicit policy thresholds (for testing or regional overrides).
func ClassifyWithPolicy(input ClassificationInput, thresholds PolicyThresholds) ClassificationResult {
	metrics := ComputeAllMetrics(input)
	flags := applyThresholds(input, metrics, thresholds)
	treatment := assignTreatment(flags, metrics, thresholds)
	confidence := computeConfidence(input, flags)
	return ClassificationResult{
		Treatment:     treatment,
		Flags:         flags,
		Confidence:    confidence,
		Metrics:       metrics,
		EngineVersion: EngineVersion,
	}
}

// applyThresholds evaluates policy thresholds and returns the set of triggered flags.
func applyThresholds(input ClassificationInput, m ContextMetrics, t PolicyThresholds) []Flag {
	var flags []Flag
	if m.EMIRatio.GreaterThan(t.MaxSafeEMIRatio) {
		flags = append(flags, FlagHighCashflowRigidity)
	}
	if m.FragmentationIndex.GreaterThanOrEqual(t.MaxFragmentationIndex) {
		flags = append(flags, FlagShortTermDebtStacking)
	}
	// Only flag low shock buffer when there are debts (otherwise buffer is not meaningful).
	if len(input.Debts) > 0 && m.ShockBufferMonths.LessThan(t.MinShockBufferMonths) {
		flags = append(flags, FlagLowShockBuffer)
	}
	if m.UnsecuredRatio.GreaterThan(t.MaxUnsecuredRatio) && !m.UnsecuredRatio.Equal(decimalZero) {
		flags = append(flags, FlagUnsecuredOverexposure)
	}
	return flags
}

func assignTreatment(flags []Flag, m ContextMetrics, t PolicyThresholds) TreatmentCategory {
	hasHighRigidity := containsFlag(flags, FlagHighCashflowRigidity)
	hasStacking := containsFlag(flags, FlagShortTermDebtStacking)
	hasLowBuffer := containsFlag(flags, FlagLowShockBuffer)
	hasUnsecured := containsFlag(flags, FlagUnsecuredOverexposure)

	// ACTIVELY_REDUCE: high unsecured + (high rigidity or low buffer)
	if hasUnsecured && (hasHighRigidity || hasLowBuffer) {
		return ActivelyReduce
	}
	// AVOID_ADDING: high fragmentation OR low buffer
	if hasStacking || hasLowBuffer {
		return AvoidAdding
	}
	// NEEDS_MONITORING: moderate stress (e.g. high EMI ratio but structure stable)
	if hasHighRigidity {
		return NeedsMonitoring
	}
	// SAFE_TO_HOLD: no flags or only unsecured without other stress
	return SafeToHold
}

func containsFlag(flags []Flag, f Flag) bool {
	for _, x := range flags {
		if x == f {
			return true
		}
	}
	return false
}

func computeConfidence(input ClassificationInput, flags []Flag) ConfidenceLevel {
	if input.Context.MonthlyNetIncome.LessThanOrEqual(decimalZero) {
		return ConfidenceLow
	}
	if len(input.Debts) == 0 {
		return ConfidenceLow
	}
	if len(flags) >= 2 {
		return ConfidenceHigh
	}
	if len(flags) == 1 {
		return ConfidenceMedium
	}
	return ConfidenceHigh
}
