package stress_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
	stress_v1 "DebtEase/internal/domain/stress/v1"

	"github.com/shopspring/decimal"
)

func debt(id, emi string, months int, secured bool, inst cv1.DebtInstrumentType) stress_v1.StressDebtEntry {
	return stress_v1.StressDebtEntry{
		ID:              id,
		EMI:             decimal.RequireFromString(emi),
		RemainingMonths: months,
		Secured:         secured,
		InstrumentType:  inst,
	}
}

func makeInput(income, fixed string, debts []stress_v1.StressDebtEntry) stress_v1.StressInput {
	return stress_v1.StressInput{
		Debts:            debts,
		MonthlyNetIncome: decimal.RequireFromString(income),
		FixedObligations: decimal.RequireFromString(fixed),
	}
}

func TestRun_NoDebts_ResilientStable(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "10000", nil)
	outcome := stress_v1.Run(input)

	if outcome.StressLevel != stress_v1.StressResilient {
		t.Fatalf("expected RESILIENT, got %s", outcome.StressLevel)
	}
	if outcome.SurvivabilityBand != stress_v1.SurvivabilityStable {
		t.Fatalf("expected STABLE, got %s", outcome.SurvivabilityBand)
	}
}

