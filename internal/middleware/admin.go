package middleware

import (
	"DebtEase/internal/api"
	"DebtEase/internal/utils"
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// IsAdminEmail returns true if email matches any configured admin (case-insensitive).
func IsAdminEmail(cfg *api.Config, email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(cfg.AdminEmails) == 0 {
		return false
	}
	for _, a := range cfg.AdminEmails {
		allowed := strings.TrimSpace(strings.Trim(a, `"'`))
		if allowed != "" && strings.EqualFold(allowed, email) {
			return true
		}
	}
	return false
}

// AdminRequired must run after AuthRequired; rejects non-allowlisted users with 403.
func AdminRequired(cfg *api.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(ContextUserIDKey)
		userID, ok := val.(uuid.UUID)
		if !ok || userID == uuid.Nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		user, err := cfg.DB.GetUserByID(r.Context(), userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		if !IsAdminEmail(cfg, user.Email) {
			utils.RespondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AdminAuth chains AuthRequired then AdminRequired.
func AdminAuth(cfg *api.Config, next http.Handler) http.Handler {
	return AuthRequired(cfg, AdminRequired(cfg, next))
}

// GetUserIDFromContext returns the authenticated user ID from context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(ContextUserIDKey)
	userID, ok := val.(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}
