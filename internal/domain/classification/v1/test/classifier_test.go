package classification_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
)

// Case 1: Clean borrower — low EMI ratio, 1 long-term secured loan, no BNPL.
func TestClassification_CleanBorrower_SafeToHoldNoFlags(t *testing.T) {
	t.Parallel()
	// Income 100k, EMI 20k -> ratio 0.2. One mortgage (secured), no BNPL.
	input := makeInput("100000", "5000", []cv1.DebtEntry{
		debtEntry("hl", "20000", 120, true, cv1.InstrumentMortgage),
	})
	result := cv1.Classify(input)
	if result.Treatment != cv1.SafeToHold {
		t.Fatalf("expected SAFE_TO_HOLD, got %s", result.Treatment)
	}
	if len(result.Flags) != 0 {
		t.Fatalf("expected no flags, got %v", result.Flags)
	}
}

// Case 2: Moderate stress — EMI ratio ~40%, 2 debts. Unsecured + high rigidity -> ACTIVELY_REDUCE or NEEDS_MONITORING.
func TestClassification_ModerateStress_FlagsHighCashflowRigidity(t *testing.T) {
	t.Parallel()
	// Income 100k, EMI 40k -> ratio 0.4. Buffer: (100k-0-40k)/40k = 1.5 (at threshold).
	input := makeInput("100000", "0", []cv1.DebtEntry{
		debtEntry("d1", "25000", 24, false, cv1.InstrumentPersonalLoan),
		debtEntry("d2", "15000", 36, false, cv1.InstrumentCreditCard),
	})
	result := cv1.Classify(input)
	if result.Treatment != cv1.ActivelyReduce && result.Treatment != cv1.NeedsMonitoring {
		t.Fatalf("expected ACTIVELY_REDUCE or NEEDS_MONITORING, got %s", result.Treatment)
	}
	if !hasFlag(result.Flags, cv1.FlagHighCashflowRigidity) {
		t.Fatalf("expected HIGH_CASHFLOW_RIGIDITY flag, got %v", result.Flags)
	}
}

// Case 3: Debt loop pattern — multiple BNPL, short-term obligations, moderate EMI.
func TestClassification_DebtLoopPattern_AvoidAdding(t *testing.T) {
	t.Parallel()
	// 5 BNPL -> fragmentation 5 + 5*0.5 = 7.5 >= 4. EMI ratio kept moderate so we get AVOID_ADDING from stacking.
	debts := []cv1.DebtEntry{
		debtEntry("b1", "2000", 2, false, cv1.InstrumentBNPL),
		debtEntry("b2", "1500", 2, false, cv1.InstrumentBNPL),
		debtEntry("b3", "1000", 3, false, cv1.InstrumentBNPL),
		debtEntry("b4", "2500", 2, false, cv1.InstrumentBNPL),
		debtEntry("b5", "500", 1, false, cv1.InstrumentBNPL),
	}
	input := makeInput("80000", "0", debts)
	result := cv1.Classify(input)
	if result.Treatment != cv1.AvoidAdding {
		t.Fatalf("expected AVOID_ADDING, got %s", result.Treatment)
	}
	if !hasFlag(result.Flags, cv1.FlagShortTermDebtStacking) {
		t.Fatalf("expected SHORT_TERM_DEBT_STACKING, got %v", result.Flags)
	}
}

// Case 4: Dangerous structure — high unsecured, low shock buffer, long lock-in.
func TestClassification_DangerousStructure_ActivelyReduce(t *testing.T) {
	t.Parallel()
	// High unsecured (100%), low buffer: income 60k, fixed 5k, EMI 40k -> surplus 15k, buffer 15/40 = 0.375 < 1.5.
	// So we get LOW_SHOCK_BUFFER and UNSECURED_OVEREXPOSURE -> ACTIVELY_REDUCE.
	input := makeInput("60000", "5000", []cv1.DebtEntry{
		debtEntry("d1", "25000", 48, false, cv1.InstrumentPersonalLoan),
		debtEntry("d2", "15000", 24, false, cv1.InstrumentCreditCard),
	})
	result := cv1.Classify(input)
	if result.Treatment != cv1.ActivelyReduce {
		t.Fatalf("expected ACTIVELY_REDUCE, got %s", result.Treatment)
	}
	if len(result.Flags) < 2 {
		t.Fatalf("expected multiple flags, got %v", result.Flags)
	}
}

// Priority: when multiple conditions trigger, strongest treatment wins (e.g. shock buffer critical -> ACTIVELY_REDUCE).
func TestClassification_PriorityResolution_ActivelyReduceWins(t *testing.T) {
	t.Parallel()
	// Moderate EMI ratio (e.g. 36%) but critical shock buffer: income 60k, fixed 10k, EMI 22k -> surplus 28k, buffer 28/22 ~1.27 < 1.5.
	// So LOW_SHOCK_BUFFER. Also EMI ratio 22/60 = 0.367 > 0.35 -> HIGH_CASHFLOW_RIGIDITY.
	// Unsecured 100% -> UNSECURED_OVEREXPOSURE. So we have unsecured + low buffer -> ACTIVELY_REDUCE (not just NEEDS_MONITORING).
	input := makeInput("60000", "10000", []cv1.DebtEntry{
		debtEntry("d1", "22000", 36, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	if result.Treatment != cv1.ActivelyReduce {
		t.Fatalf("expected ACTIVELY_REDUCE when shock buffer critical and unsecured, got %s", result.Treatment)
	}
}
