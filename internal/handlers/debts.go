package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/domain/debt"
	financial_events "DebtEase/internal/domain/financial_events/v1"
	"DebtEase/internal/ledger"
	"DebtEase/internal/logger"
	"DebtEase/internal/middleware"
	"DebtEase/internal/models"
	"DebtEase/internal/projections"
	"DebtEase/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sqlc-dev/pqtype"
)

func getUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(middleware.ContextUserIDKey)
	if val == nil {
		return uuid.Nil, false
	}
	uid, ok := val.(uuid.UUID)
	return uid, ok
}

func HandleCreateDebt(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var req models.CreateDebtRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		if req.Name == "" || req.Type == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing required fields")
			return
		}

		// req.DueDate is time.Time per model; JSON must be RFC3339. We accept as-is.
		dueDate := req.DueDate

		var paymentFreq database.NullPaymentFreq
		if req.PaymentFrequency != nil && *req.PaymentFrequency != "" {
			paymentFreq = database.NullPaymentFreq{PaymentFreq: database.PaymentFreq(*req.PaymentFrequency), Valid: true}
		}

		var debtTaken sql.NullTime
		if req.DebtTaken != nil {
			debtTaken = sql.NullTime{Time: *req.DebtTaken, Valid: true}
		}

		var lender sql.NullString
		if req.Lender != nil {
			lender = sql.NullString{String: *req.Lender, Valid: *req.Lender != ""}
		}

		var priority sql.NullInt16
		if req.Priority != nil {
			priority = sql.NullInt16{Int16: *req.Priority, Valid: true}
		}

		var notes sql.NullString
		if req.Notes != nil {
			notes = sql.NullString{String: *req.Notes, Valid: *req.Notes != ""}
		}

		var paymentDueDay sql.NullInt32
		if req.PaymentDueDay != nil {
			paymentDueDay = sql.NullInt32{Int32: *req.PaymentDueDay, Valid: true}
		}

		var moratoriumUntil sql.NullTime
		if req.MoratoriumUntil != nil {
			moratoriumUntil = sql.NullTime{Time: *req.MoratoriumUntil, Valid: true}
		}

		inMoratorium := req.MoratoriumUntil != nil && req.MoratoriumUntil.After(time.Now())

		obl := debt.NormalizeCreate(debt.CreateObligationInput{
			DebtType:      database.DebtType(req.Type),
			InterestRate:  req.InterestRate,
			MinPayment:    req.MinPayment,
			Principal:     req.Principal,
			EMIAmount:     req.EMIAmount,
			TenureMonths:  req.TenureMonths,
			PaymentDueDay: req.PaymentDueDay,
			DueDate:       dueDate,
		})

		var treatment sql.NullString
		if req.TreatmentClassification != nil && *req.TreatmentClassification != "" {
			treatment = sql.NullString{String: *req.TreatmentClassification, Valid: true}
		}

		params := database.CreateDebtParams{
			UserID:                  userID,
			Name:                    req.Name,
			Type:                    database.DebtType(req.Type),
			Lender:                  lender,
			PrincipalDecimal:        req.Principal,
			OutstandingBalance:      req.OutstandingBalance,
			InterestRate:            req.InterestRate,
			MinPayment:              req.MinPayment,
			DueDate:                 dueDate,
			PaymentFrequency:        paymentFreq,
			DebtTaken:               debtTaken,
			Priority:                priority,
			Notes:                   notes,
			PaymentDueDay:           paymentDueDay,
			MoratoriumUntil:         moratoriumUntil,
			InMoratorium:            sql.NullBool{Bool: inMoratorium, Valid: true},
			InstrumentType:          obl.InstrumentType,
			InterestModel:           obl.InterestModel,
			MonthlyObligation:       obl.MonthlyObligation,
			MinimumObligation:       obl.MinimumObligation,
			RemainingMonths:         obl.RemainingMonths,
			DueDay:                  obl.DueDay,
			IsSecured:               obl.IsSecured,
			IsRevolving:             obl.IsRevolving,
			IsShortTerm:             obl.IsShortTerm,
			IsInterestBearing:       obl.IsInterestBearing,
			FragmentationWeight:     obl.FragmentationWeight,
			StressPriority:          obl.StressPriority,
			TreatmentClassification: treatment,
			EngineVersion:           obl.EngineVersion,
			PolicyVersion:           obl.PolicyVersion,
			ExtraMetadata:           pqtype.NullRawMessage{},
		}

		if req.Type == "credit_card" {
			if req.BillingDay == nil || req.GracePeriodDays == nil {
				utils.RespondWithError(w, http.StatusBadRequest, "billing_day and grace_period_days required for credit cards")
				return
			}
		}

		ctx := context.Background()
		if cfg.DBPool == nil {
			logger.Logger.Errorw("DBPool not configured")
			utils.RespondWithError(w, http.StatusInternalServerError, "Server misconfiguration")
			return
		}

		tx, err := cfg.DBPool.BeginTx(ctx, nil)
		if err != nil {
			logger.Logger.Errorw("BeginTx failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create debt")
			return
		}
		defer func() { _ = tx.Rollback() }()

		qtx := cfg.DB.WithTx(tx)

		debt, err := qtx.CreateDebt(ctx, params)
		if err != nil {
			logger.Logger.Errorw("CreateDebt failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create debt")
			return
		}

		if req.Type == "credit_card" {
			minPayPercent := decimal.NewFromFloat(5.00)
			if req.MinPaymentPercent != nil {
				minPayPercent = *req.MinPaymentPercent
			}

			ccParams := database.CreateCreditCardParams{
				DebtID:            debt.ID,
				BillingDay:        *req.BillingDay,
				GracePeriodDays:   *req.GracePeriodDays,
				MinPaymentPercent: minPayPercent,
				BilledBalance:     decimal.Zero,
				UnbilledBalance:   req.OutstandingBalance,
				InGracePeriod:     true,
				MinimumDue:        req.MinPayment,
			}

			if _, err := qtx.CreateCreditCard(ctx, ccParams); err != nil {
				logger.Logger.Errorw("CreateCreditCard failed", "error", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create credit card")
				return
			}
		}

		meta := map[string]interface{}{
			financial_events.MetaDebtID:         debt.ID.String(),
			financial_events.MetaDebtType:       string(debt.Type),
			financial_events.MetaOpeningBalance: debt.OutstandingBalance.StringFixed(2),
		}
		if !debt.MonthlyObligation.IsZero() {
			meta[financial_events.MetaEMIAmount] = debt.MonthlyObligation.StringFixed(2)
		}
		bal := debt.OutstandingBalance
		if err := ledger.Append(ctx, qtx, ledger.Params{
			UserID:     userID,
			EventType:  financial_events.EventDebtInstrumentCreated,
			EntityType: financial_events.EntityDebt,
			EntityID:   debt.ID,
			Amount:     &bal,
			OccurredAt: time.Now().UTC(),
			Metadata:   meta,
		}); err != nil {
			logger.Logger.Errorw("financial ledger append failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create debt")
			return
		}

		if err := projections.RegenerateAutoGeneratedFinancialSnapshot(ctx, qtx, userID, financial_events.EventDebtInstrumentCreated); err != nil {
			logger.Logger.Errorw("snapshot regeneration failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create debt")
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Logger.Errorw("Commit failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create debt")
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, toDebtResponse(debt))
	}
}

func HandleListDebts(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx := context.Background()
		items, err := cfg.DB.ListUserDebts(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("ListUserDebts failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list debts")
			return
		}
		resp := make([]models.DebtResponse, 0, len(items))
		for _, d := range items {
			resp = append(resp, toDebtResponse(d))
		}
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func HandleGetDebt(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		ctx := context.Background()
		debt, err := cfg.DB.GetDebtByID(ctx, database.GetDebtByIDParams{ID: debtID, UserID: userID})
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			logger.Logger.Errorw("GetDebtByID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch debt")
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, toDebtResponse(debt))
	}
}

func HandleUpdateDebtBalance(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		var req models.UpdateDebtBalanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		if req.OutstandingBalance.IsZero() && !req.OutstandingBalance.GreaterThan(decimal.Zero) {
			utils.RespondWithError(w, http.StatusBadRequest, "outstanding_balance is required")
			return
		}

		ctx := context.Background()
		debt, err := cfg.DB.GetDebtByID(ctx, database.GetDebtByIDParams{ID: debtID, UserID: userID})
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			logger.Logger.Errorw("GetDebtByID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load debt")
			return
		}
		prevBal := debt.OutstandingBalance

		if cfg.DBPool == nil {
			logger.Logger.Errorw("DBPool not configured")
			utils.RespondWithError(w, http.StatusInternalServerError, "Server misconfiguration")
			return
		}

		tx, err := cfg.DBPool.BeginTx(ctx, nil)
		if err != nil {
			logger.Logger.Errorw("BeginTx failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update balance")
			return
		}
		defer func() { _ = tx.Rollback() }()

		qtx := cfg.DB.WithTx(tx)

		isOverdue := sql.NullBool{Bool: req.IsOverdue, Valid: true}

		if err := qtx.UpdateDebtBalance(ctx, database.UpdateDebtBalanceParams{
			ID:                 debtID,
			OutstandingBalance: req.OutstandingBalance,
			IsOverdue:          isOverdue,
		}); err != nil {
			logger.Logger.Errorw("UpdateDebtBalance failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update balance")
			return
		}

		delta := req.OutstandingBalance.Sub(prevBal)
		meta := map[string]interface{}{
			financial_events.MetaDebtID:          debtID.String(),
			financial_events.MetaPreviousBalance: prevBal.StringFixed(2),
			financial_events.MetaNewBalance:      req.OutstandingBalance.StringFixed(2),
		}
		if err := ledger.Append(ctx, qtx, ledger.Params{
			UserID:     userID,
			EventType:  financial_events.EventBalanceManuallyAdjusted,
			EntityType: financial_events.EntityDebt,
			EntityID:   debtID,
			Amount:     &delta,
			OccurredAt: time.Now().UTC(),
			Metadata:   meta,
		}); err != nil {
			logger.Logger.Errorw("financial ledger append failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update balance")
			return
		}

		if err := projections.RegenerateAutoGeneratedFinancialSnapshot(ctx, qtx, userID, financial_events.EventBalanceManuallyAdjusted); err != nil {
			logger.Logger.Errorw("snapshot regeneration failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update balance")
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Logger.Errorw("Commit failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update balance")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleUpdateDebtHealth(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getUserIDFromContext(r.Context()); !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		var req models.UpdateDebtHealthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		var score sql.NullString
		if req.DebtHealthScore != nil {
			score = sql.NullString{String: req.DebtHealthScore.StringFixed(2), Valid: true}
		}

		ctx := context.Background()
		if err := cfg.DB.UpdateDebtHealthScore(ctx, database.UpdateDebtHealthScoreParams{
			ID:              debtID,
			DebtHealthScore: score,
		}); err != nil {
			logger.Logger.Errorw("UpdateDebtHealthScore failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update health score")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleUpdateDebt(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		var req models.UpdateDebtRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		ctx := context.Background()
		existing, err := cfg.DB.GetDebtByID(ctx, database.GetDebtByIDParams{ID: debtID, UserID: userID})
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			logger.Logger.Errorw("GetDebtByID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load debt")
			return
		}

		name := existing.Name
		if req.Name != nil && *req.Name != "" {
			name = *req.Name
		}

		lender := existing.Lender
		if req.Lender != nil {
			lender = sql.NullString{String: *req.Lender, Valid: *req.Lender != ""}
		}

		principal := existing.PrincipalDecimal
		if req.Principal != nil {
			principal = *req.Principal
		}

		outstanding := existing.OutstandingBalance
		if req.OutstandingBalance != nil {
			outstanding = *req.OutstandingBalance
		}

		interestRate := existing.InterestRate
		if req.InterestRate != nil {
			interestRate = *req.InterestRate
		}

		minPayment := existing.MinPayment
		if req.MinPayment != nil {
			minPayment = *req.MinPayment
		}

		dueDate := existing.DueDate
		if req.DueDate != nil {
			dueDate = *req.DueDate
		}

		notes := existing.Notes
		if req.Notes != nil {
			notes = sql.NullString{String: *req.Notes, Valid: *req.Notes != ""}
		}

		paymentDueDay := existing.PaymentDueDay
		if req.PaymentDueDay != nil {
			paymentDueDay = sql.NullInt32{Int32: *req.PaymentDueDay, Valid: true}
		}

		var tenureMonths *int32
		if existing.RemainingMonths.Valid {
			v := existing.RemainingMonths.Int32
			tenureMonths = &v
		}
		if req.TenureMonths != nil {
			tenureMonths = req.TenureMonths
		}

		emiAmount := req.EMIAmount
		if emiAmount == nil && !minPayment.IsZero() {
			v := minPayment
			emiAmount = &v
		}

		var paymentDueDayPtr *int32
		if paymentDueDay.Valid {
			v := paymentDueDay.Int32
			paymentDueDayPtr = &v
		}

		obl := debt.NormalizeCreate(debt.CreateObligationInput{
			DebtType:      existing.Type,
			InterestRate:  interestRate,
			MinPayment:    minPayment,
			Principal:     principal,
			EMIAmount:     emiAmount,
			TenureMonths:  tenureMonths,
			PaymentDueDay: paymentDueDayPtr,
			DueDate:       dueDate,
		})

		if cfg.DBPool == nil {
			logger.Logger.Errorw("DBPool not configured")
			utils.RespondWithError(w, http.StatusInternalServerError, "Server misconfiguration")
			return
		}

		tx, err := cfg.DBPool.BeginTx(ctx, nil)
		if err != nil {
			logger.Logger.Errorw("BeginTx failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update debt")
			return
		}
		defer func() { _ = tx.Rollback() }()

		qtx := cfg.DB.WithTx(tx)

		updated, err := qtx.UpdateDebtDetails(ctx, database.UpdateDebtDetailsParams{
			ID:                 debtID,
			UserID:             userID,
			Name:               name,
			Lender:             lender,
			PrincipalDecimal:   principal,
			OutstandingBalance: outstanding,
			InterestRate:       interestRate,
			MinPayment:         minPayment,
			DueDate:            dueDate,
			Notes:              notes,
			PaymentDueDay:      paymentDueDay,
			MonthlyObligation:  obl.MonthlyObligation,
			MinimumObligation:  obl.MinimumObligation,
			RemainingMonths:    obl.RemainingMonths,
			DueDay:             obl.DueDay,
			IsInterestBearing:  obl.IsInterestBearing,
		})
		if err != nil {
			logger.Logger.Errorw("UpdateDebtDetails failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update debt")
			return
		}

		if existing.Type == database.DebtTypeCreditCard {
			cc, ccErr := qtx.GetCreditCardByDebtID(ctx, debtID)
			if ccErr != nil && ccErr != sql.ErrNoRows {
				logger.Logger.Errorw("GetCreditCardByDebtID failed", "error", ccErr)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update debt")
				return
			}
			if ccErr == nil {
				billingDay := cc.BillingDay
				if req.BillingDay != nil {
					billingDay = *req.BillingDay
				}
				graceDays := cc.GracePeriodDays
				if req.GracePeriodDays != nil {
					graceDays = *req.GracePeriodDays
				}
				if err := qtx.UpdateCreditCardDetails(ctx, database.UpdateCreditCardDetailsParams{
					DebtID:          debtID,
					BillingDay:      billingDay,
					GracePeriodDays: graceDays,
					MinimumDue:      minPayment,
				}); err != nil {
					logger.Logger.Errorw("UpdateCreditCardDetails failed", "error", err)
					utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update debt")
					return
				}
			}
		}

		if err := projections.RegenerateAutoGeneratedFinancialSnapshot(ctx, qtx, userID, financial_events.EventBalanceManuallyAdjusted); err != nil {
			logger.Logger.Errorw("snapshot regeneration failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update debt")
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Logger.Errorw("Commit failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update debt")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, toDebtResponse(updated))
	}
}

func HandleMarkDebtOverdue(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getUserIDFromContext(r.Context()); !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		ctx := context.Background()
		if err := cfg.DB.MarkDebtOverdue(ctx, debtID); err != nil {
			logger.Logger.Errorw("MarkDebtOverdue failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to mark overdue")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// func HandleDeleteDebt(cfg *api.Config) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if _, ok := getUserIDFromContext(r.Context()); !ok {
// 			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
// 			return
// 		}

// 		idStr := r.PathValue("id")
// 		debtID, err := uuid.Parse(idStr)
// 		if err != nil {
// 			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
// 			return
// 		}

// 		ctx := context.Background()
// 		if err := cfg.DB.DeleteDebt(ctx, debtID); err != nil {
// 			logger.Logger.Errorw("DeleteDebt failed", "error", err)
// 			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete debt")
// 			return
// 		}
// 		w.WriteHeader(http.StatusNoContent)
// 	}
// }

func toDebtResponse(d database.Debt) models.DebtResponse {
	var lender *string
	if d.Lender.Valid {
		lender = &d.Lender.String
	}
	var pf *string
	if d.PaymentFrequency.Valid {
		s := string(d.PaymentFrequency.PaymentFreq)
		pf = &s
	}
	isOverdue := false
	if d.IsOverdue.Valid {
		isOverdue = d.IsOverdue.Bool
	}
	var lateFees *decimal.Decimal
	if d.LateFees.Valid {
		if v, err := decimal.NewFromString(d.LateFees.String); err == nil {
			lateFees = &v
		}
	}
	var otherCharges *string
	if d.OtherCharges.Valid {
		otherCharges = &d.OtherCharges.String
	}
	var debtTaken *time.Time
	if d.DebtTaken.Valid {
		debtTaken = &d.DebtTaken.Time
	}
	var score *decimal.Decimal
	if d.DebtHealthScore.Valid {
		if v, err := decimal.NewFromString(d.DebtHealthScore.String); err == nil {
			score = &v
		}
	}
	var priority *int16
	if d.Priority.Valid {
		p := d.Priority.Int16
		priority = &p
	}
	var notes *string
	if d.Notes.Valid {
		notes = &d.Notes.String
	}

	createdAt := time.Time{}
	if d.CreatedAt.Valid {
		createdAt = d.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if d.UpdatedAt.Valid {
		updatedAt = d.UpdatedAt.Time
	}

	// AccruedInterest and LastAccrualDate are direct fields, not Null types
	accruedInterest := d.AccruedInterest
	lastAccrualDate := d.LastAccrualDate

	var treatment *string
	if d.TreatmentClassification.Valid {
		s := d.TreatmentClassification.String
		treatment = &s
	}

	var remainingMonths *int32
	if d.RemainingMonths.Valid {
		v := d.RemainingMonths.Int32
		remainingMonths = &v
	}
	var dueDay *int32
	if d.DueDay.Valid {
		v := d.DueDay.Int32
		dueDay = &v
	}
	var paymentDueDay *int32
	if d.PaymentDueDay.Valid {
		v := d.PaymentDueDay.Int32
		paymentDueDay = &v
	}
	var moratoriumUntil *time.Time
	if d.MoratoriumUntil.Valid {
		moratoriumUntil = &d.MoratoriumUntil.Time
	}
	var inMoratorium *bool
	if d.InMoratorium.Valid {
		v := d.InMoratorium.Bool
		inMoratorium = &v
	}

	return models.DebtResponse{
		ID:                 d.ID.String(),
		UserID:             d.UserID.String(),
		Name:               d.Name,
		Type:               string(d.Type),
		Lender:             lender,
		Principal:          d.PrincipalDecimal,
		OutstandingBalance: d.OutstandingBalance,
		InterestRate:       d.InterestRate,
		MinPayment:         d.MinPayment,
		DueDate:            d.DueDate.Format("2006-01-02"),
		PaymentFrequency:   pf,
		IsOverdue:          isOverdue,
		LateFees:           lateFees,
		OtherCharges:       otherCharges,
		DebtTaken:          debtTaken,
		DebtHealthScore:    score,
		Priority:           priority,
		Notes:              notes,
		AccruedInterest:    accruedInterest,
		LastAccrualDate:    lastAccrualDate,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,

		InstrumentType:          string(d.InstrumentType),
		InterestModel:           string(d.InterestModel),
		MonthlyObligation:       d.MonthlyObligation,
		MinimumObligation:       d.MinimumObligation,
		RemainingMonths:         remainingMonths,
		DueDay:                  dueDay,
		IsSecured:               d.IsSecured,
		IsRevolving:             d.IsRevolving,
		IsShortTerm:             d.IsShortTerm,
		IsInterestBearing:       d.IsInterestBearing,
		FragmentationWeight:     d.FragmentationWeight,
		StressPriority:          d.StressPriority,
		TreatmentClassification: treatment,
		EngineVersion:           d.EngineVersion,
		PolicyVersion:           d.PolicyVersion,

		PaymentDueDay:   paymentDueDay,
		MoratoriumUntil: moratoriumUntil,
		InMoratorium:    inMoratorium,
	}
}
