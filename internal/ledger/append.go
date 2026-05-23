package ledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"DebtEase/internal/database"
	financial_events "DebtEase/internal/domain/financial_events/v1"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Params describes one append-only financial_events row.
type Params struct {
	UserID     uuid.UUID
	EventType  string
	EntityType string
	EntityID   uuid.UUID
	Amount     *decimal.Decimal
	Currency   string
	OccurredAt time.Time
	Metadata   map[string]interface{}
}

// Append inserts a financial_events row (must run inside the caller's transaction when paired with mutations).
func Append(ctx context.Context, q *database.Queries, p Params) error {
	meta := p.Metadata
	if meta == nil {
		meta = map[string]interface{}{}
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	var amt sql.NullString
	if p.Amount != nil {
		amt = sql.NullString{String: p.Amount.StringFixed(2), Valid: true}
	}

	entity := uuid.NullUUID{}
	if p.EntityID != uuid.Nil {
		entity = uuid.NullUUID{UUID: p.EntityID, Valid: true}
	}

	cur := p.Currency
	if cur == "" {
		cur = financial_events.DefaultCurrency
	}

	_, err = q.InsertFinancialEvent(ctx, database.InsertFinancialEventParams{
		UserID:        p.UserID,
		EventType:     p.EventType,
		EntityType:    p.EntityType,
		EntityID:      entity,
		Amount:        amt,
		Currency:      cur,
		OccurredAt:    p.OccurredAt.UTC(),
		Metadata:      raw,
		SchemaVersion: 1,
	})
	return err
}
