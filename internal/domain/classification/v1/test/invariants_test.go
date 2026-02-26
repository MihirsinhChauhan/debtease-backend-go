package classification_v1

import (
	"math/rand"
	"reflect"
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
)

// TestDeterminism_SameInputSameOutput ensures identical inputs produce identical outputs.
func TestDeterminism_SameInputSameOutput(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "5000", []cv1.DebtEntry{
		debtEntry("d1", "15000", 24, false, cv1.InstrumentPersonalLoan),
	})
	cv1.AssertDeterministic(input)
}

// TestNoHiddenState_ShuffleDebtOrder verifies classification does not depend on debt order.
func TestNoHiddenState_ShuffleDebtOrder(t *testing.T) {
	t.Parallel()
	debts := []cv1.DebtEntry{
		debtEntry("a", "5000", 12, false, cv1.InstrumentBNPL),
		debtEntry("b", "10000", 36, false, cv1.InstrumentPersonalLoan),
		debtEntry("c", "7000", 24, false, cv1.InstrumentCreditCard),
	}
	input1 := makeInput("50000", "0", debts)
	rng := rand.New(rand.NewSource(42))
	shuffled := make([]cv1.DebtEntry, len(debts))
	copy(shuffled, debts)
	rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	input2 := makeInput("50000", "0", shuffled)

	r1 := cv1.Classify(input1)
	r2 := cv1.Classify(input2)

	if r1.Treatment != r2.Treatment {
		t.Fatalf("treatment depends on order: %s vs %s", r1.Treatment, r2.Treatment)
	}
	if len(r1.Flags) != len(r2.Flags) {
		t.Fatalf("flags depend on order: %v vs %v", r1.Flags, r2.Flags)
	}
	if !r1.Metrics.EMIRatio.Equal(r2.Metrics.EMIRatio) {
		t.Fatalf("EMI ratio depends on order")
	}
}

// TestNoProbabilisticOutputs ensures result struct has no probability-like fields.
func TestNoProbabilisticOutputs(t *testing.T) {
	t.Parallel()
	var r cv1.ClassificationResult
	typ := reflect.TypeOf(r)
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		name := f.Name
		// Disallow fields that suggest probability or stochastic output
		if name == "Probability" || name == "RiskScore" || name == "DefaultProbability" {
			t.Fatalf("probabilistic or score field introduced: %s", name)
		}
	}
}

// TestAssertValidTreatment_ValidResult does not panic.
func TestAssertValidTreatment_ValidResult(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "0", []cv1.DebtEntry{
		debtEntry("d1", "10000", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	cv1.AssertValidTreatment(result)
}

// TestAssertMetricsNonNegative_ValidMetrics does not panic.
func TestAssertMetricsNonNegative_ValidMetrics(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "0", []cv1.DebtEntry{
		debtEntry("d1", "15000", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	cv1.AssertMetricsNonNegative(result.Metrics)
}

// TestAssertFlagsMatchTreatment_ValidResult does not panic.
func TestAssertFlagsMatchTreatment_ValidResult(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "0", []cv1.DebtEntry{
		debtEntry("d1", "10000", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := cv1.Classify(input)
	cv1.AssertFlagsMatchTreatment(result)
}
