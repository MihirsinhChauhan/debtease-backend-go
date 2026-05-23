package handlers

import (
	"context"
	"time"

	"DebtEase/internal/api"
	"DebtEase/internal/auth"
	"DebtEase/internal/database"
	"DebtEase/internal/models"
	"DebtEase/internal/utils"
)

// buildLoginSession issues access + refresh JWTs and persists refresh (same semantics as POST /api/login).
func buildLoginSession(ctx context.Context, cfg *api.Config, user database.User) (models.LoginResponse, error) {
	accessToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		return models.LoginResponse{}, err
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return models.LoginResponse{}, err
	}

	expiresAt := time.Now().Add(60 * 24 * time.Hour)
	err = cfg.DB.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return models.LoginResponse{}, err
	}

	resp := models.LoginResponse{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Name:         utils.JoinPersonName(user.FirstName, user.LastName),
		Email:        user.Email,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
