package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"DebtEase/internal/database"
	health_v1 "DebtEase/internal/domain/health/v1"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	webapp_service "DebtEase/internal/webapp_specifics/service"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

func parseAdminDays(r *http.Request) int {
	days := 7
	if raw := r.URL.Query().Get("days"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 && n <= 90 {
			days = n
		}
	}
	return days
}

func sinceFromDays(days int) time.Time {
	return time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
}

func parseSinceQuery(r *http.Request) sql.NullTime {
	if raw := r.URL.Query().Get("since"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			return sql.NullTime{Time: t.UTC(), Valid: true}
		}
	}
	return sql.NullTime{}
}

func parseFeedbackFilterParams(r *http.Request) database.ListAnalysisFeedbackParams {
	limit := int32(50)
	offset := int32(0)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 && n <= 200 {
			limit = int32(n)
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	params := database.ListAnalysisFeedbackParams{
		Limit:  limit,
		Offset: offset,
	}

	if since := parseSinceQuery(r); since.Valid {
		params.Since = since
	} else if daysRaw := r.URL.Query().Get("days"); daysRaw != "" {
		if days, err := strconv.Atoi(daysRaw); err == nil && days >= 1 && days <= 90 {
			params.Since = sql.NullTime{Time: sinceFromDays(days), Valid: true}
		}
	}

	if sn := r.URL.Query().Get("screen_name"); sn != "" {
		params.ScreenName = sql.NullString{String: sn, Valid: true}
	}
	if r.URL.Query().Get("priority_only") == "true" || r.URL.Query().Get("priority_only") == "1" {
		params.PriorityOnly = sql.NullBool{Bool: true, Valid: true}
	}
	if raw := r.URL.Query().Get("trust_lt"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			params.TrustLt = sql.NullInt16{Int16: int16(n), Valid: true}
		}
	}
	if acc := r.URL.Query().Get("accuracy"); acc != "" {
		params.Accuracy = sql.NullString{String: acc, Valid: true}
	}

	return params
}

func countFeedbackParams(list database.ListAnalysisFeedbackParams) database.CountAnalysisFeedbackParams {
	return database.CountAnalysisFeedbackParams{
		Since:        list.Since,
		ScreenName:   list.ScreenName,
		PriorityOnly: list.PriorityOnly,
		TrustLt:      list.TrustLt,
		Accuracy:     list.Accuracy,
	}
}

func mapFeedbackListRow(row database.ListAnalysisFeedbackRow) webapp_models.AdminFeedbackRow {
	return mapFeedbackRow(
		row.ID,
		row.CreatedAt,
		row.UserID,
		row.UserEmail,
		row.SessionID,
		row.ScreenName,
		row.TrustScore,
		row.SeqScore,
		row.AccuracyFeedback,
		row.FeedbackText,
		row.ContextMetadata,
	)
}

func mapFeedbackByUserRow(row database.ListFeedbackByUserIDRow) webapp_models.AdminFeedbackRow {
	return mapFeedbackRow(
		row.ID,
		row.CreatedAt,
		row.UserID,
		row.UserEmail,
		row.SessionID,
		row.ScreenName,
		row.TrustScore,
		row.SeqScore,
		row.AccuracyFeedback,
		row.FeedbackText,
		row.ContextMetadata,
	)
}

func mapFeedbackBySessionRow(row database.ListFeedbackBySessionIDRow) webapp_models.AdminFeedbackRow {
	return mapFeedbackRow(
		row.ID,
		row.CreatedAt,
		row.UserID,
		row.UserEmail,
		row.SessionID,
		row.ScreenName,
		row.TrustScore,
		row.SeqScore,
		row.AccuracyFeedback,
		row.FeedbackText,
		row.ContextMetadata,
	)
}

func mapFeedbackRow(
	id uuid.UUID,
	createdAt time.Time,
	userID uuid.NullUUID,
	userEmail sql.NullString,
	sessionID uuid.NullUUID,
	screenName string,
	trust sql.NullInt16,
	seq sql.NullInt16,
	accuracy sql.NullString,
	feedbackText sql.NullString,
	meta pqtype.NullRawMessage,
) webapp_models.AdminFeedbackRow {
	out := webapp_models.AdminFeedbackRow{
		ID:         id.String(),
		CreatedAt:  createdAt.UTC().Format(time.RFC3339),
		ScreenName: screenName,
	}
	if userID.Valid {
		s := userID.UUID.String()
		out.UserID = &s
	}
	if userEmail.Valid {
		s := userEmail.String
		out.UserEmail = &s
	}
	if sessionID.Valid {
		s := sessionID.UUID.String()
		out.SessionID = &s
	}
	if trust.Valid {
		out.TrustScore = &trust.Int16
	}
	if seq.Valid {
		out.SeqScore = &seq.Int16
	}
	if accuracy.Valid {
		s := accuracy.String
		out.AccuracyFeedback = &s
	}
	if feedbackText.Valid && feedbackText.String != "" {
		s := feedbackText.String
		out.FeedbackText = &s
	}
	if meta.Valid && len(meta.RawMessage) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(meta.RawMessage, &m); err == nil {
			out.ContextMetadata = m
		}
	}
	return out
}

func snapshotToAdminView(snap database.FinancialSnapshot) (webapp_models.AdminSnapshotView, bool) {
	hin, err := webapp_service.HealthInputFromFinancialSnapshot(snap)
	if err != nil {
		return webapp_models.AdminSnapshotView{}, false
	}
	result := health_v1.Run(hin)
	active := hin.ActiveDebtCount
	var activePtr *int
	if active != nil {
		activePtr = active
	}
	return webapp_models.AdminSnapshotView{
		SnapshotID:      snap.ID.String(),
		CapturedAt:      snap.CreatedAt.UTC().Format(time.RFC3339),
		SnapshotType:    snap.SnapshotType,
		Status:          string(result.Status),
		Score:           result.Score,
		MonthlyIncome:   snap.MonthlyIncome.String(),
		TotalMonthlyEMI: snap.TotalMonthlyEmi.String(),
		ActiveDebtCount: activePtr,
		EngineVersion:   result.EngineVersion,
	}, true
}

func avgFromTotals(total int64, avg float64) *float64 {
	if total == 0 {
		return nil
	}
	return &avg
}

func metadataValueString(v interface{}) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, t != ""
	case []byte:
		return string(t), len(t) > 0
	case sql.NullString:
		return t.String, t.Valid && t.String != ""
	default:
		return "", false
	}
}
