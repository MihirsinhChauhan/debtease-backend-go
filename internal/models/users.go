package models

import (
	"github.com/google/uuid"
)
type CreateUserRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	MonthlyIncome string `json:"monthly_income"`
}

type UpdateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	MonthlyIncome string `json:"monthly_income"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}