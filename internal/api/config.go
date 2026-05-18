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
}
