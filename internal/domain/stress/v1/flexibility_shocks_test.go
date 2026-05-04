package stress_v1

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFlexibilityBuffersForIncomeShocks(t *testing.T) {
	t.Parallel()
	income := decimal.RequireFromString("100000")
	fixed := decimal.RequireFromString("20000")
	emi := decimal.RequireFromString("30000")
	cur, s10, s20 := FlexibilityBuffersForIncomeShocks(income, fixed, emi)
	// 100k - 20k - 30k = 50k
	if !cur.Equal(decimal.RequireFromString("50000")) {
		t.Fatalf("current want 50000 got %s", cur.String())
	}
	// 90k - 20k - 30k = 40k
	if !s10.Equal(decimal.RequireFromString("40000")) {
		t.Fatalf("shock10 want 40000 got %s", s10.String())
	}
	// 80k - 20k - 30k = 30k
	if !s20.Equal(decimal.RequireFromString("30000")) {
		t.Fatalf("shock20 want 30000 got %s", s20.String())
	}
}

func TestWorstSurvivabilityMonthsAcrossIncomeShocks(t *testing.T) {
	t.Parallel()
	income := decimal.RequireFromString("100000")
	fixed := decimal.RequireFromString("20000")
	emi := decimal.RequireFromString("30000")
	w := WorstSurvivabilityMonthsAcrossIncomeShocks(income, fixed, emi)
	// months = buffer/EMI: 50k/30k, 40k/30k, 30k/30k=1 -> min is 1
	exp := decimal.NewFromInt(1)
	if !w.Equal(exp) {
		t.Fatalf("worst months want %s got %s", exp.String(), w.String())
	}
}
