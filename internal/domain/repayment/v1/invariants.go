package repayment_v1

import "math"

const eps = 1e-2

func AssertMoneyConservation(summary PortfolioSummary) {
	if math.Abs(summary.TotalPaid-(summary.TotalInterest+summary.TotalPrincipal)) > eps {
		panic("money conservation violated")
	}
}

func AssertNoNegativeBalances(schedule map[int][]EMIRecord) {
	for _, records := range schedule {
		for _, r := range records {
			if r.ClosingBalance < -eps {
				panic("negative balance detected")
			}
		}
	}
}
