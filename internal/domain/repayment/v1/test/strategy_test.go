package repayment_v1

import "testing"
import "DebtEase/internal/domain/repayment/v1"
import "math"
func TestAvalancheBeatsBaseline(t *testing.T) {
	loans := []repayment_v1.Loan{
		{ID: "HL", Principal: 2000000, AnnualRate: 8.5, TenureMonths: 240},
		{ID: "PL", Principal: 500000, AnnualRate: 13, TenureMonths: 60},
	}

	baseline := repayment_v1.GeneratePortfolioSchedule(loans)
	avalanche := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategyAvalanche, 10000)

	baseSummary := repayment_v1.SummarizePortfolio(baseline)
	avaSummary := repayment_v1.SummarizePortfolio(avalanche)

	if avaSummary.TotalInterest >= baseSummary.TotalInterest {
		t.Fatal("avalanche should reduce total interest")
	}
}

func TestStrategyDoesNotRewritePast(t *testing.T) {
	loans := []repayment_v1.Loan{
		{ID: "PL", Principal: 500000, AnnualRate: 12, TenureMonths: 60},
	}

	baseline := repayment_v1.GeneratePortfolioSchedule(loans)
	switched := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategySnowball, 5000)

	for month := 1; month <= 6; month++ {
		baseEMI := sumEMI(baseline[month])
		switchEMI := sumEMI(switched[month])

		if math.Abs(baseEMI-switchEMI) > 1e-2 {
			t.Fatalf("EMI mismatch in month %d", month)
		}
	}
}

func TestAggressivePrepaymentNeverGoesNegative(t *testing.T) {
	loans := []repayment_v1.Loan{
		{ID: "CL", Principal: 300000, AnnualRate: 10, TenureMonths: 36},
	}

	schedule := repayment_v1.ApplyStrategy(loans, repayment_v1.StrategyAvalanche, 50000)

	repayment_v1.AssertNoNegativeBalances(schedule)
}

func sumEMI(records []repayment_v1.EMIRecord) float64 {
	var total float64
	for _, r := range records {
		total += r.EMI
	}
	return total
}

