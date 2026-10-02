package handlers

import (
	"context"

	"github.com/fouched/go-todo/internal/core/models"
)

type UserService interface {
	RegisterUser(ctx context.Context, email, password string, role models.Role) (*models.User, error)
	LoginUser(ctx context.Context, email, password string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

type TaskService interface {
	CreateTask(ctx context.Context, userID int64, task *models.Task) (*models.Task, error)
	GetTasksByUserAndCategory(ctx context.Context, userID int64, category string) ([]models.Task, error)
	UpdateTask(ctx context.Context, userID int64, t *models.Task) error
	DeleteTask(ctx context.Context, userID int64, id int64) error
	ToggleTaskCompletion(ctx context.Context, userID int64, taskID int64, isCompleted bool) (*models.Task, error)
}
