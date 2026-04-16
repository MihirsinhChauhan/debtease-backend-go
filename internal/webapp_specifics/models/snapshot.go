package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// SnapshotAnalyzeRequest is the onboarding financial snapshot (aggregate inputs).
type SnapshotAnalyzeRequest struct {
	MonthlyIncome     decimal.Decimal  `json:"monthly_income"`
	FixedExpenses     decimal.Decimal  `json:"fixed_expenses"`
	VariableExpenses  *decimal.Decimal `json:"variable_expenses,omitempty"`
	TotalMonthlyEMI   decimal.Decimal  `json:"total_monthly_emi"`
	SavingsCash       *decimal.Decimal `json:"savings_cash,omitempty"`
	InvestmentsValue  *decimal.Decimal `json:"investments_value,omitempty"`
	ActiveDebtCount   int              `json:"active_debt_count"`
	UnsecuredEmiRatio *decimal.Decimal `json:"unsecured_emi_ratio,omitempty"`
	EntryPoint        string           `json:"entry_point"`
	EngineVersion     string           `json:"engine_version"`
}

// SnapshotAnalyzeResponse is debt health from health_v1 plus persisted snapshot id.
type SnapshotAnalyzeResponse struct {
	Status            string          `json:"status"`
	Score             int             `json:"score"`
	Drivers           []string        `json:"drivers"`
	EMIRatio          decimal.Decimal `json:"emi_ratio"`
	ShockBufferMonths decimal.Decimal `json:"shock_buffer_months"`
	SnapshotID        uuid.UUID       `json:"snapshot_id"`
	EngineVersion     string          `json:"engine_version"`
}
