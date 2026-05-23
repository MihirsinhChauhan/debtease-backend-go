package database

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserProfile is used by sqlc-generated user_profiles queries. The canonical
// user_profiles DDL is maintained outside bundled schema snippets; defining
// the struct here keeps sqlc codegen working without wiping this type.

type UserProfile struct {
	UserID         uuid.UUID
	RiskComfort    string
	FinancialGoals json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
