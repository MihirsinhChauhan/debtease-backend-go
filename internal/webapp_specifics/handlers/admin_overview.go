package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	"database/sql"
	"net/http"
)

func HandleAdminOverview(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		days := parseAdminDays(r)
		since := sinceFromDays(days)

		sessions, err := cfg.DB.CountSessionsSince(ctx, since)
		if err != nil {
			logger.Logger.Errorw("CountSessionsSince failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load overview")
			return
		}
		snapshots, err := cfg.DB.CountOnboardingSnapshotsSince(ctx, since)
		if err != nil {
			logger.Logger.Errorw("CountOnboardingSnapshotsSince failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load overview")
			return
		}
		signups, err := cfg.DB.CountSignupsSince(ctx, since)
		if err != nil {
			logger.Logger.Errorw("CountSignupsSince failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load overview")
			return
		}
		feedbackCount, err := cfg.DB.CountFeedbackSince(ctx, since)
		if err != nil {
			logger.Logger.Errorw("CountFeedbackSince failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load overview")
			return
		}

		sinceNull := sql.NullTime{Time: since, Valid: true}
		var avgTrust, avgSeq *float64
		if totals, err := cfg.DB.FeedbackSummaryTotals(ctx, sinceNull); err == nil {
			avgTrust = avgFromTotals(totals.Total, totals.AvgTrust)
			avgSeq = avgFromTotals(totals.Total, totals.AvgSeq)
		}

		accRows, err := cfg.DB.FeedbackAccuracyDistributionSince(ctx, since)
		if err != nil {
			logger.Logger.Errorw("FeedbackAccuracyDistributionSince failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load overview")
			return
		}
		accDist := make([]webapp_models.AccuracyCount, 0, len(accRows))
		for _, row := range accRows {
			if !row.AccuracyFeedback.Valid {
				continue
			}
			accDist = append(accDist, webapp_models.AccuracyCount{
				AccuracyFeedback: row.AccuracyFeedback.String,
				Count:            row.Count,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminOverviewResponse{
			Days:                 days,
			Since:                since.Format("2006-01-02T15:04:05Z07:00"),
			SessionsStarted:      sessions,
			SnapshotsSubmitted:   snapshots,
			Signups:              signups,
			FeedbackResponses:    feedbackCount,
			AvgTrust:             avgTrust,
			AvgSeq:               avgSeq,
			AccuracyDistribution: accDist,
		})
	}
}
