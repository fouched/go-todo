package repositories

import (
	"context"

	"github.com/fouched/go-todo/internal/core/models"
)

type TaskRepository interface {
	Create(ctx context.Context, t *models.Task) error
	FindByID(ctx context.Context, id int64) (*models.Task, error)
	FindAllByUser(ctx context.Context, userID int64) ([]models.Task, error)
	Update(ctx context.Context, t *models.Task) error
	Delete(ctx context.Context, id int64) error
}
