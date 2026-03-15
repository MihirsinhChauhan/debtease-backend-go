package health_v1

import "github.com/shopspring/decimal"

// HealthStatus is the user-facing status derived from the score (doc §10).
type HealthStatus string

const (
	StatusStable    HealthStatus = "Stable"    // 80–100
	StatusSensitive HealthStatus = "Sensitive" // 60–79
	StatusFragile   HealthStatus = "Fragile"   // <60
)

// HealthDebtEntry is a minimal view of a debt for health score (EMI + secured flag).
// Handler/service layers map database or snapshot data into this type.
type HealthDebtEntry struct {
	ID      string
	EMI     decimal.Decimal
	Secured bool
}

// HealthScoreInput is the input to the health score engine.
// Built from financial snapshot aggregates + debt portfolio.
type HealthScoreInput struct {
	// From snapshot
	MonthlyIncome     decimal.Decimal
	FixedExpenses     decimal.Decimal
	VariableExpenses  decimal.Decimal
	SavingsCash       decimal.Decimal
	InvestmentsValue  decimal.Decimal
	TotalMonthlyEMI   decimal.Decimal
	// From debt portfolio (for fragmentation + unsecured ratio)
	Debts []HealthDebtEntry
}

// HealthScoreResult is the deterministic output of the health score engine.
type HealthScoreResult struct {
	Score         int           // 0–100
	Status        HealthStatus
	Drivers       []string      // explainability
	EngineVersion string
}
