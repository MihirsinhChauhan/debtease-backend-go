package repayment_v1

import "testing"
import "DebtEase/internal/domain/repayment/v1"

func TestPortfolioScheduleAggregatesLoans(t *testing.T) {
	loans := []repayment_v1.Loan{
		{ID: "HL", Principal: 1000000, AnnualRate: 9, TenureMonths: 120},
		{ID: "PL", Principal: 300000, AnnualRate: 12, TenureMonths: 36},
	}

	portfolio := repayment_v1.GeneratePortfolioSchedule(loans)

	if len(portfolio) == 0 {
		t.Fatal("portfolio schedule empty")
	}

	month1 := portfolio[1]
	if len(month1) != 2 {
		t.Fatalf("expected 2 loan records in month 1, got %d", len(month1))
	}
}
