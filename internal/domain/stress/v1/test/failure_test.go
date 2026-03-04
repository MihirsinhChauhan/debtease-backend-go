package stress_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
	stress_v1 "DebtEase/internal/domain/stress/v1"

	"github.com/shopspring/decimal"
)

func TestComputeFailureOrder_DeficitConsumesInPriorityOrder(t *testing.T) {
	t.Parallel()

	debts := []stress_v1.StressDebtEntry{
		debt("bnpl1", "1000", 3, false, cv1.InstrumentBNPL),
		debt("short_unsec", "2000", 6, false, cv1.InstrumentPersonalLoan),
		debt("long_unsec", "3000", 24, false, cv1.InstrumentPersonalLoan),
		debt("secured", "4000", 60, true, cv1.InstrumentMortgage),
	}

	shocked := []decimal.Decimal{
		decimal.RequireFromString("1000"),
		decimal.RequireFromString("2000"),
		decimal.RequireFromString("3000"),
		decimal.RequireFromString("4000"),
	}

	// Deficit 2500 should drop BNPL first (1000) then short-term unsecured (2000).
	deficit := decimal.RequireFromString("-2500")
	cats := stress_v1.ComputeFailureOrder(deficit, shocked, debts)

	if len(cats) < 2 {
		t.Fatalf("expected at least 2 failure categories, got %v", cats)
	}
	if cats[0] != stress_v1.FailureBNPLOrZeroEMI {
		t.Fatalf("expected BNPL first, got %s", cats[0])
	}
	if cats[1] != stress_v1.FailureShortTermUnsecured {
		t.Fatalf("expected SHORT_TERM_UNSECURED second, got %s", cats[1])
	}
}

