package repayment_v1


func SummarizePortfolio(schedule map[int][]EMIRecord) PortfolioSummary {
	var paid, interest, principal float64
	lastMonth := 0

	for m, records := range schedule {
		if m > lastMonth {
			lastMonth = m
		}
		for _, r := range records {
			paid += r.EMI
			interest += r.Interest
			principal += r.Principal
		}
	}

	return PortfolioSummary{
		Months:         lastMonth,
		TotalPaid:      round(paid, 2),
		TotalInterest:  round(interest, 2),
		TotalPrincipal: round(principal, 2),
	}
}
