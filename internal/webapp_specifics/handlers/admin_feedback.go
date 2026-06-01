package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	"database/sql"
	"net/http"
)

func HandleAdminFeedbackList(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := parseFeedbackFilterParams(r)
		countParams := countFeedbackParams(params)

		total, err := cfg.DB.CountAnalysisFeedback(ctx, countParams)
		if err != nil {
			logger.Logger.Errorw("CountAnalysisFeedback failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list feedback")
			return
		}

		rows, err := cfg.DB.ListAnalysisFeedback(ctx, params)
		if err != nil {
			logger.Logger.Errorw("ListAnalysisFeedback failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list feedback")
			return
		}

		items := make([]webapp_models.AdminFeedbackRow, 0, len(rows))
		for _, row := range rows {
			items = append(items, mapFeedbackListRow(row))
		}

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminFeedbackListResponse{
			Items:  items,
			Total:  total,
			Limit:  int(params.Limit),
			Offset: int(params.Offset),
		})
	}
}

func HandleAdminFeedbackSummary(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		since := parseSinceQuery(r)
		if !since.Valid {
			days := parseAdminDays(r)
			since = sql.NullTime{Time: sinceFromDays(days), Valid: true}
		}

		totals, err := cfg.DB.FeedbackSummaryTotals(ctx, since)
		if err != nil {
			logger.Logger.Errorw("FeedbackSummaryTotals failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load feedback summary")
			return
		}

		accRows, err := cfg.DB.FeedbackSummaryAccuracy(ctx, since)
		if err != nil {
			logger.Logger.Errorw("FeedbackSummaryAccuracy failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load feedback summary")
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

		screenRows, err := cfg.DB.FeedbackSummaryByScreen(ctx, since)
		if err != nil {
			logger.Logger.Errorw("FeedbackSummaryByScreen failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load feedback summary")
			return
		}
		byScreen := make([]webapp_models.ScreenNameCount, 0, len(screenRows))
		for _, row := range screenRows {
			byScreen = append(byScreen, webapp_models.ScreenNameCount{
				ScreenName: row.ScreenName,
				Count:      row.Count,
			})
		}

		bandRows, err := cfg.DB.FeedbackTrustByScoreBand(ctx, since)
		if err != nil {
			logger.Logger.Errorw("FeedbackTrustByScoreBand failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load feedback summary")
			return
		}
		trustByBand := make([]webapp_models.MetadataTrustAvg, 0, len(bandRows))
		for _, row := range bandRows {
			val, ok := metadataValueString(row.MetadataValue)
			if !ok {
				continue
			}
			trustByBand = append(trustByBand, webapp_models.MetadataTrustAvg{
				Key:   "score_band",
				Value: val,
				Avg:   row.AvgTrust,
				Count: row.Count,
			})
		}

		healthRows, err := cfg.DB.FeedbackTrustByHealthStatus(ctx, since)
		if err != nil {
			logger.Logger.Errorw("FeedbackTrustByHealthStatus failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load feedback summary")
			return
		}
		trustByHealth := make([]webapp_models.MetadataTrustAvg, 0, len(healthRows))
		for _, row := range healthRows {
			val, ok := metadataValueString(row.MetadataValue)
			if !ok {
				continue
			}
			trustByHealth = append(trustByHealth, webapp_models.MetadataTrustAvg{
				Key:   "health_status",
				Value: val,
				Avg:   row.AvgTrust,
				Count: row.Count,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, webapp_models.AdminFeedbackSummaryResponse{
			Total:                totals.Total,
			AvgTrust:             avgFromTotals(totals.Total, totals.AvgTrust),
			AvgSeq:               avgFromTotals(totals.Total, totals.AvgSeq),
			AccuracyDistribution: accDist,
			ByScreenName:         byScreen,
			TrustByScoreBand:     trustByBand,
			TrustByHealthStatus:  trustByHealth,
		})
	}
}
