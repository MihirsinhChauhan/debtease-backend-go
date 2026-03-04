package stress_v1

import "github.com/shopspring/decimal"

const EngineVersion = "v1.0.0"

// StressPolicyThresholds holds configurable thresholds and shock parameters.
// All ratios are 0–1 where applicable.
type StressPolicyThresholds struct {
	// IncomeShockAlphas are multiplicative shocks applied to income (e.g. 0.9, 0.8).
	IncomeShockAlphas []decimal.Decimal

	// UnsecuredRateShockBeta is the proportional EMI bump applied to unsecured loans
	// (excluding BNPL/0% EMI) when simulating interest rate shock (e.g. 0.10 for +10%).
	UnsecuredRateShockBeta decimal.Decimal

	// ImmediateStressMaxMonths is the upper bound (exclusive) for immediate stress band.
	ImmediateStressMaxMonths decimal.Decimal
	// ShortBandMaxMonths is the upper bound (exclusive) for short survivability band.
	ShortBandMaxMonths decimal.Decimal
	// ModerateBandMaxMonths is the upper bound (exclusive) for moderate survivability band.
	ModerateBandMaxMonths decimal.Decimal

	// ResilientMinMonths is the minimum buffer months required to be considered RESILIENT.
	ResilientMinMonths decimal.Decimal

	// UnsecuredExposureHigh is the threshold above which unsecured EMI share is considered high.
	UnsecuredExposureHigh decimal.Decimal
	// ShortTermRigidityHigh is the threshold above which short-term unsecured EMI share is high.
	ShortTermRigidityHigh decimal.Decimal
}

// DefaultStressPolicy returns India v1 defaults derived from the domain spec.
func DefaultStressPolicy() StressPolicyThresholds {
	return StressPolicyThresholds{
		IncomeShockAlphas: []decimal.Decimal{
			decimal.RequireFromString("0.9"),
			decimal.RequireFromString("0.8"),
		},
		UnsecuredRateShockBeta:   decimal.RequireFromString("0.10"),
		ImmediateStressMaxMonths: decimal.RequireFromString("1"),
		ShortBandMaxMonths:       decimal.RequireFromString("3"),
		ModerateBandMaxMonths:    decimal.RequireFromString("6"),
		ResilientMinMonths:       decimal.RequireFromString("2"),
		UnsecuredExposureHigh:    decimal.RequireFromString("0.5"),
		ShortTermRigidityHigh:    decimal.RequireFromString("0.3"),
	}
}

// Policy documents explicit assumptions for this domain (versioned, auditable).
var Policy = []string{
	"Stress & Shock Simulation evaluates debt system fragility under deterministic shocks; it is not a predictor or advisor.",
	"All monetary values and ratios use decimal.Decimal to avoid precision errors.",
	"Income shocks are modeled as multiplicative factors on monthly net income (e.g. 0.9, 0.8).",
	"Interest rate shocks are applied only to unsecured, non-BNPL loans via a proportional EMI bump.",
	"Flexibility buffer = Income - Fixed_Obligations - Total_EMI' after shocks.",
	"Survivability bands are coarse: IMMEDIATE_STRESS (<1 month), SHORT (1–3), MODERATE (3–6), STABLE (>=6).",
	"Stress levels are qualitative: FRAGILE (negative buffer), SENSITIVE (thin positive buffer), RESILIENT (comfortable buffer).",
	"Failure ordering is deterministic: BNPL / 0% EMI → short-term unsecured → long-term unsecured → secured.",
	"Same inputs and policy parameters always produce the same StressOutcome (deterministic engine).",
	"No probabilistic language, no scores, and no default probabilities are produced by this domain.",
}

