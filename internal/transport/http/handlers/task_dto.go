package handlers

import "github.com/fouched/go-todo/internal/core/models"

type CreateTaskRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Category    models.Category `json:"category"`
}

type UpdateTaskRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Category    models.Category `json:"category"`
	IsCompleted bool            `json:"is_completed"`
}

type TaskResponse struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Category    models.Category `json:"category"`
	IsCompleted bool            `json:"is_completed"`
}
