package handlers

import (
	"context"
	"net/http"
	"strconv"

	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/models"
	"DebtEase/internal/utils"
)

func HandleListMyFinancialEvents(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		limit := 50
		if raw := r.URL.Query().Get("limit"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 200 {
				utils.RespondWithError(w, http.StatusBadRequest, "limit must be 1–200")
				return
			}
			limit = n
		}
		offset := 0
		if raw := r.URL.Query().Get("offset"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 0 {
				utils.RespondWithError(w, http.StatusBadRequest, "offset must be >= 0")
				return
			}
			offset = n
		}

		ctx := context.Background()
		items, err := cfg.DB.ListFinancialEventsByUser(ctx, database.ListFinancialEventsByUserParams{
			UserID: userID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			logger.Logger.Errorw("ListFinancialEventsByUser failed", "error", err, "userID", userID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list financial events")
			return
		}

		out := make([]models.FinancialEventResponse, 0, len(items))
		for _, e := range items {
			out = append(out, models.FinancialEventResponseFromDB(e))
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"events": out,
			"limit":  limit,
			"offset": offset,
		})
	}
}
