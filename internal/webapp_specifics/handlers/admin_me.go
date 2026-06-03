package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/middleware"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	"net/http"
)

func HandleAdminMe(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		user, err := cfg.DB.GetUserByID(r.Context(), userID)
		if err != nil {
			logger.Logger.Warnw("Admin me: user lookup failed", "error", err)
			utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminMeResponse{
				IsAdmin: false,
				Email:   "",
			})
			return
		}

		isAdmin := middleware.IsAdminEmail(cfg, user.Email)
		logger.Logger.Infow("Admin me check", "email", user.Email, "is_admin", isAdmin, "allowlist_count", len(cfg.AdminEmails))

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminMeResponse{
			IsAdmin: isAdmin,
			Email:   user.Email,
		})
	}
}
