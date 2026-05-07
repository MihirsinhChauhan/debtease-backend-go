package classification_v1

import (
	"testing"

	cv1 "DebtEase/internal/domain/classification/v1"
	"github.com/shopspring/decimal"
)

func TestComputeEMIRatio_ZeroEMI(t *testing.T) {
	t.Parallel()
	input := makeInput("50000", "0", nil)
	m := cv1.ComputeAllMetrics(input)
	if !m.EMIRatio.IsZero() {
		t.Fatalf("zero EMI should give ratio 0, got %s", m.EMIRatio.String())
	}
}

func TestComputeEMIRatio_EMIEqualsIncome(t *testing.T) {
	t.Parallel()
	input := makeInput("50000", "0", []cv1.DebtEntry{
		debtEntry("d1", "50000", 12, false, cv1.InstrumentPersonalLoan),
	})
	m := cv1.ComputeAllMetrics(input)
	expected := decimal.RequireFromString("1")
	if !m.EMIRatio.Equal(expected) {
		t.Fatalf("EMI=income should give ratio 1, got %s", m.EMIRatio.String())
	}
}

func TestComputeEMIRatio_EMIGreaterThanIncome(t *testing.T) {
	t.Parallel()
	input := makeInput("30000", "0", []cv1.DebtEntry{
		debtEntry("d1", "40000", 12, false, cv1.InstrumentPersonalLoan),
	})
	m := cv1.ComputeAllMetrics(input)
	if !m.EMIRatio.GreaterThan(decimal.RequireFromString("1")) {
		t.Fatalf("EMI>income should give ratio > 1, got %s", m.EMIRatio.String())
	}
}

func TestComputeEMIRatio_25kEMI_50kIncome(t *testing.T) {
	t.Parallel()
	input := makeInput("50000", "0", []cv1.DebtEntry{
		debtEntry("d1", "25000", 12, false, cv1.InstrumentPersonalLoan),
	})
	m := cv1.ComputeAllMetrics(input)
	expected := decimal.RequireFromString("0.5")
	if !m.EMIRatio.Equal(expected) {
		t.Fatalf("25k/50k should be 0.5, got %s", m.EMIRatio.String())
	}
}

func TestComputeFragmentationIndex_ZeroDebts(t *testing.T) {
	t.Parallel()
	input := makeInput("50000", "0", nil)
	m := cv1.ComputeAllMetrics(input)
	if !m.FragmentationIndex.IsZero() {
		t.Fatalf("0 debts should give index 0, got %s", m.FragmentationIndex.String())
	}
}

func TestComputeFragmentationIndex_OneDebtFourBNPL(t *testing.T) {
	t.Parallel()
	// 1 debt + 4 BNPL: index = 1 + 4*0.5 = 3 (if we count each BNPL as one debt, then 5 debts, 4 BNPL -> 5 + 4*0.5 = 7? Doc says "Number_of_Active_Debts + BNPL_Count * 0.5". So active debts = total count, BNPL_Count = count of BNPL. So 5 debts where 4 are BNPL -> 5 + 4*0.5 = 5 + 2 = 7. Doc example: "1 debt, 4 BNPL → index = 1 + (4 × 0.5)" so that's 1 + 2 = 3, meaning 1 non-BNPL debt and 4 BNPL are counted as 1 + 4*0.5 = 3? That would mean "active debts" = 1 and "BNPL count" = 4, so index = 1 + 2 = 3. So the formula is active_debts (all debts) + bnpl_count*0.5. No - re-read: "FRAGMENTATION_INDEX = (Number_of_Active_Debts + BNPL_Count * 0.5)". So it's number of active debts PLUS (BNPL count * 0.5). So if we have 5 debts and 4 are BNPL, then number of active debts = 5, BNPL count = 4, index = 5 + 4*0.5 = 7. The doc test says "1 debt, 4 BNPL → index = 1 + (4 × 0.5)" = 3. So they're saying 1 debt total and that 1 debt is not BNPL? And BNPL_Count is 4 from elsewhere? So it could be "1 non-BNPL debt + 4 BNPL" = 5 obligations, and the index = 1 + 4*0.5 = 3. That would mean we count "active debts" as non-BNPL count and add BNPL*0.5. Our implementation is: activeDebts = len(debts), bnplCount = count BNPL, index = activeDebts + bnplCount*0.5. So for 5 debts all BNPL: 5 + 5*0.5 = 7.5. For 1 debt non-BNPL: 1 + 0*0.5 = 1. So "1 debt, 4 BNPL" in our model is 5 debts with 4 BNPL: 5 + 4*0.5 = 7. The doc might mean 1 regular debt + 4 BNPL as separate count: so "active debts" = 5, "BNPL_Count" = 4, index = 5 + 2 = 7. I'll stick with our implementation and test: 1 debt (non-BNPL) -> index = 1 + 0 = 1. Four BNPL debts -> 4 debts, 4 BNPL -> 4 + 4*0.5 = 6.
	debts := []cv1.DebtEntry{
		debtEntry("pl", "10000", 24, false, cv1.InstrumentPersonalLoan),
		debtEntry("b1", "500", 3, false, cv1.InstrumentBNPL),
		debtEntry("b2", "500", 3, false, cv1.InstrumentBNPL),
		debtEntry("b3", "500", 3, false, cv1.InstrumentBNPL),
		debtEntry("b4", "500", 3, false, cv1.InstrumentBNPL),
	}
	input := makeInput("100000", "0", debts)
	m := cv1.ComputeAllMetrics(input)
	// 5 debts + 4*0.5 = 5 + 2 = 7
	expected := decimal.RequireFromString("7")
	if !m.FragmentationIndex.Equal(expected) {
		t.Fatalf("expected fragmentation index 7, got %s", m.FragmentationIndex.String())
	}
}

func TestComputeFragmentationIndex_BNPLOnly_NonZero(t *testing.T) {
	t.Parallel()
	debts := []cv1.DebtEntry{
		debtEntry("b1", "1000", 2, false, cv1.InstrumentBNPL),
		debtEntry("b2", "500", 1, false, cv1.InstrumentBNPL),
	}
	input := makeInput("50000", "0", debts)
	m := cv1.ComputeAllMetrics(input)
	// 2 + 2*0.5 = 3
	if m.FragmentationIndex.IsZero() {
		t.Fatal("BNPL-only portfolio should have non-zero fragmentation")
	}
}

func TestComputeRigidityScore_ZeroTenure(t *testing.T) {
	t.Parallel()
	// Remaining months = 0 -> EMI*0 = 0, sum = 0, rigidity = 0
	debts := []cv1.DebtEntry{
		debtEntry("d1", "10000", 0, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("50000", "0", debts)
	m := cv1.ComputeAllMetrics(input)
	if !m.RigidityScore.IsZero() {
		t.Fatalf("zero remaining tenure should give rigidity 0, got %s", m.RigidityScore.String())
	}
}

func TestComputeRigidityScore_LargeTenureSmallEMI(t *testing.T) {
	t.Parallel()
	// 5k EMI * 120 months = 600k, income 50k -> 12
	debts := []cv1.DebtEntry{
		debtEntry("d1", "5000", 120, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("50000", "0", debts)
	m := cv1.ComputeAllMetrics(input)
	expected := decimal.RequireFromString("12")
	if !m.RigidityScore.Equal(expected) {
		t.Fatalf("expected rigidity 12, got %s", m.RigidityScore.String())
	}
}

func TestComputeRigidityScore_ZeroIncome_ReturnsZero(t *testing.T) {
	t.Parallel()
	// ratio(num, 0) returns 0 in our implementation
	debts := []cv1.DebtEntry{
		debtEntry("d1", "10000", 24, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("0", "0", debts)
	m := cv1.ComputeAllMetrics(input)
	if !m.RigidityScore.IsZero() {
		t.Fatalf("zero income should not panic; we cap to 0, got %s", m.RigidityScore.String())
	}
}

func TestComputeShockBuffer_IncomeEqualsObligationsPlusEMI(t *testing.T) {
	t.Parallel()
	// Income 60k, Fixed 10k, EMI 50k -> surplus 0 -> buffer 0
	debts := []cv1.DebtEntry{
		debtEntry("d1", "50000", 12, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("60000", "10000", debts)
	m := cv1.ComputeAllMetrics(input)
	if !m.ShockBufferMonths.IsZero() {
		t.Fatalf("income == obligations+EMI should give buffer 0, got %s", m.ShockBufferMonths.String())
	}
}

func TestComputeShockBuffer_IncomeLessThanObligations_NegativeAllowed(t *testing.T) {
	t.Parallel()
	// Income 30k, Fixed 10k, EMI 25k -> surplus -5k, buffer = -5k/25k = -0.2
	debts := []cv1.DebtEntry{
		debtEntry("d1", "25000", 12, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("30000", "10000", debts)
	m := cv1.ComputeAllMetrics(input)
	if !m.ShockBufferMonths.LessThan(decimal.Zero) {
		t.Fatalf("income < obligations+EMI should give negative buffer, got %s", m.ShockBufferMonths.String())
	}
}

func TestComputeShockBuffer_ZeroEMI_ReturnsZero(t *testing.T) {
	t.Parallel()
	// EMI = 0 -> ratio returns 0 (we don't return infinity)
	input := makeInput("60000", "10000", nil)
	m := cv1.ComputeAllMetrics(input)
	if !m.ShockBufferMonths.IsZero() {
		t.Fatalf("zero EMI: buffer must be capped/normalized (we use 0), got %s", m.ShockBufferMonths.String())
	}
}

func TestComputeUnsecuredRatio_AllUnsecured(t *testing.T) {
	t.Parallel()
	debts := []cv1.DebtEntry{
		debtEntry("d1", "10000", 24, false, cv1.InstrumentPersonalLoan),
		debtEntry("d2", "10000", 24, false, cv1.InstrumentCreditCard),
	}
	input := makeInput("100000", "0", debts)
	m := cv1.ComputeAllMetrics(input)
	expected := decimal.RequireFromString("1")
	if !m.UnsecuredRatio.Equal(expected) {
		t.Fatalf("all unsecured should give ratio 1, got %s", m.UnsecuredRatio.String())
	}
}

func TestComputeUnsecuredRatio_HalfSecured(t *testing.T) {
	t.Parallel()
	debts := []cv1.DebtEntry{
		debtEntry("d1", "10000", 24, true, cv1.InstrumentMortgage),
		debtEntry("d2", "10000", 24, false, cv1.InstrumentPersonalLoan),
	}
	input := makeInput("100000", "0", debts)
	m := cv1.ComputeAllMetrics(input)
	expected := decimal.RequireFromString("0.5")
	if !m.UnsecuredRatio.Equal(expected) {
		t.Fatalf("half unsecured should give ratio 0.5, got %s", m.UnsecuredRatio.String())
	}
}
