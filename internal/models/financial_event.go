package models

import (
	"encoding/json"
	"time"

	"DebtEase/internal/database"
)

// FinancialEventResponse is the JSON shape for GET /api/users/me/financial-events.
type FinancialEventResponse struct {
	ID            string          `json:"id"`
	EventType     string          `json:"event_type"`
	EntityType    string          `json:"entity_type"`
	EntityID      *string         `json:"entity_id,omitempty"`
	Amount        *string         `json:"amount,omitempty"`
	Currency      string          `json:"currency"`
	OccurredAt    time.Time       `json:"occurred_at"`
	CreatedAt     time.Time       `json:"created_at"`
	SchemaVersion int16           `json:"schema_version"`
	Metadata      json.RawMessage `json:"metadata"`
}

func FinancialEventResponseFromDB(e database.FinancialEvent) FinancialEventResponse {
	out := FinancialEventResponse{
		ID:            e.ID.String(),
		EventType:     e.EventType,
		EntityType:    e.EntityType,
		Currency:      e.Currency,
		OccurredAt:    e.OccurredAt,
		CreatedAt:     e.CreatedAt,
		SchemaVersion: e.SchemaVersion,
		Metadata:      e.Metadata,
	}
	if e.EntityID.Valid {
		s := e.EntityID.UUID.String()
		out.EntityID = &s
	}
	if e.Amount.Valid && e.Amount.String != "" {
		s := e.Amount.String
		out.Amount = &s
	}
	return out
}
