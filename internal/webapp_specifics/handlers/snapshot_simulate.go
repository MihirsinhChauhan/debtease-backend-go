package handlers

import (
	"DebtEase/internal/api"
	health_v1 "DebtEase/internal/domain/health/v1"
	repayment_v1 "DebtEase/internal/domain/repayment/v1"
	stress_v1 "DebtEase/internal/domain/stress/v1"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var snapshotLoanTypes = map[string]bool{
	"personal_loan": true,
	"car_loan":      true,
	"home_loan":     true,
	"consumer_emi":  true,
	"bnpl":          true,
}

func healthBlockFromResult(in health_v1.HealthScoreInput, result health_v1.HealthScoreResult) webapp_models.SnapshotHealthBlock {
	return webapp_models.SnapshotHealthBlock{
		Status:            string(result.Status),
		Score:             result.Score,
		Drivers:           result.Drivers,
		EMIRatio:          health_v1.ComputeEMIRatio(in),
		ShockBufferMonths: health_v1.ComputeShockBufferMonths(in),
		TotalMonthlyEMI:   in.TotalMonthlyEMI,
		EngineVersion:     result.EngineVersion,
	}
}

// HandleSnapshotSimulateLoan computes EMI for a hypothetical loan and returns health before/after (not persisted).
func HandleSnapshotSimulateLoan(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Snapshot simulate loan request", "path", r.URL.Path, "method", r.Method)

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			sessionID = uuid.New()
		}

		var req webapp_models.SnapshotSimulateLoanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for snapshot simulate loan", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		lt := strings.TrimSpace(req.LoanType)
		if lt == "" || !snapshotLoanTypes[lt] {
			utils.RespondWithError(w, http.StatusBadRequest, "loan_type must be personal_loan, car_loan, home_loan, consumer_emi, or bnpl")
			return
		}

		if req.Principal.LessThanOrEqual(decimal.Zero) {
			utils.RespondWithError(w, http.StatusBadRequest, "principal must be greater than zero")
			return
		}
		if req.AnnualInterestRate.IsNegative() {
			utils.RespondWithError(w, http.StatusBadRequest, "annual_interest_rate must be non-negative")
			return
		}
		if req.TenureMonths <= 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "tenure_months must be greater than zero")
			return
		}

		sessEngine := req.EngineVersion
		if sessEngine == "" {
			sessEngine = webapp_service.DefaultEngineVersion
		}

		sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, sessEngine, nil)
		if err != nil {
			logger.Logger.Errorw("Failed to get or create session", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process session")
			return
		}

		var base health_v1.HealthScoreInput

		if req.Snapshot != nil {
			s := req.Snapshot
			if s.MonthlyIncome.LessThanOrEqual(decimal.Zero) {
				utils.RespondWithError(w, http.StatusBadRequest, "snapshot.monthly_income must be greater than zero")
				return
			}
			if s.FixedExpenses.IsNegative() || s.TotalMonthlyEMI.IsNegative() {
				utils.RespondWithError(w, http.StatusBadRequest, "snapshot.fixed_expenses and snapshot.total_monthly_emi must be non-negative")
				return
			}
			if s.ActiveDebtCount < 0 {
				utils.RespondWithError(w, http.StatusBadRequest, "snapshot.active_debt_count must be non-negative")
				return
			}
			if s.UnsecuredEmiRatio != nil && (s.UnsecuredEmiRatio.IsNegative() || s.UnsecuredEmiRatio.GreaterThan(decimal.NewFromInt(1))) {
				utils.RespondWithError(w, http.StatusBadRequest, "snapshot.unsecured_emi_ratio must be between 0 and 1")
				return
			}
			variable := decimal.Zero
			if s.VariableExpenses != nil {
				variable = *s.VariableExpenses
			}
			if variable.IsNegative() {
				utils.RespondWithError(w, http.StatusBadRequest, "snapshot.variable_expenses must be non-negative")
				return
			}
			savings := decimal.Zero
			if s.SavingsCash != nil {
				savings = *s.SavingsCash
			}
			investments := decimal.Zero
			if s.InvestmentsValue != nil {
				investments = *s.InvestmentsValue
			}
			if savings.IsNegative() || investments.IsNegative() {
				utils.RespondWithError(w, http.StatusBadRequest, "snapshot.savings_cash and snapshot.investments_value must be non-negative")
				return
			}
			base = webapp_service.HealthInputFromInlineSnapshot(
				s.MonthlyIncome, s.FixedExpenses, variable, s.TotalMonthlyEMI, savings, investments,
				s.ActiveDebtCount, s.UnsecuredEmiRatio,
			)
		} else {
			snap, err := cfg.DB.GetLatestSnapshotBySession(ctx, uuid.NullUUID{UUID: sessionID, Valid: true})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					utils.RespondWithError(w, http.StatusBadRequest, "No financial snapshot for this session; run analyze first or pass snapshot in the request body")
					return
				}
				logger.Logger.Errorw("Failed to load latest snapshot", "error", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load snapshot")
				return
			}
			base, err = webapp_service.HealthInputFromFinancialSnapshot(snap)
			if err != nil {
				logger.Logger.Errorw("Failed to map snapshot to health input", "error", err)
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid stored snapshot data")
				return
			}
		}

		newEMI := repayment_v1.CalculateEMI(req.Principal, req.AnnualInterestRate, req.TenureMonths)
		if newEMI.IsNegative() {
			utils.RespondWithError(w, http.StatusBadRequest, "Computed EMI is invalid")
			return
		}

		beforeResult := health_v1.Run(base)
		beforeBlock := healthBlockFromResult(base, beforeResult)

		oldUnsecuredRatio := health_v1.ComputeUnsecuredRatio(base)
		newTotalEMI := base.TotalMonthlyEMI.Add(newEMI)
		secured := webapp_service.LoanTypeSecured(lt)
		nextRatio := webapp_service.NextUnsecuredEmiRatio(base.TotalMonthlyEMI, oldUnsecuredRatio, newEMI, secured)

		nextCount := 0
		if base.ActiveDebtCount != nil {
			nextCount = *base.ActiveDebtCount
		}
		nextCount++

		afterIn := health_v1.HealthScoreInput{
			MonthlyIncome:     base.MonthlyIncome,
			FixedExpenses:     base.FixedExpenses,
			VariableExpenses:  base.VariableExpenses,
			SavingsCash:       base.SavingsCash,
			InvestmentsValue:  base.InvestmentsValue,
			TotalMonthlyEMI:   newTotalEMI,
			Debts:             nil,
			ActiveDebtCount:   &nextCount,
			UnsecuredEmiRatio: &nextRatio,
		}
		afterResult := health_v1.Run(afterIn)
		afterBlock := healthBlockFromResult(afterIn, afterResult)

		baseFixed := base.FixedExpenses.Add(base.VariableExpenses)
		afterFixed := afterIn.FixedExpenses.Add(afterIn.VariableExpenses)
		beforeCur, beforeS10, beforeS20 := stress_v1.FlexibilityBuffersForIncomeShocks(
			base.MonthlyIncome, baseFixed, base.TotalMonthlyEMI,
		)
		afterCur, afterS10, afterS20 := stress_v1.FlexibilityBuffersForIncomeShocks(
			afterIn.MonthlyIncome, afterFixed, afterIn.TotalMonthlyEMI,
		)
		worstMonths := stress_v1.WorstSurvivabilityMonthsAcrossIncomeShocks(
			afterIn.MonthlyIncome, afterFixed, afterIn.TotalMonthlyEMI,
		)
		survBand := stress_v1.ComputeSurvivabilityBand(worstMonths, stress_v1.DefaultStressPolicy())

		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "loan_simulation_started", map[string]interface{}{
			"loan_type":            lt,
			"principal":            req.Principal.String(),
			"annual_interest_rate": req.AnnualInterestRate.String(),
			"tenure_months":        req.TenureMonths,
			"new_emi":              newEMI.String(),
			"before_status":        beforeBlock.Status,
			"after_status":         afterBlock.Status,
			"engine_version":       afterResult.EngineVersion,
		}); err != nil {
			logger.Logger.Warnw("Failed to emit loan_simulation_started event", "error", err)
		}

		webapp_utils.SetSessionCookie(w, sessionID)

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.SnapshotSimulateLoanResponse{
			Before:            beforeBlock,
			After:             afterBlock,
			NewLoanEMIMonthly: newEMI,
			LoanType:          lt,
			EngineVersion:     afterResult.EngineVersion,
			Visual: webapp_models.SnapshotSimulateLoanVisual{
				Buffers: webapp_models.SnapshotSimulateLoanBuffers{
					Current: afterCur,
					Shock10: afterS10,
					Shock20: afterS20,
				},
				BuffersBefore: webapp_models.SnapshotSimulateLoanBuffers{
					Current: beforeCur,
					Shock10: beforeS10,
					Shock20: beforeS20,
				},
			},
			SurvivabilityBand: string(survBand),
			Delta: webapp_models.SnapshotSimulateLoanDelta{
				StressLevelChange: beforeBlock.Status != afterBlock.Status,
				BufferDrop:        afterCur.LessThan(beforeCur),
			},
		})
	}
}
