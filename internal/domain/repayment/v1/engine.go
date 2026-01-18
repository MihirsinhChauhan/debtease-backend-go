package repayment_v1

import "math"

func CalculateEMI(principal float64, annualRate float64, tenureMonths int) float64 {
	r := annualRate / 12 / 100
	if r == 0 {
		return round(principal/float64(tenureMonths), 2)
	}

	emi := principal * r * math.Pow(1+r, float64(tenureMonths)) /
		(math.Pow(1+r, float64(tenureMonths)) - 1)

	return round(emi, 2)
}

func GenerateLoanSchedule(loan Loan) []EMIRecord {
	r := loan.AnnualRate / 12 / 100
	emi := CalculateEMI(loan.Principal, loan.AnnualRate, loan.TenureMonths)
	balance := loan.Principal

	var schedule []EMIRecord
	month := 1

	// Moratorium
	for i := 0; i < loan.MoratoriumMonths; i++ {
		interest := round(balance*r, 2)
		balance = round(balance+interest, 2)

		schedule = append(schedule, EMIRecord{
			LoanID:         loan.ID,
			Month:          month,
			OpeningBalance: balance - interest,
			EMI:            0,
			Interest:       interest,
			Principal:      0,
			ClosingBalance: balance,
		})
		month++
	}

	// Repayment
	for balance > 0 {
		opening := balance
		interest := round(balance*r, 2)

		effectiveEMI := math.Min(emi, round(balance+interest, 2))
		principal := round(effectiveEMI-interest, 2)
		balance = round(balance-principal, 2)

		if balance < 0 {
			balance = 0
		}

		schedule = append(schedule, EMIRecord{
			LoanID:         loan.ID,
			Month:          month,
			OpeningBalance: opening,
			EMI:            effectiveEMI,
			Interest:       interest,
			Principal:      principal,
			ClosingBalance: balance,
		})

		month++
	}

	return schedule
}

func round(v float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(v*p) / p
}
