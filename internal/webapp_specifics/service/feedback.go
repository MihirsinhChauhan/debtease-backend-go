package service

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

var (
	ErrFeedbackValidation = errors.New("feedback validation failed")
)

var allowedFeedbackScreens = map[string]bool{
	"debt_health_onboarding":   true,
	"loan_simulation_impact":   true,
	"financial_health_details": true,
}

var allowedAccuracy = map[string]bool{
	"Yes":    true,
	"Mostly": true,
	"No":     true,
}

type CreateFeedbackInput struct {
	UserID           uuid.NullUUID
	SessionID        uuid.NullUUID
	TrustScore       sql.NullInt16
	SeqScore         sql.NullInt16
	AccuracyFeedback sql.NullString
	FeedbackText     sql.NullString
	ScreenName       string
	ContextMetadata  pqtype.NullRawMessage
}

func ValidateFeedbackInput(in *CreateFeedbackInput) error {
	if in.ScreenName == "" || !allowedFeedbackScreens[in.ScreenName] {
		return fmt.Errorf("%w: invalid screen_name", ErrFeedbackValidation)
	}

	hasTrust := in.TrustScore.Valid
	hasAccuracy := in.AccuracyFeedback.Valid && in.AccuracyFeedback.String != ""
	hasSeq := in.SeqScore.Valid
	if !hasTrust && !hasAccuracy && !hasSeq {
		return fmt.Errorf("%w: at least one of trust_score, accuracy_feedback, or seq_score is required", ErrFeedbackValidation)
	}

	if hasTrust {
		v := in.TrustScore.Int16
		if v < 1 || v > 7 {
			return fmt.Errorf("%w: trust_score must be between 1 and 7", ErrFeedbackValidation)
		}
	}
	if hasSeq {
		v := in.SeqScore.Int16
		if v < 1 || v > 7 {
			return fmt.Errorf("%w: seq_score must be between 1 and 7", ErrFeedbackValidation)
		}
	}
	if hasAccuracy {
		if !allowedAccuracy[in.AccuracyFeedback.String] {
			return fmt.Errorf("%w: accuracy_feedback must be Yes, Mostly, or No", ErrFeedbackValidation)
		}
	}

	if in.FeedbackText.Valid && in.FeedbackText.String != "" {
		allowText := false
		if hasTrust && in.TrustScore.Int16 <= 4 {
			allowText = true
		}
		if hasAccuracy && in.AccuracyFeedback.String == "No" {
			allowText = true
		}
		if !allowText {
			return fmt.Errorf("%w: feedback_text only allowed when trust_score <= 4 or accuracy_feedback is No", ErrFeedbackValidation)
		}
	}

	return nil
}

func CreateAnalysisFeedback(ctx context.Context, db *database.Queries, in CreateFeedbackInput) (database.AnalysisFeedback, error) {
	if err := ValidateFeedbackInput(&in); err != nil {
		return database.AnalysisFeedback{}, err
	}

	meta := in.ContextMetadata
	if !meta.Valid {
		meta = pqtype.NullRawMessage{RawMessage: []byte("{}"), Valid: true}
	}

	row, err := db.CreateAnalysisFeedback(ctx, database.CreateAnalysisFeedbackParams{
		UserID:           in.UserID,
		SessionID:        in.SessionID,
		SeqScore:         in.SeqScore,
		TrustScore:       in.TrustScore,
		AccuracyFeedback: in.AccuracyFeedback,
		FeedbackText:     in.FeedbackText,
		ScreenName:       in.ScreenName,
		ContextMetadata:  meta,
	})
	if err != nil {
		logger.Logger.Errorw("Failed to insert analysis feedback", "error", err, "screen_name", in.ScreenName)
		return database.AnalysisFeedback{}, err
	}

	logger.Logger.Infow("Created analysis feedback", "id", row.ID, "screen_name", in.ScreenName)
	return row, nil
}
