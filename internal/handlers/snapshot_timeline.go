package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"DebtEase/internal/api"
	"DebtEase/internal/database"
	health_v1 "DebtEase/internal/domain/health/v1"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_service "DebtEase/internal/webapp_specifics/service"

	"github.com/google/uuid"
)

// SnapshotTimelineEntry is one point in `/api/users/me/financial-snapshots/recent`.
type SnapshotTimelineEntry struct {
	SnapshotID     uuid.UUID `json:"snapshot_id"`
	CapturedAt     string    `json:"captured_at"`
	SnapshotType   string    `json:"snapshot_type"`
	SnapshotSource string    `json:"snapshot_source"`
	Status         string    `json:"status"`
	Score          int       `json:"score"`
	EngineVersion  string    `json:"engine_version"`
}

// HandleListRecentFinancialSnapshots lists recent persisted snapshots per user with health_v1 recomputed from row aggregates.
func HandleListRecentFinancialSnapshots(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		limit := int32(12)
		if raw := r.URL.Query().Get("limit"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 50 {
				utils.RespondWithError(w, http.StatusBadRequest, "limit must be 1–50")
				return
			}
			limit = int32(n)
		}

		ctx := context.Background()
		rows, err := cfg.DB.ListFinancialSnapshotsForUser(ctx, database.ListFinancialSnapshotsForUserParams{
			UserID: uuid.NullUUID{UUID: userID, Valid: true},
			Limit:  limit,
		})
		if err != nil {
			logger.Logger.Errorw("ListFinancialSnapshotsForUser failed", "error", err, "userID", userID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list snapshots")
			return
		}

		out := make([]SnapshotTimelineEntry, 0, len(rows))
		for _, snap := range rows {
			hin, err2 := webapp_service.HealthInputFromFinancialSnapshot(snap)
			if err2 != nil {
				logger.Logger.Warnw("HealthInputFromFinancialSnapshot skipped timeline row", "error", err2, "snapshot_id", snap.ID)
				continue
			}
			result := health_v1.Run(hin)
			out = append(out, SnapshotTimelineEntry{
				SnapshotID:     snap.ID,
				CapturedAt:     snap.CreatedAt.Format(time.RFC3339),
				SnapshotType:   snap.SnapshotType,
				SnapshotSource: snap.SnapshotSource,
				Status:         string(result.Status),
				Score:          result.Score,
				EngineVersion:  result.EngineVersion,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"snapshots": out,
			"limit":     limit,
		})
	}
}
