package handlers

import (
	"context"

	"github.com/fouched/go-todo/internal/core/models"
)

type TaskService interface {
	CreateTask(ctx context.Context, userID int64, task *models.Task) (*models.Task, error)
	GetTaskByID(ctx context.Context, id int64) (*models.Task, error)
	GetTasksForUser(ctx context.Context, userID int64) ([]models.Task, error)
	UpdateTask(ctx context.Context, t *models.Task) error
	DeleteTask(ctx context.Context, id int64) error
}
