// Package dbconn opens PostgreSQL via pgx with simple query protocol so pooled
// connections (PgBouncer, Neon/Railway/Supabase pooler) do not collide on
// lib/pq unnamed prepared statements.
package dbconn

import (
	"database/sql"
	"net/url"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const simpleProtocolParam = "default_query_exec_mode"
const simpleProtocolValue = "simple_protocol"

// WithSimpleProtocol appends default_query_exec_mode=simple_protocol when absent.
func WithSimpleProtocol(dbURL string) string {
	if strings.Contains(dbURL, simpleProtocolParam+"=") {
		return dbURL
	}
	u, err := url.Parse(dbURL)
	if err != nil {
		sep := "?"
		if strings.Contains(dbURL, "?") {
			sep = "&"
		}
		return dbURL + sep + simpleProtocolParam + "=" + simpleProtocolValue
	}
	q := u.Query()
	if q.Get(simpleProtocolParam) == "" {
		q.Set(simpleProtocolParam, simpleProtocolValue)
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// Open returns a *sql.DB using the pgx driver and simple query protocol.
func Open(dbURL string) (*sql.DB, error) {
	return sql.Open("pgx", WithSimpleProtocol(dbURL))
}
