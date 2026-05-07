package health_v1

import (
	"testing"

	hv1 "DebtEase/internal/domain/health/v1"
)

// Golden: Stable — high income, low EMI, good buffer, few debts, low unsecured.
func TestGolden_Stable(t *testing.T) {
	t.Parallel()
	input := makeInput("100000", "10000", "15000", "200000", "100000", "20000", []hv1.HealthDebtEntry{
		debtEntry("hl", "20000", true),
	})
	result := hv1.Run(input)
	if result.Status != hv1.StatusStable {
		t.Fatalf("golden: expected Stable, got %s (score %d)", result.Status, result.Score)
	}
	if result.Score < 80 || result.Score > 100 {
		t.Fatalf("golden: expected score 80–100, got %d", result.Score)
	}
	if result.EngineVersion != hv1.EngineVersion {
		t.Fatalf("golden: engine version must be %s, got %s", hv1.EngineVersion, result.EngineVersion)
	}
}

// Golden: Sensitive — moderate EMI ratio, buffer in 3–6mo, 3 debts, unsecured in 40–60%.
func TestGolden_Sensitive(t *testing.T) {
	t.Parallel()
	// EMI 35k/100k = 35% -> -10; buffer 200k/(10k+15k+35k)=3.33mo -> -5; 3 debts -> -5; unsecured 40% -> -5. Total 25, score 75.
	input := makeInput("100000", "10000", "15000", "200000", "0", "35000", []hv1.HealthDebtEntry{
		debtEntry("u1", "14000", false),
		debtEntry("u2", "10000", false),
		debtEntry("s1", "11000", true),
	})
	result := hv1.Run(input)
	if result.Status != hv1.StatusSensitive {
		t.Fatalf("golden: expected Sensitive, got %s (score %d)", result.Status, result.Score)
	}
	if result.Score < 60 || result.Score >= 80 {
		t.Fatalf("golden: expected score 60–79, got %d", result.Score)
	}
	if result.EngineVersion != hv1.EngineVersion {
		t.Fatalf("golden: engine version must be %s, got %s", hv1.EngineVersion, result.EngineVersion)
	}
}

// Golden: Fragile — high EMI, low buffer, multiple debts, high unsecured.
func TestGolden_Fragile(t *testing.T) {
	t.Parallel()
	// EMI 50k/60k > 45% -> -25; buffer low; 5 debts -> -15; all unsecured -> -10. Score can go below 60.
	input := makeInput("60000", "10000", "10000", "5000", "0", "50000", []hv1.HealthDebtEntry{
		debtEntry("d1", "12000", false),
		debtEntry("d2", "12000", false),
		debtEntry("d3", "10000", false),
		debtEntry("d4", "8000", false),
		debtEntry("d5", "8000", false),
	})
	result := hv1.Run(input)
	if result.Status != hv1.StatusFragile {
		t.Fatalf("golden: expected Fragile, got %s (score %d)", result.Status, result.Score)
	}
	if result.Score >= 60 {
		t.Fatalf("golden: expected score < 60, got %d", result.Score)
	}
	if result.EngineVersion != hv1.EngineVersion {
		t.Fatalf("golden: engine version must be %s, got %s", hv1.EngineVersion, result.EngineVersion)
	}
}

// Golden: deterministic — same input gives same output.
func TestGolden_Deterministic(t *testing.T) {
	t.Parallel()
	input := makeInput("70000", "8000", "10000", "80000", "20000", "25000", []hv1.HealthDebtEntry{
		debtEntry("a", "15000", false),
		debtEntry("b", "10000", true),
	})
	hv1.AssertDeterministic(input)
}
