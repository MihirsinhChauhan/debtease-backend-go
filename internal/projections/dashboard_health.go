package projections

import (
	health_v1 "DebtEase/internal/domain/health/v1"

	"DebtEase/internal/database"
	"DebtEase/internal/models"
)

// MergePortfolioRollupIntoHealthInput overlays live portfolio EMI load and active debt count onto snapshot-derived input.
// Single projection surface for dashboard health until engines consume one unified builder everywhere.
func MergePortfolioRollupIntoHealthInput(base health_v1.HealthScoreInput, rollup database.GetPortfolioRollupForUserRow) health_v1.HealthScoreInput {
	base.TotalMonthlyEMI = rollup.TotalMonthlyEmi
	ac := int(rollup.ActiveDebtCount)
	base.ActiveDebtCount = &ac
	return base
}

// FinancialSimulationFromMergedHealthInput builds dashboard loan-simulation payload from merged health inputs (live EMI count).
func FinancialSimulationFromMergedHealthInput(hin health_v1.HealthScoreInput) *models.FinancialSnapshotForSimulation {
	active := 0
	if hin.ActiveDebtCount != nil {
		active = *hin.ActiveDebtCount
	}
	out := models.FinancialSnapshotForSimulation{
		MonthlyIncome:     hin.MonthlyIncome,
		FixedExpenses:     hin.FixedExpenses,
		TotalMonthlyEMI:   hin.TotalMonthlyEMI,
		ActiveDebtCount:   active,
		UnsecuredEmiRatio: hin.UnsecuredEmiRatio,
	}
	if !hin.VariableExpenses.IsZero() {
		v := hin.VariableExpenses
		out.VariableExpenses = &v
	}
	if !hin.SavingsCash.IsZero() {
		v := hin.SavingsCash
		out.SavingsCash = &v
	}
	if !hin.InvestmentsValue.IsZero() {
		v := hin.InvestmentsValue
		out.InvestmentsValue = &v
	}
	return &out
}
