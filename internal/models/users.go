package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1"`
	LastName  string `json:"last_name" validate:"required,min=1"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// Pointer fields allow partial updates (nil = not provided).
type UpdateUserRequest struct {
	FirstName     *string          `json:"first_name,omitempty"`
	LastName      *string          `json:"last_name,omitempty"`
	Email         *string          `json:"email,omitempty"`
	Password      *string          `json:"password,omitempty"`
	MonthlyIncome *decimal.Decimal `json:"monthly_income,omitempty"`
}

type UpdateUserResponse struct {
	ID            uuid.UUID       `json:"id"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	MonthlyIncome decimal.Decimal `json:"monthly_income"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}
