package classification_v1

import (
	cv1 "DebtEase/internal/domain/classification/v1"
	"github.com/shopspring/decimal"
)

func testDec(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}

func makeInput(income, fixedObl string, debts []cv1.DebtEntry) cv1.ClassificationInput {
	return cv1.ClassificationInput{
		Debts: debts,
		Context: cv1.FinancialContext{
			MonthlyNetIncome: testDec(income),
			FixedObligations: testDec(fixedObl),
		},
	}
}

func debtEntry(id string, emi string, remainingMonths int, secured bool, inst cv1.DebtInstrumentType) cv1.DebtEntry {
	return cv1.DebtEntry{
		ID:              id,
		EMI:             testDec(emi),
		RemainingMonths: remainingMonths,
		Secured:         secured,
		InstrumentType:  inst,
	}
}

func hasFlag(flags []cv1.Flag, f cv1.Flag) bool {
	for _, x := range flags {
		if x == f {
			return true
		}
	}
	return false
}
