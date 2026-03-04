package stress_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
	stress_v1 "DebtEase/internal/domain/stress/v1"
)

// Golden snapshot: clean borrower, single secured mortgage, resilient & stable.
func TestGolden_CleanBorrower_ResilientStable(t *testing.T) {
	t.Parallel()
	debts := []stress_v1.StressDebtEntry{
		debt("hl", "10000", 120, true, cv1.InstrumentMortgage),
	}
	input := makeInput("100000", "5000", debts)
	outcome := stress_v1.Run(input)

	if outcome.StressLevel != stress_v1.StressResilient {
		t.Fatalf("golden: expected RESILIENT, got %s", outcome.StressLevel)
	}
	if outcome.SurvivabilityBand != stress_v1.SurvivabilityStable {
		t.Fatalf("golden: expected STABLE, got %s", outcome.SurvivabilityBand)
	}
	if outcome.EngineVersion != stress_v1.EngineVersion {
		t.Fatalf("golden: engine version changed to %s", outcome.EngineVersion)
	}
}

// Golden snapshot: fragile case under income shock, with unsecured stack.
func TestGolden_FragileCase_FailureCategoriesNonEmpty(t *testing.T) {
	t.Parallel()
	debts := []stress_v1.StressDebtEntry{
		debt("pl1", "22000", 24, false, cv1.InstrumentPersonalLoan),
		debt("bnpl1", "1000", 3, false, cv1.InstrumentBNPL),
		debt("bnpl2", "1000", 3, false, cv1.InstrumentBNPL),
	}
	input := makeInput("30000", "5000", debts)

	outcome := stress_v1.Run(input)
	if outcome.StressLevel != stress_v1.StressFragile && outcome.StressLevel != stress_v1.StressSensitive {
		t.Fatalf("golden: expected FRAGILE or SENSITIVE under stress, got %s", outcome.StressLevel)
	}
	if len(outcome.FailedDebtCategories) == 0 {
		t.Fatalf("golden: expected some failed categories in fragile case")
	}
}

// Golden snapshot: near-zero buffer -> SENSITIVE.
func TestGolden_SensitiveCase_ThinBuffer(t *testing.T) {
	t.Parallel()
	debts := []stress_v1.StressDebtEntry{
		debt("pl1", "25000", 24, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("60000", "10000", debts)

	outcome := stress_v1.Run(input)
	if outcome.StressLevel == stress_v1.StressResilient {
		t.Fatalf("golden: expected at least SENSITIVE, got RESILIENT")
	}
}

