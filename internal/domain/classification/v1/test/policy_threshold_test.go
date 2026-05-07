package classification_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
	"github.com/shopspring/decimal"
)

// EMI ratio boundary: default MaxSafeEMIRatio = 0.35. We use GreaterThan(0.35) so 0.35 is safe, 0.351 triggers.
func TestEMIRatioBoundary_BelowThreshold_NoFlag(t *testing.T) {
	t.Parallel()
	// ratio 0.349: income 100000, EMI 34900
	input := makeInput("100000", "0", []cv1.DebtEntry{
		debtEntry("d1", "34900", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	if hasFlag(result.Flags, cv1.FlagHighCashflowRigidity) {
		t.Fatalf("0.349 should not trigger HIGH_CASHFLOW_RIGIDITY")
	}
}

func TestEMIRatioBoundary_AtThreshold_NoFlag(t *testing.T) {
	t.Parallel()
	// ratio exactly 0.35
	input := makeInput("100000", "0", []cv1.DebtEntry{
		debtEntry("d1", "35000", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	if hasFlag(result.Flags, cv1.FlagHighCashflowRigidity) {
		t.Fatalf("0.35 at threshold should not trigger (inclusive safe side)")
	}
}

func TestEMIRatioBoundary_AboveThreshold_Flags(t *testing.T) {
	t.Parallel()
	// ratio 0.351
	input := makeInput("100000", "0", []cv1.DebtEntry{
		debtEntry("d1", "35100", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	if !hasFlag(result.Flags, cv1.FlagHighCashflowRigidity) {
		t.Fatalf("0.351 should trigger HIGH_CASHFLOW_RIGIDITY")
	}
}

func TestPolicyOverride_ChangingThresholds_ChangesClassification(t *testing.T) {
	t.Parallel()
	// Same input: EMI ratio 0.4. With default (0.35) we get HIGH_CASHFLOW_RIGIDITY and NEEDS_MONITORING.
	// With relaxed threshold 0.5 we should get no flag and SAFE_TO_HOLD.
	input := makeInput("100000", "0", []cv1.DebtEntry{
		debtEntry("d1", "40000", 24, false, cv1.InstrumentPersonalLoan),
	})
	defaultResult := cv1.Classify(input)
	relaxed := cv1.PolicyThresholds{
		MaxSafeEMIRatio:       decimal.RequireFromString("0.5"),
		MaxFragmentationIndex: decimal.NewFromInt(4),
		MinShockBufferMonths:  decimal.RequireFromString("1.5"),
		MaxUnsecuredRatio:     decimal.RequireFromString("0.60"),
	}
	relaxedResult := cv1.ClassifyWithPolicy(input, relaxed)

	// Metrics must be unchanged
	if !defaultResult.Metrics.EMIRatio.Equal(relaxedResult.Metrics.EMIRatio) {
		t.Fatalf("policy change must not alter metrics")
	}
	// Classification can change
	if defaultResult.Treatment == relaxedResult.Treatment && len(defaultResult.Flags) == len(relaxedResult.Flags) {
		t.Fatalf("relaxed threshold should change classification or flags: default %s %v vs relaxed %s %v",
			defaultResult.Treatment, defaultResult.Flags, relaxedResult.Treatment, relaxedResult.Flags)
	}
}

