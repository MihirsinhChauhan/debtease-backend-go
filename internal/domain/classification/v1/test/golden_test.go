package classification_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
	"github.com/shopspring/decimal"
)

// Golden snapshot: freeze inputs, policy, and expected output. Do not change without new engine version.
func TestGolden_DocExample_AvoidAdding(t *testing.T) {
	t.Parallel()
	// From debt_classifications_user_risk.md §9: Income 60k, EMIs 22k, 3 BNPL, new loan 3L personal.
	// We classify the portfolio: existing EMIs 22k + assume new loan EMI ~7k (approx) = 29k? Doc says "Existing EMIs: ₹22,000", "3 BNPL".
	// For golden we use a fixed portfolio: 22k EMI from 1 personal loan + 3 BNPL (small EMI each).
	debts := []cv1.DebtEntry{
		debtEntry("pl", "22000", 24, false, cv1.InstrumentPersonalLoan),
		debtEntry("b1", "500", 2, false, cv1.InstrumentBNPL),
		debtEntry("b2", "500", 2, false, cv1.InstrumentBNPL),
		debtEntry("b3", "500", 2, false, cv1.InstrumentBNPL),
	}
	input := makeInput("60000", "0", debts)
	result := cv1.Classify(input)

	// Doc expected AVOID_ADDING or ACTIVELY_REDUCE (both indicate elevated risk); engine may choose ACTIVELY_REDUCE when unsecured + high rigidity.
	if result.Treatment != cv1.AvoidAdding && result.Treatment != cv1.ActivelyReduce {
		t.Fatalf("golden: expected AVOID_ADDING or ACTIVELY_REDUCE, got %s", result.Treatment)
	}
	// Engine version must not change without explicit version bump
	if result.EngineVersion != cv1.EngineVersion {
		t.Fatalf("golden: engine version changed to %s", result.EngineVersion)
	}
	// Metrics sanity: EMI ratio ~23.5/60 ≈ 0.39
	expectedEMIRatioMin := decimal.RequireFromString("0.38")
	expectedEMIRatioMax := decimal.RequireFromString("0.42")
	if result.Metrics.EMIRatio.LessThan(expectedEMIRatioMin) || result.Metrics.EMIRatio.GreaterThan(expectedEMIRatioMax) {
		t.Fatalf("golden: EMI ratio drift: got %s", result.Metrics.EMIRatio.String())
	}
}

// Golden: clean borrower snapshot (v1.0.0).
func TestGolden_CleanBorrower_SafeToHold(t *testing.T) {
	t.Parallel()
	input := makeInput("100000", "5000", []cv1.DebtEntry{
		debtEntry("hl", "20000", 120, true, cv1.InstrumentMortgage),
	})
	result := cv1.Classify(input)
	if result.Treatment != cv1.SafeToHold {
		t.Fatalf("golden: expected SAFE_TO_HOLD, got %s", result.Treatment)
	}
	if result.EngineVersion != "v1.0.0" {
		t.Fatalf("golden: engine version must remain v1.0.0 for this snapshot, got %s", result.EngineVersion)
	}
}
