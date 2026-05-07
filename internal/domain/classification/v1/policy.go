package classification_v1

import "github.com/shopspring/decimal"

const EngineVersion = "v1.0.0"

// PolicyThresholds holds India v1 policy thresholds. All ratios are 0–1 except where noted.
type PolicyThresholds struct {
	MaxSafeEMIRatio       decimal.Decimal // e.g. 0.35 — above this is tight/fragile
	MaxFragmentationIndex decimal.Decimal // e.g. 4 — above is debt loop risk
	MinShockBufferMonths  decimal.Decimal // e.g. 1.5 — below is fragile
	MaxUnsecuredRatio     decimal.Decimal // e.g. 0.60 — above is overexposure
}

// DefaultPolicyThresholds returns India v1 defaults.
func DefaultPolicyThresholds() PolicyThresholds {
	return PolicyThresholds{
		MaxSafeEMIRatio:       decimal.RequireFromString("0.35"),
		MaxFragmentationIndex: decimal.NewFromInt(4),
		MinShockBufferMonths:  decimal.RequireFromString("1.5"),
		MaxUnsecuredRatio:     decimal.RequireFromString("0.60"),
	}
}

// Policy documents explicit assumptions for this domain (versioned, auditable).
var Policy = []string{
	"Debt classification is user-risk (financial fragility), not bank risk or default prediction.",
	"All monetary values and ratios use decimal.Decimal to avoid precision errors.",
	"EMI ratio = Total_Monthly_EMI / Monthly_Net_Income; < 30% manageable, 30–45% tight, > 45% fragile.",
	"Fragmentation index = Active_Debts + BNPL_Count*0.5; high value indicates cognitive/operational overload.",
	"Rigidity score = SUM(EMI * Remaining_Months) / Monthly_Income; higher means longer unavoidable commitment.",
	"Shock buffer months = (Income - Fixed_Obligations - EMIs) / EMIs; how many months of EMI user can absorb if income dips.",
	"Unsecured ratio = Unsecured_EMI / Total_EMI; high unsecured exposure amplifies downside risk.",
	"Mortgage is secured; personal_loan, credit_card, student_loan, medical, BNPL, LOC, other are unsecured (India v1).",
	"Classification is deterministic: same inputs always produce same output.",
	"No probabilistic language; no irreversible labels; user can override planner decisions.",
	"Policies are versioned and auditable; thresholds can be adjusted regionally without changing logic.",
}
