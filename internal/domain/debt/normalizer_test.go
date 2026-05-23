package debt

import (
	"strings"
	"testing"
	"time"

	"DebtEase/internal/database"

	"github.com/shopspring/decimal"
)

func TestNormalizeCreate_creditCardRevolving(t *testing.T) {
	p := decimal.RequireFromString("5000")
	d := NormalizeCreate(CreateObligationInput{
		DebtType:     database.DebtTypeCreditCard,
		InterestRate: decimal.RequireFromString("24"),
		MinPayment:   p,
		Principal:    decimal.RequireFromString("10000"),
		DueDate:      mustDate(t, "2026-01-15"),
	})
	if d.InstrumentType != database.DebtInstrumentTypeCreditCard {
		t.Fatalf("instrument: %s", d.InstrumentType)
	}
	if d.InterestModel != database.InterestModelRevolving {
		t.Fatalf("interest model: %s", d.InterestModel)
	}
	if !d.IsRevolving || d.IsSecured {
		t.Fatalf("flags secured=%v revolving=%v", d.IsSecured, d.IsRevolving)
	}
	if !d.MonthlyObligation.Equal(p) {
		t.Fatalf("monthly %v", d.MonthlyObligation)
	}
}

func TestNormalizeCreate_bnplShortTermZeroInterest(t *testing.T) {
	min := decimal.RequireFromString("2500")
	d := NormalizeCreate(CreateObligationInput{
		DebtType:     database.DebtTypeBnpl,
		InterestRate: decimal.Zero,
		MinPayment:   min,
		Principal:    decimal.RequireFromString("7500"),
		TenureMonths: ptrInt32(3),
		DueDate:      mustDate(t, "2026-02-01"),
	})
	if d.InstrumentType != database.DebtInstrumentTypeBnpl {
		t.Fatalf("instrument: %s", d.InstrumentType)
	}
	if !d.IsShortTerm || d.IsInterestBearing {
		t.Fatalf("bnpl flags short=%v interest=%v", d.IsShortTerm, d.IsInterestBearing)
	}
	if !d.MonthlyObligation.Equal(min) {
		t.Fatalf("monthly %v", d.MonthlyObligation)
	}
}

func ptrInt32(v int32) *int32 {
	return &v
}

func mustDate(t *testing.T, ymd string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", strings.TrimSpace(ymd))
	if err != nil {
		t.Fatalf("date: %v", err)
	}
	return d
}
