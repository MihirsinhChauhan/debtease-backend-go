package repayment_v1

import "testing"
import "DebtEase/internal/domain/repayment/v1"
import "math"
func TestGoldenCase_SingleLoanBaseline(t *testing.T) {
	loan := repayment_v1.Loan{
		ID:           "HL",
		Principal:    1000000,
		AnnualRate:   9,
		TenureMonths: 120,
	}

	schedule := repayment_v1.GeneratePortfolioSchedule([]repayment_v1.Loan{loan})
	summary := repayment_v1.SummarizePortfolio(schedule)

	// 🔒 DO NOT CHANGE WITHOUT NEW ENGINE VERSION
	if summary.Months != 120 {
		t.Fatalf("months changed: %d", summary.Months)
	}
	const expectedInterest = 520109.10
	if math.Abs(summary.TotalInterest-expectedInterest) > 1 {
		t.Fatalf(
			"interest drift detected: expected %.2f got %.2f",
			expectedInterest,
			summary.TotalInterest,
		)
	}
}
