package services

import (
	"context"
	"log/slog"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/repositories"
)

type TaskService struct {
	tasks  repositories.TaskRepository
	logger *slog.Logger
}

func NewTaskService(tasks repositories.TaskRepository, logger *slog.Logger) *TaskService {
	return &TaskService{
		tasks:  tasks,
		logger: logger,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, userID int64, req *models.Task) (*models.Task, error) {
	req.UserID = userID
	req.IsCompleted = false

	if err := s.tasks.Create(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *TaskService) GetTasksByUserAndCategory(ctx context.Context, userID int64, category string) ([]models.Task, error) {
	if category == "" {
		return s.tasks.FindAllByUser(ctx, userID)
	}
	return s.tasks.FindAllByUserAndCategory(ctx, userID, category)
}

func (s *TaskService) UpdateTask(ctx context.Context, userID int64, t *models.Task) error {
	return s.tasks.Update(ctx, t)
}

func (s *TaskService) DeleteTask(ctx context.Context, userID int64, id int64) error {
	return s.tasks.Delete(ctx, userID, id)
}

func (s *TaskService) ToggleTaskCompletion(ctx context.Context, userID int64, taskID int64, isCompleted bool) (*models.Task, error) {
	return s.tasks.UpdateCompletionStatus(ctx, userID, taskID, isCompleted)
}
