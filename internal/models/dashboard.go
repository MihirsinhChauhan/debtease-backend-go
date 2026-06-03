package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DashboardResponse struct {
	// Total Outstanding Debt
	TotalOutstanding decimal.Decimal `json:"total_outstanding"`

	// Repayment Progress (for ring chart)
	RepaymentProgress RepaymentProgress `json:"repayment_progress"`

	// Debt Health
	DebtHealth DebtHealth `json:"debt_health"`

	// OnboardingHealth is recomputed from the latest financial_snapshots row (health_v1), when present.
	OnboardingHealth *OnboardingHealthSnapshot `json:"onboarding_health,omitempty"`

	// IncomeStressPreview shows health status at current income and under −10% / −20% shocks (health_v1).
	IncomeStressPreview *IncomeStressPreview `json:"income_stress_preview,omitempty"`

	// FinancialSnapshotForSimulation matches POST /api/v1/snapshot/simulate-loan "snapshot" body (when user has a linked snapshot).
	FinancialSnapshotForSimulation *FinancialSnapshotForSimulation `json:"financial_snapshot_for_simulation,omitempty"`

	// Expected Debt Free
	ExpectedDebtFree *time.Time `json:"expected_debt_free,omitempty"`
	MonthsToDebtFree *int32     `json:"months_to_debt_free,omitempty"`

	// Monthly Payment
	MonthlyMinimumPayment decimal.Decimal `json:"monthly_minimum_payment"`

	// Average Interest
	AverageInterestRate decimal.Decimal `json:"average_interest_rate"`

	// Upcoming EMIs
	UpcomingEMIs []UpcomingEMIResponse `json:"upcoming_emis"`

	// Payment History Summary
	PaymentHistory PaymentHistorySummary `json:"payment_history"`
}

// FinancialSnapshotForSimulation is aggregate onboarding inputs for loan impact simulation (aligned with SnapshotInlineForSim in the webapp API).
type FinancialSnapshotForSimulation struct {
	MonthlyIncome     decimal.Decimal  `json:"monthly_income"`
	FixedExpenses     decimal.Decimal  `json:"fixed_expenses"`
	VariableExpenses  *decimal.Decimal `json:"variable_expenses,omitempty"`
	TotalMonthlyEMI   decimal.Decimal  `json:"total_monthly_emi"`
	SavingsCash       *decimal.Decimal `json:"savings_cash,omitempty"`
	InvestmentsValue  *decimal.Decimal `json:"investments_value,omitempty"`
	ActiveDebtCount   int              `json:"active_debt_count"`
	UnsecuredEmiRatio *decimal.Decimal `json:"unsecured_emi_ratio,omitempty"`
}

// IncomeStressPreview is health status under income shocks for dashboard resilience card.
type IncomeStressPreview struct {
	CurrentStatus string `json:"current_status"`
	Shock10Status string `json:"shock_10_status"`
	Shock20Status string `json:"shock_20_status"`
	PrimaryReason string `json:"primary_reason"`
}

// OnboardingHealthSnapshot is aggregate onboarding debt health (distinct from portfolio-weighted DebtHealth).
type OnboardingHealthSnapshot struct {
	SnapshotID        uuid.UUID       `json:"snapshot_id"`
	Status            string          `json:"status"`
	Score             int             `json:"score"`
	Drivers           []string        `json:"drivers"`
	EMIRatio          decimal.Decimal `json:"emi_ratio"`
	ShockBufferMonths decimal.Decimal `json:"shock_buffer_months"`
	EngineVersion     string          `json:"engine_version"`
	CapturedAt        time.Time       `json:"captured_at"`
}

type RepaymentProgress struct {
	TotalPrincipal      decimal.Decimal `json:"total_principal"`
	TotalPaid           decimal.Decimal `json:"total_paid"`
	RepaymentPercentage decimal.Decimal `json:"repayment_percentage"`
}

type DebtHealth struct {
	OverallHealthScore decimal.Decimal `json:"overall_health_score"`
	TotalActiveDebts   int64           `json:"total_active_debts"`
	OverdueCount       int64           `json:"overdue_count"`
	HighRiskCount      int64           `json:"high_risk_count"`
}

type UpcomingEMIResponse struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	Type               string           `json:"type"`
	Lender             *string          `json:"lender,omitempty"`
	OutstandingBalance decimal.Decimal  `json:"outstanding_balance"`
	AccruedInterest    decimal.Decimal  `json:"accrued_interest"`
	PaymentAmount      decimal.Decimal  `json:"payment_amount"`
	DueDate            string           `json:"due_date"` // ISO date format
	PaymentDueDay      *int32           `json:"payment_due_day,omitempty"`
	IsOverdue          bool             `json:"is_overdue"`
	LateFees           *decimal.Decimal `json:"late_fees,omitempty"`
	DaysUntilDue       int32            `json:"days_until_due"`
}

type PaymentHistorySummary struct {
	PaidLast30Days     decimal.Decimal `json:"paid_last_30_days"`
	PaidLast90Days     decimal.Decimal `json:"paid_last_90_days"`
	PaidLastYear       decimal.Decimal `json:"paid_last_year"`
	PaymentsLast30Days int64           `json:"payments_last_30_days"`
}
