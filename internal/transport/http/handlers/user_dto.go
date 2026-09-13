package handlers

import "github.com/fouched/go-todo/internal/core/models"

type AuthRequest struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Role     models.Role `json:"role,omitempty"`
}

type UserResponse struct {
	ID    int64       `json:"id"`
	Email string      `json:"email"`
	Role  models.Role `json:"role"`
}
