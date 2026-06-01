package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func HandleCreateFeedback(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Create feedback request", "path", r.URL.Path, "method", r.Method)

		var req webapp_models.AnalysisFeedbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for feedback", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			if req.SessionID != nil && *req.SessionID != uuid.Nil {
				sessionID = *req.SessionID
			}
		}

		if sessionID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id")
			return
		}

		exists, err := webapp_service.SessionExists(ctx, cfg.DB, sessionID)
		if err != nil {
			logger.Logger.Errorw("Failed to check session existence", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to validate session")
			return
		}
		if !exists {
			sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, webapp_service.DefaultEngineVersion, nil)
			if err != nil {
				logger.Logger.Errorw("Failed to ensure session for feedback", "error", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to validate session")
				return
			}
		}

		var userID uuid.NullUUID
		if authUser := optionalAuthUserID(cfg, r); authUser != uuid.Nil {
			userID = uuid.NullUUID{UUID: authUser, Valid: true}
		}

		in := webapp_service.CreateFeedbackInput{
			UserID:    userID,
			SessionID: uuid.NullUUID{UUID: sessionID, Valid: true},
			ScreenName: req.ScreenName,
		}

		if req.TrustScore != nil {
			in.TrustScore = sql.NullInt16{Int16: *req.TrustScore, Valid: true}
		}
		if req.SeqScore != nil {
			in.SeqScore = sql.NullInt16{Int16: *req.SeqScore, Valid: true}
		}
		if req.AccuracyFeedback != nil && *req.AccuracyFeedback != "" {
			in.AccuracyFeedback = sql.NullString{String: *req.AccuracyFeedback, Valid: true}
		}
		if req.FeedbackText != nil && *req.FeedbackText != "" {
			in.FeedbackText = sql.NullString{String: *req.FeedbackText, Valid: true}
		}

		if req.ContextMetadata != nil {
			metaBytes, err := json.Marshal(req.ContextMetadata)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid context_metadata")
				return
			}
			in.ContextMetadata = pqtype.NullRawMessage{RawMessage: metaBytes, Valid: true}
		} else {
			in.ContextMetadata = pqtype.NullRawMessage{RawMessage: []byte("{}"), Valid: true}
		}

		row, err := webapp_service.CreateAnalysisFeedback(ctx, cfg.DB, in)
		if err != nil {
			if errors.Is(err, webapp_service.ErrFeedbackValidation) {
				utils.RespondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
			logger.Logger.Errorw("Failed to create feedback", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save feedback")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AnalysisFeedbackResponse{
			ID:     row.ID,
			Status: "success",
		})
	}
}
