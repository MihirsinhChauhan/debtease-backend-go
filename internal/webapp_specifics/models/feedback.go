package models

import "github.com/google/uuid"

type AnalysisFeedbackRequest struct {
	SessionID        *uuid.UUID             `json:"session_id,omitempty"`
	TrustScore       *int16                 `json:"trust_score,omitempty"`
	AccuracyFeedback *string                `json:"accuracy_feedback,omitempty"`
	SeqScore         *int16                 `json:"seq_score,omitempty"`
	FeedbackText     *string                `json:"feedback_text,omitempty"`
	ScreenName       string                 `json:"screen_name"`
	ContextMetadata  map[string]interface{} `json:"context_metadata,omitempty"`
}

type AnalysisFeedbackResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}
