package service

import (
	"encoding/json"

	"DebtEase/internal/database"
	health_v1 "DebtEase/internal/domain/health/v1"

	"github.com/shopspring/decimal"
)

type snapshotExtraData struct {
	ActiveDebtCount   int     `json:"active_debt_count"`
	UnsecuredEmiRatio *string `json:"unsecured_emi_ratio,omitempty"`
}

// HealthInputFromFinancialSnapshot maps a persisted financial_snapshots row to health_v1 input.
func HealthInputFromFinancialSnapshot(snap database.FinancialSnapshot) (health_v1.HealthScoreInput, error) {
	var extra snapshotExtraData
	if len(snap.ExtraData) > 0 {
		if err := json.Unmarshal(snap.ExtraData, &extra); err != nil {
			return health_v1.HealthScoreInput{}, err
		}
	}
	active := extra.ActiveDebtCount
	var unsecuredPtr *decimal.Decimal
	if extra.UnsecuredEmiRatio != nil && *extra.UnsecuredEmiRatio != "" {
		u, err := decimal.NewFromString(*extra.UnsecuredEmiRatio)
		if err != nil {
			return health_v1.HealthScoreInput{}, err
		}
		unsecuredPtr = &u
	}
	return health_v1.HealthScoreInput{
		MonthlyIncome:     snap.MonthlyIncome,
		FixedExpenses:     snap.FixedExpenses,
		VariableExpenses:  snap.VariableExpenses,
		SavingsCash:       snap.SavingsCash,
		InvestmentsValue:  snap.InvestmentsValue,
		TotalMonthlyEMI:   snap.TotalMonthlyEmi,
		Debts:             nil,
		ActiveDebtCount:   &active,
		UnsecuredEmiRatio: unsecuredPtr,
	}, nil
}

// HealthInputFromInlineSnapshot maps aggregate fields (API body) to health_v1 input.
func HealthInputFromInlineSnapshot(
	monthlyIncome, fixedExpenses, variableExpenses, totalMonthlyEMI, savingsCash, investments decimal.Decimal,
	activeDebtCount int,
	unsecuredEmiRatio *decimal.Decimal,
) health_v1.HealthScoreInput {
	ac := activeDebtCount
	return health_v1.HealthScoreInput{
		MonthlyIncome:     monthlyIncome,
		FixedExpenses:     fixedExpenses,
		VariableExpenses:  variableExpenses,
		SavingsCash:       savingsCash,
		InvestmentsValue:  investments,
		TotalMonthlyEMI:   totalMonthlyEMI,
		Debts:             nil,
		ActiveDebtCount:   &ac,
		UnsecuredEmiRatio: unsecuredEmiRatio,
	}
}

// LoanTypeSecured returns true for car and home loans (secured collateral).
func LoanTypeSecured(loanType string) bool {
	switch loanType {
	case "car_loan", "home_loan":
		return true
	default:
		return false
	}
}

// NextUnsecuredEmiRatio blends existing aggregate unsecured share with a new loan EMI.
func NextUnsecuredEmiRatio(oldTotalEMI, oldUnsecuredRatio, newEMI decimal.Decimal, newLoanSecured bool) decimal.Decimal {
	if oldTotalEMI.IsNegative() || newEMI.IsNegative() {
		return decimal.Zero
	}
	oldUnsecuredPart := oldTotalEMI.Mul(oldUnsecuredRatio)
	var add decimal.Decimal
	if newLoanSecured {
		add = decimal.Zero
	} else {
		add = newEMI
	}
	newTotal := oldTotalEMI.Add(newEMI)
	if newTotal.IsZero() {
		return decimal.Zero
	}
	return oldUnsecuredPart.Add(add).Div(newTotal)
}
