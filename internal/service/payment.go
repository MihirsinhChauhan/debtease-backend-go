package service

import (
	"DebtEase/internal/database"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	DB *sql.DB
	Q  *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{DB: db, Q: database.New(db)}
}

type RecordPaymentParams struct {
	UserID        uuid.UUID
	DebtID        uuid.UUID
	Amount        decimal.Decimal
	PaymentDate   time.Time
	Type          database.PaymentType
	Notes         *string
	PaymentMethod *string
}

// accrueInterestUpTo calculates and persists interest accrual up to and including the given date.
func (s *Service) accrueInterestUpTo(ctx context.Context, q *database.Queries, debt database.Debt, upTo time.Time) error {
	// guard: if last_accrual_date is null, start from upTo (no accrual before)
	lastDate := upTo
	lastDate = debt.LastAccrualDate
	// compute whole days difference based on date only
	start := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(upTo.Year(), upTo.Month(), upTo.Day(), 0, 0, 0, 0, time.UTC)
	days := int(end.Sub(start).Hours() / 24)
	if days <= 0 {
		return nil
	}

	// daily_rate = interest_rate / 100 / 365
	dailyRate := debt.InterestRate.Div(decimal.NewFromInt(36500))
	accrued := debt.OutstandingBalance.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days)))

	return q.AccrueInterest(ctx, database.AccrueInterestParams{
		ID:              debt.ID,
		AccruedInterest: accrued,
		LastAccrualDate: end,
	})
}

// RecordPayment accrues interest to payment date, splits amount into interest/principal, records payment and updates debt.
func (s *Service) RecordPayment(ctx context.Context, arg RecordPaymentParams) (database.PaymentsHistory, database.Debt, error) {
	var created database.PaymentsHistory
	var updatedDebt database.Debt

	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return created, updatedDebt, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	q := s.Q.WithTx(tx)

	debt, err := q.GetDebtForUpdate(ctx, database.GetDebtForUpdateParams{ID: arg.DebtID, UserID: arg.UserID})
	if err != nil {
		return created, updatedDebt, err
	}
	// Accrue to payment date first
	if err = s.accrueInterestUpTo(ctx, q, debt, arg.PaymentDate); err != nil {
		return created, updatedDebt, err
	}
	// Refresh debt after accrual
	debt, err = q.GetDebtForUpdate(ctx, database.GetDebtForUpdateParams{ID: arg.DebtID, UserID: arg.UserID})
	if err != nil {
		return created, updatedDebt, err
	}

	// Calculate total due
	var accrued decimal.Decimal
	accrued = debt.AccruedInterest
	totalDue := debt.OutstandingBalance.Add(accrued)
	amount := arg.Amount
	if amount.GreaterThan(totalDue) {
		amount = totalDue
	}

	// Split
	interestPaid := amount
	if interestPaid.GreaterThan(accrued) {
		interestPaid = accrued
	}
	principalPaid := amount.Sub(interestPaid)

	// New balances
	newAccrued := accrued.Sub(interestPaid)
	if newAccrued.IsNegative() {
		newAccrued = decimal.Zero
	}
	newOutstanding := debt.OutstandingBalance.Sub(principalPaid)
	if newOutstanding.IsNegative() {
		newOutstanding = decimal.Zero
	}

	// Record payment
	created, err = q.RecordPayment(ctx, database.RecordPaymentParams{
		DebtID:           arg.DebtID,
		Amount:           amount,
		PaymentDate:      arg.PaymentDate,
		Type:             arg.Type,
		Notes:            nullString(arg.Notes),
		PrincipalApplied: principalPaid,
		InterestApplied:  interestPaid,
		PaymentMethod:    nullString(arg.PaymentMethod),
	})
	if err != nil {
		return created, updatedDebt, err
	}

	// Persist debt updates
	err = q.UpdateDebtAfterPayment(ctx, database.UpdateDebtAfterPaymentParams{
		ID:                 arg.DebtID,
		OutstandingBalance: newOutstanding,
		AccruedInterest:    newAccrued,
		LastAccrualDate:    arg.PaymentDate,
	})
	if err != nil {
		return created, updatedDebt, err
	}
	updatedDebt, err = q.GetDebtByID(ctx, database.GetDebtByIDParams{ID: arg.DebtID, UserID: arg.UserID})
	return created, updatedDebt, err
}

func nullString(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *p, Valid: *p != ""}
}
