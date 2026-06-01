package projections

import (
	health_v1 "DebtEase/internal/domain/health/v1"

	"DebtEase/internal/database"
	"DebtEase/internal/models"

	"github.com/shopspring/decimal"
)

var (
	incomeShock10 = decimal.RequireFromString("0.9")
	incomeShock20 = decimal.RequireFromString("0.8")
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

// IncomeStressPreviewFromHealthInput re-runs health_v1 at full, −10%, and −20% income.
func IncomeStressPreviewFromHealthInput(hin health_v1.HealthScoreInput) models.IncomeStressPreview {
	current := health_v1.Run(hin)

	shock10 := hin
	shock10.MonthlyIncome = hin.MonthlyIncome.Mul(incomeShock10)
	result10 := health_v1.Run(shock10)

	shock20 := hin
	shock20.MonthlyIncome = hin.MonthlyIncome.Mul(incomeShock20)
	result20 := health_v1.Run(shock20)

	primaryReason := ""
	if len(current.Drivers) > 0 {
		primaryReason = current.Drivers[0]
	}

	return models.IncomeStressPreview{
		CurrentStatus: string(current.Status),
		Shock10Status: string(result10.Status),
		Shock20Status: string(result20.Status),
		PrimaryReason: primaryReason,
	}
}
