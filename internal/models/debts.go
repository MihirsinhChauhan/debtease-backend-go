package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateDebtRequest struct {
	Name               string          `json:"name" validate:"required"`
	Type               string          `json:"type" validate:"required,oneof=credit_card personal_loan student_loan mortgage medical other home_loan car_loan consumer_emi bnpl"`
	Lender             *string         `json:"lender,omitempty"`
	Principal          decimal.Decimal `json:"principal" validate:"required"`
	OutstandingBalance decimal.Decimal `json:"outstanding_balance" validate:"required"`
	InterestRate       decimal.Decimal `json:"interest_rate" validate:"gte=0"`
	MinPayment         decimal.Decimal `json:"min_payment" validate:"required,gt=0"`
	DueDate            time.Time       `json:"due_date" validate:"required"` // YYYY-MM-DD
	PaymentFrequency   *string         `json:"payment_frequency,omitempty" validate:"omitempty,oneof=monthly biweekly weekly"`
	DebtTaken          *time.Time      `json:"debt_taken,omitempty"`
	Priority           *int16          `json:"priority,omitempty"`
	Notes              *string         `json:"notes,omitempty"`

	TreatmentClassification *string `json:"treatment_classification,omitempty"`

	// Credit Card specific fields (required if type=credit_card)
	BillingDay        *int32           `json:"billing_day,omitempty" validate:"omitempty,min=1,max=31"`
	GracePeriodDays   *int32           `json:"grace_period_days,omitempty" validate:"omitempty,min=0"`
	MinPaymentPercent *decimal.Decimal `json:"min_payment_percent,omitempty" validate:"omitempty,gt=0"`

	// Loan / installment fields (optional — normalized to monthly_obligation + remaining_months)
	PaymentDueDay   *int32           `json:"payment_due_day,omitempty" validate:"omitempty,min=1,max=31"`
	EMIAmount       *decimal.Decimal `json:"emi_amount,omitempty"`
	TenureMonths    *int32           `json:"tenure_months,omitempty" validate:"omitempty,gt=0"`
	MoratoriumUntil *time.Time       `json:"moratorium_until,omitempty"`
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
	DebtHealthScore    *decimal.Decimal `json:"debt_health_score,omitempty"`
	Priority           *int16           `json:"priority,omitempty"`
	Notes              *string          `json:"notes,omitempty"`
	AccruedInterest    decimal.Decimal  `json:"accrued_interest"`
	LastAccrualDate    time.Time        `json:"last_accrual_date"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`

	InstrumentType          string          `json:"instrument_type"`
	InterestModel           string          `json:"interest_model"`
	MonthlyObligation       decimal.Decimal `json:"monthly_obligation"`
	MinimumObligation       decimal.Decimal `json:"minimum_obligation"`
	RemainingMonths         *int32          `json:"remaining_months,omitempty"`
	DueDay                  *int32          `json:"due_day,omitempty"`
	IsSecured               bool            `json:"is_secured"`
	IsRevolving             bool            `json:"is_revolving"`
	IsShortTerm             bool            `json:"is_short_term"`
	IsInterestBearing       bool            `json:"is_interest_bearing"`
	FragmentationWeight     decimal.Decimal `json:"fragmentation_weight"`
	StressPriority          int16           `json:"stress_priority"`
	TreatmentClassification *string         `json:"treatment_classification,omitempty"`
	EngineVersion           string          `json:"engine_version"`
	PolicyVersion           string          `json:"policy_version"`

	PaymentDueDay   *int32              `json:"payment_due_day,omitempty"`
	MoratoriumUntil *time.Time          `json:"moratorium_until,omitempty"`
	InMoratorium    *bool               `json:"in_moratorium,omitempty"`
	CreditCard      *CreditCardResponse `json:"credit_card,omitempty"`
}

type CreditCardResponse struct {
	BillingDay        int32            `json:"billing_day"`
	GracePeriodDays   int32            `json:"grace_period_days"`
	MinPaymentPercent decimal.Decimal  `json:"min_payment_percent"`
	LastStatementDate *time.Time       `json:"last_statement_date,omitempty"`
	PaymentDueDate    *time.Time       `json:"payment_due_date,omitempty"`
	BilledBalance     decimal.Decimal  `json:"billed_balance"`
	UnbilledBalance   decimal.Decimal  `json:"unbilled_balance"`
	InGracePeriod     bool             `json:"in_grace_period"`
	MinimumDue        *decimal.Decimal `json:"minimum_due,omitempty"`
}
