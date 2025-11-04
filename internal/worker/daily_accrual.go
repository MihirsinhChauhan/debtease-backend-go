package worker

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

func StartDailyAccrual(ctx context.Context, q *database.Queries, tick time.Duration) {
	if tick <= 0 {
		tick = 24 * time.Hour
	}
	ticker := time.NewTicker(tick)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				processAccrual(ctx, q, t)
			}
		}
	}()
}

func processAccrual(ctx context.Context, q *database.Queries, now time.Time) {
	// Use date-only for comparisons
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	debts, err := q.ListAllActiveDebts(ctx)
	if err != nil {
		logger.Logger.Errorw("ListAllActiveDebts failed", "error", err)
		return
	}
	for _, d := range debts {
		last := today
		ld := d.LastAccrualDate
		last = time.Date(ld.Year(), ld.Month(), ld.Day(), 0, 0, 0, 0, time.UTC)
		days := int(today.Sub(last).Hours() / 24)
		if days <= 0 {
			continue
		}
		dailyRate := d.InterestRate.Div(decimal.NewFromInt(36500))
		delta := d.OutstandingBalance.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days)))
		if err := q.AccrueInterest(ctx, database.AccrueInterestParams{ID: d.ID, AccruedInterest: delta, LastAccrualDate: today}); err != nil {
			logger.Logger.Errorw("AccrueInterest failed", "debt_id", d.ID, "error", err)
		}
	}
}
