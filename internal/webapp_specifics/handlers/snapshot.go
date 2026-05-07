package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/auth"
	"DebtEase/internal/database"
	health_v1 "DebtEase/internal/domain/health/v1"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var snapshotEntryPoints = map[string]bool{
	"debt_health": true,
	"loan_test":   true,
	"landing":     true,
}

func optionalAuthUserID(cfg *api.Config, r *http.Request) uuid.UUID {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.Nil
	}
	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		return uuid.Nil
	}
	return userID
}

// HandleSnapshotAnalyze persists a financial_snapshots row and returns health_v1 results.
func HandleSnapshotAnalyze(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Snapshot analyze request", "path", r.URL.Path, "method", r.Method)

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			sessionID = uuid.New()
		}

		var req webapp_models.SnapshotAnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for snapshot analyze", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		ep := strings.TrimSpace(req.EntryPoint)
		if ep == "" || !snapshotEntryPoints[ep] {
			utils.RespondWithError(w, http.StatusBadRequest, "entry_point must be debt_health, loan_test, or landing")
			return
		}

		if req.MonthlyIncome.LessThanOrEqual(decimal.Zero) {
			utils.RespondWithError(w, http.StatusBadRequest, "monthly_income must be greater than zero")
			return
		}
		if req.FixedExpenses.IsNegative() || req.TotalMonthlyEMI.IsNegative() {
			utils.RespondWithError(w, http.StatusBadRequest, "fixed_expenses and total_monthly_emi must be non-negative")
			return
		}
		if req.ActiveDebtCount < 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "active_debt_count must be non-negative")
			return
		}
		if req.UnsecuredEmiRatio != nil && (req.UnsecuredEmiRatio.IsNegative() || req.UnsecuredEmiRatio.GreaterThan(decimal.NewFromInt(1))) {
			utils.RespondWithError(w, http.StatusBadRequest, "unsecured_emi_ratio must be between 0 and 1")
			return
		}

		sessEngine := req.EngineVersion
		if sessEngine == "" {
			sessEngine = webapp_service.DefaultEngineVersion
		}

		authUser := optionalAuthUserID(cfg, r)
		meta := &webapp_service.SessionMeta{EntryPoint: ep}
		if authUser != uuid.Nil {
			meta.UserID = authUser
		}

		sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, sessEngine, meta)
		if err != nil {
			logger.Logger.Errorw("Failed to get or create session", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process session")
			return
		}

		variable := decimal.Zero
		if req.VariableExpenses != nil {
			variable = *req.VariableExpenses
		}
		if variable.IsNegative() {
			utils.RespondWithError(w, http.StatusBadRequest, "variable_expenses must be non-negative")
			return
		}

		savings := decimal.Zero
		if req.SavingsCash != nil {
			savings = *req.SavingsCash
		}
		investments := decimal.Zero
		if req.InvestmentsValue != nil {
			investments = *req.InvestmentsValue
		}
		if savings.IsNegative() || investments.IsNegative() {
			utils.RespondWithError(w, http.StatusBadRequest, "savings_cash and investments_value must be non-negative")
			return
		}

		activeCount := req.ActiveDebtCount
		healthIn := health_v1.HealthScoreInput{
			MonthlyIncome:     req.MonthlyIncome,
			FixedExpenses:     req.FixedExpenses,
			VariableExpenses:  variable,
			SavingsCash:       savings,
			InvestmentsValue:  investments,
			TotalMonthlyEMI:   req.TotalMonthlyEMI,
			Debts:             nil,
			ActiveDebtCount:   &activeCount,
			UnsecuredEmiRatio: req.UnsecuredEmiRatio,
		}

		result := health_v1.Run(healthIn)

		extra := map[string]interface{}{
			"active_debt_count": activeCount,
			"entry_point":       ep,
		}
		if req.UnsecuredEmiRatio != nil {
			extra["unsecured_emi_ratio"] = req.UnsecuredEmiRatio.String()
		}
		extraBytes, err := json.Marshal(extra)
		if err != nil {
			logger.Logger.Errorw("Failed to marshal snapshot extra_data", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to persist snapshot")
			return
		}

		var userID uuid.NullUUID
		if authUser != uuid.Nil {
			userID = uuid.NullUUID{UUID: authUser, Valid: true}
		}
		snap, err := cfg.DB.InsertFinancialSnapshot(ctx, database.InsertFinancialSnapshotParams{
			UserID:           userID,
			SessionID:        uuid.NullUUID{UUID: sessionID, Valid: true},
			SnapshotType:     "onboarding",
			SnapshotSource:   "user_input",
			MonthlyIncome:    req.MonthlyIncome,
			FixedExpenses:    req.FixedExpenses,
			VariableExpenses: variable,
			SavingsCash:      savings,
			InvestmentsValue: investments,
			TotalMonthlyEmi:  req.TotalMonthlyEMI,
			OtherIncome:      decimal.Zero,
			TotalDebtBalance: decimal.Zero,
			TriggerEvent:     sql.NullString{String: "financial_snapshot_submitted", Valid: true},
			ExtraData:        extraBytes,
		})
		if err != nil {
			logger.Logger.Errorw("Failed to insert financial snapshot", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to persist snapshot")
			return
		}

		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "financial_snapshot_submitted", map[string]interface{}{
			"snapshot_id":    snap.ID.String(),
			"entry_point":    ep,
			"engine_version": result.EngineVersion,
		}); err != nil {
			logger.Logger.Warnw("Failed to emit financial_snapshot_submitted event", "error", err)
		}

		webapp_utils.SetSessionCookie(w, sessionID)

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.SnapshotAnalyzeResponse{
			Status:            string(result.Status),
			Score:             result.Score,
			Drivers:           result.Drivers,
			EMIRatio:          health_v1.ComputeEMIRatio(healthIn),
			ShockBufferMonths: health_v1.ComputeShockBufferMonths(healthIn),
			SnapshotID:        snap.ID,
			EngineVersion:     result.EngineVersion,
		})
	}
}
