package api

import (
	"database/sql"
	"sync/atomic"

	"DebtEase/internal/database"
)

type Config struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	DBPool         *sql.DB
	Platform       string
	JWTSecret      string
	RabbitMQURL    string

	// Google OAuth (optional; enable by setting GOOGLE_CLIENT_ID etc. in env).
	GoogleClientID             string
	GoogleClientSecret         string
	GoogleRedirectURL          string // Must match an Authorized redirect URI in Google Cloud Console
	GoogleLoginSuccessRedirect string // After browser OAuth, optional frontend URL to receive tokens in query params
}
