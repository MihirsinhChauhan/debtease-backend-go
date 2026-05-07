package health_v1

import "github.com/shopspring/decimal"

const EngineVersion = "v1.0.0"

const baseScore = 100

// HealthPolicyThresholds holds penalty band boundaries (doc §9.4).
// Ratios are 0–1 (e.g. 0.30 = 30%).
type HealthPolicyThresholds struct {
	// EMI ratio: <30% → 0, 30–45% → -10, >45% → -25
	EMIRatioLowBound  decimal.Decimal // 0.30
	EMIRatioHighBound decimal.Decimal // 0.45
	PenaltyEMIRatioMid int             // -10
	PenaltyEMIRatioHigh int            // -25

	// Shock buffer months: >6 → 0, 3–6 → -5, 1–3 → -15, <1 → -30
	ShockBufferMonths6  decimal.Decimal // 6
	ShockBufferMonths3  decimal.Decimal // 3
	ShockBufferMonths1  decimal.Decimal // 1
	PenaltyShockBufferMid   int         // -5
	PenaltyShockBufferLow   int         // -15
	PenaltyShockBufferNone  int         // -30

	// Fragmentation: ≤2 → 0, 3–4 → -5, ≥5 → -15
	FragmentationMax2   int // 2
	FragmentationMax4   int // 4
	PenaltyFragmentationMid  int // -5
	PenaltyFragmentationHigh int // -15

	// Unsecured ratio: <40% → 0, 40–60% → -5, >60% → -10
	UnsecuredRatioMid  decimal.Decimal // 0.40
	UnsecuredRatioHigh decimal.Decimal // 0.60
	PenaltyUnsecuredMid  int // -5
	PenaltyUnsecuredHigh int // -10
}

// DefaultHealthPolicy returns India v1 defaults from the domain spec.
func DefaultHealthPolicy() HealthPolicyThresholds {
	return HealthPolicyThresholds{
		EMIRatioLowBound:   decimal.RequireFromString("0.30"),
		EMIRatioHighBound:  decimal.RequireFromString("0.45"),
		PenaltyEMIRatioMid:  10,
		PenaltyEMIRatioHigh: 25,

		ShockBufferMonths6: decimal.RequireFromString("6"),
		ShockBufferMonths3: decimal.RequireFromString("3"),
		ShockBufferMonths1: decimal.RequireFromString("1"),
		PenaltyShockBufferMid:  5,
		PenaltyShockBufferLow:  15,
		PenaltyShockBufferNone: 30,

		FragmentationMax2: 2,
		FragmentationMax4: 4,
		PenaltyFragmentationMid:  5,
		PenaltyFragmentationHigh: 15,

		UnsecuredRatioMid:  decimal.RequireFromString("0.40"),
		UnsecuredRatioHigh: decimal.RequireFromString("0.60"),
		PenaltyUnsecuredMid:  5,
		PenaltyUnsecuredHigh: 10,
	}
}

// Policy documents explicit assumptions for this domain (versioned, auditable).
var Policy = []string{
	"Financial Health Score is a structural signal (0–100), not a default predictor.",
	"All monetary values and ratios use decimal.Decimal to avoid precision errors.",
	"Score = BASE_SCORE (100) minus penalties; result clamped to [0, 100].",
	"EMI ratio = Total_Monthly_EMI / Monthly_Income; <30% no penalty, 30–45% -10, >45% -25.",
	"Shock buffer = (Savings + Investments) / monthly_burn; monthly_burn = fixed + variable + EMI. >6mo 0, 3–6 -5, 1–3 -15, <1 -30.",
	"Fragmentation = count of active debts; ≤2 no penalty, 3–4 -5, ≥5 -15.",
	"Unsecured exposure = Unsecured_EMI / Total_EMI; <40% 0, 40–60% -5, >60% -10.",
	"Status: Stable (80–100), Sensitive (60–79), Fragile (<60).",
	"Same inputs and policy always produce the same score (deterministic).",
}
