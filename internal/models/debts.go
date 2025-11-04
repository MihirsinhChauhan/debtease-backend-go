package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateDebtRequest struct {
	Name               string          `json:"name" validate:"required"`
	Type               string          `json:"type" validate:"required,oneof=credit_card personal_loan student_loan mortgage medical other"`
	Lender             *string         `json:"lender,omitempty"`
	Principal          decimal.Decimal `json:"principal" validate:"required"`
	OutstandingBalance decimal.Decimal `json:"outstanding_balance" validate:"required"`
	InterestRate       decimal.Decimal `json:"interest_rate" validate:"required,gt=0"`
	MinPayment         decimal.Decimal `json:"min_payment" validate:"required,gt=0"`
	DueDate            time.Time       `json:"due_date" validate:"required"` // YYYY-MM-DD
	PaymentFrequency   *string         `json:"payment_frequency,omitempty" validate:"omitempty,oneof=monthly biweekly weekly"`
	DebtTaken          *time.Time      `json:"debt_taken,omitempty"`
	RiskLevel          *string         `json:"risk_level,omitempty" validate:"omitempty,oneof=low medium high"`
	Priority           *int16          `json:"priority,omitempty"`
	Notes              *string         `json:"notes,omitempty"`
}

type UpdateDebtBalanceRequest struct {
	OutstandingBalance decimal.Decimal `json:"outstanding_balance" validate:"required"`
	IsOverdue          bool            `json:"is_overdue"`
}

type UpdateDebtHealthRequest struct {
	DebtHealthScore *decimal.Decimal `json:"debt_health_score" validate:"required"`
}
type DebtResponse struct {
	ID                 string           `json:"id"`
	UserID             string           `json:"user_id"`
	Name               string           `json:"name"`
	Type               string           `json:"type"`
	Lender             *string          `json:"lender,omitempty"`
	Principal          decimal.Decimal  `json:"principal"`
	OutstandingBalance decimal.Decimal  `json:"outstanding_balance"`
	InterestRate       decimal.Decimal  `json:"interest_rate"`
	MinPayment         decimal.Decimal  `json:"min_payment"`
	DueDate            string           `json:"due_date"` // ISO date only
	PaymentFrequency   *string          `json:"payment_frequency,omitempty"`
	IsOverdue          bool             `json:"is_overdue"`
	LateFees           *decimal.Decimal `json:"late_fees,omitempty"`
	OtherCharges       *string          `json:"other_charges,omitempty"`
	DebtTaken          *time.Time       `json:"debt_taken,omitempty"`
	RiskLevel          *string          `json:"risk_level,omitempty"`
	DebtHealthScore    *decimal.Decimal `json:"debt_health_score,omitempty"`
	Priority           *int16           `json:"priority,omitempty"`
	Notes              *string          `json:"notes,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}
