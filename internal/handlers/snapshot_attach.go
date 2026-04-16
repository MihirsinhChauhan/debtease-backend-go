package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/models"
	"DebtEase/internal/utils"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// HandleSnapshotAttach links the latest anonymous session snapshot to the authenticated user.
func HandleSnapshotAttach(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing session cookie; complete onboarding in this browser first")
			return
		}

		row, err := cfg.DB.AttachLatestOrphanSnapshotBySessionToUser(r.Context(), database.AttachLatestOrphanSnapshotBySessionToUserParams{
			UserID:    uuid.NullUUID{UUID: userID, Valid: true},
			SessionID: uuid.NullUUID{UUID: sessionID, Valid: true},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				utils.RespondWithJSON(w, http.StatusOK, models.SnapshotAttachResponse{Attached: false})
				return
			}
			logger.Logger.Errorw("AttachLatestOrphanSnapshotBySessionToUser failed", "error", err, "userID", userID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to attach snapshot")
			return
		}

		sid := row.ID
		utils.RespondWithJSON(w, http.StatusOK, models.SnapshotAttachResponse{
			Attached:   true,
			SnapshotID: &sid,
		})
	}
}
