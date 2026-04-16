package health_v1

import (
	"testing"

	hv1 "DebtEase/internal/domain/health/v1"
	"github.com/shopspring/decimal"
)

func TestComputeEMIRatio_ZeroIncome(t *testing.T) {
	t.Parallel()
	input := makeInput("0", "0", "0", "0", "0", "20000", nil)
	ratio := hv1.ComputeEMIRatio(input)
	// 20000/0 -> ratio returns 0 (safe division)
	if !ratio.IsZero() {
		t.Fatalf("zero income should give ratio 0 (safe), got %s", ratio.String())
	}
}

func TestComputeEMIRatio_Under30(t *testing.T) {
	t.Parallel()
	// 25k EMI on 100k income = 25% (<30%); add high savings so shock buffer and other penalties are 0.
	input := makeInput("100000", "0", "0", "500000", "0", "25000", nil)
	ratio := hv1.ComputeEMIRatio(input)
	expected := decimal.RequireFromString("0.25")
	if !ratio.Equal(expected) {
		t.Fatalf("25k/100k should be 0.25, got %s", ratio.String())
	}
	result := hv1.Run(input)
	if result.Score != 100 {
		t.Fatalf("low EMI ratio with good buffer: expected score 100, got %d", result.Score)
	}
}

func TestComputeEMIRatio_30To45(t *testing.T) {
	t.Parallel()
	// 40k/100k = 40% -> -10 penalty
	input := makeInput("100000", "0", "0", "500000", "0", "40000", []hv1.HealthDebtEntry{
		debtEntry("d1", "40000", false),
	})
	ratio := hv1.ComputeEMIRatio(input)
	expected := decimal.RequireFromString("0.4")
	if !ratio.Equal(expected) {
		t.Fatalf("40k/100k should be 0.4, got %s", ratio.String())
	}
	policy := hv1.DefaultHealthPolicy()
	penalty := hv1.PenaltyEMIRatio(ratio, policy)
	if penalty != 10 {
		t.Fatalf("40%% EMI ratio should give penalty 10, got %d", penalty)
	}
}

func TestComputeEMIRatio_Over45(t *testing.T) {
	t.Parallel()
	input := makeInput("100000", "0", "0", "500000", "0", "50000", []hv1.HealthDebtEntry{
		debtEntry("d1", "50000", false),
	})
	ratio := hv1.ComputeEMIRatio(input)
	expected := decimal.RequireFromString("0.5")
	if !ratio.Equal(expected) {
		t.Fatalf("50k/100k should be 0.5, got %s", ratio.String())
	}
	policy := hv1.DefaultHealthPolicy()
	penalty := hv1.PenaltyEMIRatio(ratio, policy)
	if penalty != 25 {
		t.Fatalf("50%% EMI ratio should give penalty 25, got %d", penalty)
	}
}

func TestComputeShockBuffer_Over6Months(t *testing.T) {
	t.Parallel()
	// savings 300k, burn 40k -> 7.5 months
	input := makeInput("80000", "10000", "10000", "300000", "0", "20000", nil)
	months := hv1.ComputeShockBufferMonths(input)
	if months.LessThan(decimal.RequireFromString("6")) {
		t.Fatalf("300k/40k should be > 6 months, got %s", months.String())
	}
	policy := hv1.DefaultHealthPolicy()
	penalty := hv1.PenaltyShockBuffer(months, policy)
	if penalty != 0 {
		t.Fatalf(">6 months buffer should give penalty 0, got %d", penalty)
	}
}

func TestComputeShockBuffer_3To6Months(t *testing.T) {
	t.Parallel()
	// 150k / 40k = 3.75 months
	months := decimal.RequireFromString("3.75")
	policy := hv1.DefaultHealthPolicy()
	penalty := hv1.PenaltyShockBuffer(months, policy)
	if penalty != 5 {
		t.Fatalf("3–6 months should give penalty 5, got %d", penalty)
	}
}

func TestComputeShockBuffer_1To3Months(t *testing.T) {
	t.Parallel()
	months := decimal.RequireFromString("2")
	policy := hv1.DefaultHealthPolicy()
	penalty := hv1.PenaltyShockBuffer(months, policy)
	if penalty != 15 {
		t.Fatalf("1–3 months should give penalty 15, got %d", penalty)
	}
}

func TestComputeShockBuffer_Under1Month(t *testing.T) {
	t.Parallel()
	input := makeInput("50000", "20000", "15000", "5000", "0", "20000", nil)
	// burn = 20+15+20 = 55k, liquid = 5k -> 5/55 < 1
	months := hv1.ComputeShockBufferMonths(input)
	if months.GreaterThanOrEqual(decimal.RequireFromString("1")) {
		t.Fatalf("5k/55k should be < 1 month, got %s", months.String())
	}
	policy := hv1.DefaultHealthPolicy()
	penalty := hv1.PenaltyShockBuffer(months, policy)
	if penalty != 30 {
		t.Fatalf("<1 month should give penalty 30, got %d", penalty)
	}
}

func TestPenaltyFragmentation_0To2Debts(t *testing.T) {
	t.Parallel()
	policy := hv1.DefaultHealthPolicy()
	if hv1.PenaltyFragmentation(0, policy) != 0 {
		t.Fatal("0 debts should give penalty 0")
	}
	if hv1.PenaltyFragmentation(1, policy) != 0 {
		t.Fatal("1 debt should give penalty 0")
	}
	if hv1.PenaltyFragmentation(2, policy) != 0 {
		t.Fatal("2 debts should give penalty 0")
	}
}

func TestPenaltyFragmentation_3To4Debts(t *testing.T) {
	t.Parallel()
	policy := hv1.DefaultHealthPolicy()
	if hv1.PenaltyFragmentation(3, policy) != 5 {
		t.Fatalf("3 debts should give penalty 5, got %d", hv1.PenaltyFragmentation(3, policy))
	}
	if hv1.PenaltyFragmentation(4, policy) != 5 {
		t.Fatalf("4 debts should give penalty 5, got %d", hv1.PenaltyFragmentation(4, policy))
	}
}

func TestPenaltyFragmentation_5OrMore(t *testing.T) {
	t.Parallel()
	policy := hv1.DefaultHealthPolicy()
	if hv1.PenaltyFragmentation(5, policy) != 15 {
		t.Fatalf("5 debts should give penalty 15, got %d", hv1.PenaltyFragmentation(5, policy))
	}
	if hv1.PenaltyFragmentation(10, policy) != 15 {
		t.Fatalf("10 debts should give penalty 15, got %d", hv1.PenaltyFragmentation(10, policy))
	}
}

func TestComputeUnsecuredRatio_AllSecured(t *testing.T) {
	t.Parallel()
	input := makeInput("100000", "0", "0", "0", "0", "30000", []hv1.HealthDebtEntry{
		debtEntry("h1", "20000", true),
		debtEntry("h2", "10000", true),
	})
	ratio := hv1.ComputeUnsecuredRatio(input)
	if !ratio.IsZero() {
		t.Fatalf("all secured should give ratio 0, got %s", ratio.String())
	}
}

func TestComputeUnsecuredRatio_AllUnsecured(t *testing.T) {
	t.Parallel()
	input := makeInput("100000", "0", "0", "0", "0", "25000", []hv1.HealthDebtEntry{
		debtEntry("p1", "15000", false),
		debtEntry("p2", "10000", false),
	})
	ratio := hv1.ComputeUnsecuredRatio(input)
	expected := decimal.RequireFromString("1")
	if !ratio.Equal(expected) {
		t.Fatalf("all unsecured should give ratio 1, got %s", ratio.String())
	}
}

func TestComputeUnsecuredRatio_HalfUnsecured(t *testing.T) {
	t.Parallel()
	input := makeInput("100000", "0", "0", "0", "0", "20000", []hv1.HealthDebtEntry{
		debtEntry("s", "10000", true),
		debtEntry("u", "10000", false),
	})
	ratio := hv1.ComputeUnsecuredRatio(input)
	expected := decimal.RequireFromString("0.5")
	if !ratio.Equal(expected) {
		t.Fatalf("half unsecured should give ratio 0.5, got %s", ratio.String())
	}
}

func TestPenaltyUnsecured_Under40(t *testing.T) {
	t.Parallel()
	policy := hv1.DefaultHealthPolicy()
	ratio := decimal.RequireFromString("0.30")
	if hv1.PenaltyUnsecured(ratio, policy) != 0 {
		t.Fatalf("<40%% unsecured should give penalty 0, got %d", hv1.PenaltyUnsecured(ratio, policy))
	}
}

func TestPenaltyUnsecured_40To60(t *testing.T) {
	t.Parallel()
	policy := hv1.DefaultHealthPolicy()
	ratio := decimal.RequireFromString("0.50")
	if hv1.PenaltyUnsecured(ratio, policy) != 5 {
		t.Fatalf("40–60%% unsecured should give penalty 5, got %d", hv1.PenaltyUnsecured(ratio, policy))
	}
}

func TestPenaltyUnsecured_Over60(t *testing.T) {
	t.Parallel()
	policy := hv1.DefaultHealthPolicy()
	ratio := decimal.RequireFromString("0.70")
	if hv1.PenaltyUnsecured(ratio, policy) != 10 {
		t.Fatalf(">60%% unsecured should give penalty 10, got %d", hv1.PenaltyUnsecured(ratio, policy))
	}
}

func TestRun_NoDebts_ScoreInRange(t *testing.T) {
	t.Parallel()
	input := makeInput("80000", "10000", "10000", "100000", "50000", "0", nil)
	result := hv1.Run(input)
	hv1.AssertScoreInRange(result)
	hv1.AssertValidStatus(result)
}

func TestRun_ScoreClampedTo100(t *testing.T) {
	t.Parallel()
	// Minimal penalties: low EMI, high buffer, 0–2 debts, low unsecured -> score 100
	input := makeInput("200000", "5000", "5000", "1000000", "500000", "20000", []hv1.HealthDebtEntry{
		debtEntry("h", "20000", true),
	})
	result := hv1.Run(input)
	if result.Score > 100 {
		t.Fatalf("score must be clamped to 100, got %d", result.Score)
	}
}

func TestRun_ScoreClampedTo0(t *testing.T) {
	t.Parallel()
	// Max penalties: zero income (EMI present), no buffer, many debts, high unsecured
	input := makeInput("0", "0", "0", "0", "0", "50000", []hv1.HealthDebtEntry{
		debtEntry("d1", "10000", false),
		debtEntry("d2", "10000", false),
		debtEntry("d3", "10000", false),
		debtEntry("d4", "10000", false),
		debtEntry("d5", "10000", false),
	})
	result := hv1.Run(input)
	if result.Score < 0 {
		t.Fatalf("score must be clamped to >= 0, got %d", result.Score)
	}
	hv1.AssertScoreInRange(result)
}

func TestDebtCount_AggregateOnly(t *testing.T) {
	t.Parallel()
	n := 4
	input := hv1.HealthScoreInput{
		MonthlyIncome:   testDec("100000"),
		TotalMonthlyEMI: testDec("10000"),
		ActiveDebtCount: &n,
	}
	if hv1.DebtCount(input) != 4 {
		t.Fatalf("expected debt count 4 from ActiveDebtCount, got %d", hv1.DebtCount(input))
	}
}

func TestComputeUnsecuredRatio_AggregateOnly(t *testing.T) {
	t.Parallel()
	r := decimal.RequireFromString("0.55")
	input := hv1.HealthScoreInput{
		MonthlyIncome:     testDec("100000"),
		TotalMonthlyEMI:   testDec("20000"),
		UnsecuredEmiRatio: &r,
	}
	got := hv1.ComputeUnsecuredRatio(input)
	if !got.Equal(r) {
		t.Fatalf("expected aggregate ratio 0.55, got %s", got.String())
	}
}

func TestComputeUnsecuredRatio_AggregateOnly_ClampedHigh(t *testing.T) {
	t.Parallel()
	r := decimal.RequireFromString("1.5")
	input := hv1.HealthScoreInput{
		UnsecuredEmiRatio: &r,
	}
	got := hv1.ComputeUnsecuredRatio(input)
	one := decimal.RequireFromString("1")
	if !got.Equal(one) {
		t.Fatalf("expected ratio clamped to 1, got %s", got.String())
	}
}
