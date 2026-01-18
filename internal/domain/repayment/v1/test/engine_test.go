package repayment_v1

import "testing"
import "DebtEase/internal/domain/repayment/v1"
func TestCalculateEMI_ZeroInterest(t *testing.T) {
	emi := repayment_v1.CalculateEMI(120000, 0, 12)
	if emi != 10000 {
		t.Fatalf("expected EMI 10000, got %.2f", emi)
	}
}

func TestLoanClosesCleanly(t *testing.T) {
	loan := repayment_v1.Loan{
		ID:           "HL",
		Principal:    1000000,
		AnnualRate:   9,
		TenureMonths: 120,
	}

	schedule := repayment_v1.GenerateLoanSchedule(loan)
	last := schedule[len(schedule)-1]

	if last.ClosingBalance != 0 {
		t.Fatalf("loan did not close, balance=%.2f", last.ClosingBalance)
	}
}

func TestMoratoriumAccruesInterest(t *testing.T) {
	loan := repayment_v1.Loan{
		ID:               "PL",
		Principal:        500000,
		AnnualRate:       12,
		TenureMonths:     60,
		MoratoriumMonths: 6,
	}

	schedule := repayment_v1.GenerateLoanSchedule(loan)

	if schedule[0].EMI != 0 {
		t.Fatal("EMI should be zero during moratorium")
	}

	if schedule[0].Interest <= 0 {
		t.Fatal("interest must accrue during moratorium")
	}
}
