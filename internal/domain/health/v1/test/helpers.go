package health_v1

import (
	hv1 "DebtEase/internal/domain/health/v1"
	"github.com/shopspring/decimal"
)

func testDec(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}

func makeInput(income, fixed, variable, savings, investments, totalEMI string, debts []hv1.HealthDebtEntry) hv1.HealthScoreInput {
	return hv1.HealthScoreInput{
		MonthlyIncome:    testDec(income),
		FixedExpenses:    testDec(fixed),
		VariableExpenses: testDec(variable),
		SavingsCash:      testDec(savings),
		InvestmentsValue: testDec(investments),
		TotalMonthlyEMI:  testDec(totalEMI),
		Debts:            debts,
	}
}

func debtEntry(id, emi string, secured bool) hv1.HealthDebtEntry {
	return hv1.HealthDebtEntry{
		ID:      id,
		EMI:     testDec(emi),
		Secured: secured,
	}
}
