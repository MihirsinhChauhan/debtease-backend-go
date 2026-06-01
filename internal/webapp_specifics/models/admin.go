package models

type AdminMeResponse struct {
	IsAdmin bool   `json:"is_admin"`
	Email   string `json:"email"`
}

type AccuracyCount struct {
	AccuracyFeedback string `json:"accuracy_feedback"`
	Count            int64  `json:"count"`
}

type AdminOverviewResponse struct {
	Days                  int               `json:"days"`
	Since                 string            `json:"since"`
	SessionsStarted       int64             `json:"sessions_started"`
	SnapshotsSubmitted    int64             `json:"snapshots_submitted"`
	Signups               int64             `json:"signups"`
	FeedbackResponses     int64             `json:"feedback_responses"`
	AvgTrust              *float64          `json:"avg_trust,omitempty"`
	AvgSeq                *float64          `json:"avg_seq,omitempty"`
	AccuracyDistribution  []AccuracyCount   `json:"accuracy_distribution"`
}

type AdminFeedbackRow struct {
	ID               string  `json:"id"`
	CreatedAt        string  `json:"created_at"`
	UserID           *string `json:"user_id,omitempty"`
	UserEmail        *string `json:"user_email,omitempty"`
	SessionID        *string `json:"session_id,omitempty"`
	ScreenName       string  `json:"screen_name"`
	TrustScore       *int16  `json:"trust_score,omitempty"`
	SeqScore         *int16  `json:"seq_score,omitempty"`
	AccuracyFeedback *string `json:"accuracy_feedback,omitempty"`
	FeedbackText     *string `json:"feedback_text,omitempty"`
	ContextMetadata  map[string]interface{} `json:"context_metadata,omitempty"`
}

type AdminFeedbackListResponse struct {
	Items  []AdminFeedbackRow `json:"items"`
	Total  int64              `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

type ScreenNameCount struct {
	ScreenName string `json:"screen_name"`
	Count      int64  `json:"count"`
}

type MetadataTrustAvg struct {
	Key   string  `json:"key"`
	Value string  `json:"value"`
	Avg   float64 `json:"avg"`
	Count int64   `json:"count"`
}

type AdminFeedbackSummaryResponse struct {
	Total            int64              `json:"total"`
	AvgTrust         *float64           `json:"avg_trust,omitempty"`
	AvgSeq           *float64           `json:"avg_seq,omitempty"`
	AccuracyDistribution []AccuracyCount `json:"accuracy_distribution"`
	ByScreenName     []ScreenNameCount  `json:"by_screen_name"`
	TrustByScoreBand []MetadataTrustAvg `json:"trust_by_score_band"`
	TrustByHealthStatus []MetadataTrustAvg `json:"trust_by_health_status"`
}

type AdminUserSearchResult struct {
	UserID    *string `json:"user_id,omitempty"`
	Email     *string `json:"email,omitempty"`
	SessionID *string `json:"session_id,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
}

type AdminUserSearchResponse struct {
	Results []AdminUserSearchResult `json:"results"`
}

type AdminSnapshotView struct {
	SnapshotID     string   `json:"snapshot_id"`
	CapturedAt     string   `json:"captured_at"`
	SnapshotType   string   `json:"snapshot_type"`
	Status         string   `json:"status"`
	Score          int      `json:"score"`
	MonthlyIncome  string   `json:"monthly_income,omitempty"`
	TotalMonthlyEMI string  `json:"total_monthly_emi,omitempty"`
	ActiveDebtCount *int    `json:"active_debt_count,omitempty"`
	EngineVersion  string   `json:"engine_version"`
}

type AdminDebtSummary struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	Type                    string  `json:"type"`
	OutstandingBalance      string  `json:"outstanding_balance"`
	TreatmentClassification *string `json:"treatment_classification,omitempty"`
}

type AdminCalculationEvent struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	CreatedAt string                 `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type AdminUserDetailResponse struct {
	UserID    string              `json:"user_id"`
	Email     string              `json:"email"`
	FirstName string              `json:"first_name"`
	LastName  string              `json:"last_name"`
	CreatedAt string              `json:"created_at"`
	Snapshots []AdminSnapshotView `json:"snapshots"`
	Debts     []AdminDebtSummary  `json:"debts"`
	Feedback  []AdminFeedbackRow  `json:"feedback"`
}

type AdminSessionDetailResponse struct {
	SessionID string              `json:"session_id"`
	Snapshots []AdminSnapshotView `json:"snapshots"`
	Events    []AdminCalculationEvent `json:"events"`
	Feedback  []AdminFeedbackRow  `json:"feedback"`
}
