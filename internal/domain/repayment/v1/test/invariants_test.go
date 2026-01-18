package repayment_v1

import "testing"
import "DebtEase/internal/domain/repayment/v1"
func TestMoneyConservation(t *testing.T) {
	loans := []repayment_v1.Loan{
		{ID: "HL", Principal: 1000000, AnnualRate: 9, TenureMonths: 120},
	}

	schedule := repayment_v1.GeneratePortfolioSchedule(loans)
	summary := repayment_v1.SummarizePortfolio(schedule)

	repayment_v1.AssertMoneyConservation(summary)
}

func TestNoNegativeBalancesInvariant(t *testing.T) {
	loans := []repayment_v1.Loan{
		{ID: "PL", Principal: 500000, AnnualRate: 12, TenureMonths: 60},
	}

	schedule := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategySnowball, 20000)

	repayment_v1.AssertNoNegativeBalances(schedule)
}
