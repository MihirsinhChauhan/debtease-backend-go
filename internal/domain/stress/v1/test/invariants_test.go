package stress_v1

import (
	"reflect"
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
	stress_v1 "DebtEase/internal/domain/stress/v1"
)

func TestDeterminism_SameInputSameOutput(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "5000", []stress_v1.StressDebtEntry{
		debt("d1", "15000", 24, false, cv1.InstrumentPersonalLoan),
	})
	stress_v1.AssertDeterministic(input)
}

func TestAssertValidStressLevel_ValidResult(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "0", []stress_v1.StressDebtEntry{
		debt("d1", "10000", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := stress_v1.Run(input)
	stress_v1.AssertValidStressLevel(result)
}

func TestAssertValidSurvivabilityBand_ValidResult(t *testing.T) {
	t.Parallel()
	input := makeInput("60000", "0", []stress_v1.StressDebtEntry{
		debt("d1", "10000", 24, false, cv1.InstrumentPersonalLoan),
	})
	result := stress_v1.Run(input)
	stress_v1.AssertValidSurvivabilityBand(result)
}

func TestAssertKeyDriversMax3_DoesNotPanic(t *testing.T) {
	t.Parallel()
	input := makeInput("80000", "10000", []stress_v1.StressDebtEntry{
		debt("d1", "30000", 24, false, cv1.InstrumentPersonalLoan),
		debt("d2", "10000", 12, false, cv1.InstrumentPersonalLoan),
	})
	result := stress_v1.Run(input)
	stress_v1.AssertKeyDriversMax3(result)
}

func TestNoProbabilisticOutputs(t *testing.T) {
	t.Parallel()
	var r stress_v1.StressOutcome
	typ := reflect.TypeOf(r)
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		name := f.Name
		if name == "Probability" || name == "RiskScore" || name == "DefaultProbability" {
			t.Fatalf("probabilistic or score field introduced: %s", name)
		}
	}
}

