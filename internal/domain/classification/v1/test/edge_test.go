package classification_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
)

// Edge 1: High-income, high-debt — still flags fragmentation, does not downgrade risk just because income is high.
func TestEdge_HighIncomeHighDebt_StillFlagsFragmentation(t *testing.T) {
	t.Parallel()
	// Income 5L, EMI 2L, multiple BNPL -> fragmentation should still trigger.
	debts := []cv1.DebtEntry{
		debtEntry("b1", "20000", 2, false, cv1.InstrumentBNPL),
		debtEntry("b2", "15000", 2, false, cv1.InstrumentBNPL),
		debtEntry("b3", "25000", 3, false, cv1.InstrumentBNPL),
		debtEntry("b4", "10000", 1, false, cv1.InstrumentBNPL),
		debtEntry("pl", "130000", 24, false, cv1.InstrumentPersonalLoan),
	}
	// Total EMI 200k, income 500k -> ratio 0.4. 5 debts, 4 BNPL -> fragmentation 5 + 4*0.5 = 7 >= 4.
	input := makeInput("500000", "0", debts)
	result := cv1.Classify(input)
	if !hasFlag(result.Flags, cv1.FlagShortTermDebtStacking) {
		t.Fatalf("high income should not downgrade fragmentation risk; expected SHORT_TERM_DEBT_STACKING, got %v", result.Flags)
	}
}

// Edge 2: Low EMI, extreme fragmentation — 8 BNPL, EMI ratio < 20% -> AVOID_ADDING (fragmentation).
func TestEdge_LowEMIExtremeFragmentation_AvoidAdding(t *testing.T) {
	t.Parallel()
	debts := make([]cv1.DebtEntry, 8)
	for i := range debts {
		debts[i] = debtEntry("b"+string(rune('1'+i)), "1000", 2, false, cv1.InstrumentBNPL)
	}
	// 8 * 1000 = 8k EMI, income 50k -> ratio 0.16. Fragmentation 8 + 8*0.5 = 12 >= 4.
	input := makeInput("50000", "0", debts)
	result := cv1.Classify(input)
	if result.Treatment != cv1.AvoidAdding {
		t.Fatalf("expected AVOID_ADDING for extreme fragmentation, got %s", result.Treatment)
	}
	if !hasFlag(result.Flags, cv1.FlagShortTermDebtStacking) {
		t.Fatalf("expected SHORT_TERM_DEBT_STACKING, got %v", result.Flags)
	}
}

// Edge 3: Zero income — classification still returns, no crash; confidence low.
func TestEdge_ZeroIncome_NoCrash(t *testing.T) {
	t.Parallel()
	debts := []cv1.DebtEntry{
		debtEntry("d1", "20000", 24, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("0", "0", debts)
	result := cv1.Classify(input)
	// Must not panic. Result may still have a treatment; confidence should be low.
	if result.Confidence != cv1.ConfidenceLow {
		t.Fatalf("zero income should yield LOW confidence, got %s", result.Confidence)
	}
}

// Edge 4: One massive secured loan — home loan EMI 45%, no other debt. Secured-only -> not ACTIVELY_REDUCE.
func TestEdge_OneMassiveSecuredLoan_NotActivelyReduce(t *testing.T) {
	t.Parallel()
	// Income 100k, EMI 45k (45%), secured mortgage only. Unsecured ratio = 0. Low buffer triggers AVOID_ADDING.
	input := makeInput("100000", "0", []cv1.DebtEntry{
		debtEntry("hl", "45000", 120, true, cv1.InstrumentMortgage),
	})
	result := cv1.Classify(input)
	// Secured-only: never ACTIVELY_REDUCE (that requires unsecured overexposure).
	if result.Treatment == cv1.ActivelyReduce {
		t.Fatalf("secured-only high EMI should not be ACTIVELY_REDUCE")
	}
	// Can be NEEDS_MONITORING (high EMI) or AVOID_ADDING (low buffer).
	if result.Treatment != cv1.NeedsMonitoring && result.Treatment != cv1.AvoidAdding {
		t.Fatalf("expected NEEDS_MONITORING or AVOID_ADDING, got %s", result.Treatment)
	}
}

// Edge 5: Negative buffer but low EMI ratio. Low buffer + unsecured -> ACTIVELY_REDUCE (strongest wins).
func TestEdge_NegativeBufferLowEMIRatio_FlagsLowShockBuffer(t *testing.T) {
	t.Parallel()
	// High fixed expenses: income 40k, fixed 25k, EMI 10k -> surplus 5k, buffer 5/10 = 0.5 < 1.5. EMI ratio 10/40 = 0.25 (low).
	input := makeInput("40000", "25000", []cv1.DebtEntry{
		debtEntry("d1", "10000", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	// Unsecured + low buffer -> ACTIVELY_REDUCE; otherwise AVOID_ADDING.
	if result.Treatment != cv1.ActivelyReduce && result.Treatment != cv1.AvoidAdding {
		t.Fatalf("expected ACTIVELY_REDUCE or AVOID_ADDING, got %s", result.Treatment)
	}
	if !hasFlag(result.Flags, cv1.FlagLowShockBuffer) {
		t.Fatalf("expected LOW_SHOCK_BUFFER, got %v", result.Flags)
	}
}
