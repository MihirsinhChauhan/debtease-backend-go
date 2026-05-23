package service

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// SessionMeta is optional funnel metadata stored when creating a new calculation session.
type SessionMeta struct {
	EntryPoint string
	UserID     uuid.UUID
}

func GetOrCreateSession(ctx context.Context, db *database.Queries, sessionID uuid.UUID, engineVersion string, meta *SessionMeta) (uuid.UUID, error) {
	if sessionID == uuid.Nil {
		sessionID = uuid.New()
	}

	exists, err := db.SessionExists(ctx, sessionID)
	if err != nil {
		logger.Logger.Errorw("Failed to check session existence", "error", err)
		return uuid.Nil, err
	}

	if !exists {
		if engineVersion == "" {
			engineVersion = "v1"
		}
		var uid uuid.NullUUID
		var ep sql.NullString
		if meta != nil {
			if meta.UserID != uuid.Nil {
				uid = uuid.NullUUID{UUID: meta.UserID, Valid: true}
			}
			if meta.EntryPoint != "" {
				ep = sql.NullString{String: meta.EntryPoint, Valid: true}
			}
		}
		_, err := db.CreateCalculationSession(ctx, database.CreateCalculationSessionParams{
			SessionID:     sessionID,
			EngineVersion: engineVersion,
			UserID:        uid,
			EntryPoint:    ep,
		})
		if err != nil {
			logger.Logger.Errorw("Failed to create calculation session", "error", err, "session_id", sessionID)
			return uuid.Nil, err
		}
		logger.Logger.Infow("Created new calculation session", "session_id", sessionID, "engine_version", engineVersion)
	}

	return sessionID, nil
}

// GetSessionByID loads a calculation session by its public session_id UUID (cookie / query).
func GetSessionByID(ctx context.Context, db *database.Queries, sessionID uuid.UUID) (database.CalculationSession, error) {
	return db.GetCalculationSessionByID(ctx, sessionID)
}

func SessionExists(ctx context.Context, db *database.Queries, sessionID uuid.UUID) (bool, error) {
	return db.SessionExists(ctx, sessionID)
}
