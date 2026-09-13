package services

import (
	"context"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/repositories"
)

type TaskService struct {
	tasks repositories.TaskRepository
}

func NewTaskService(tasks repositories.TaskRepository) *TaskService {
	return &TaskService{tasks: tasks}
}

func (s *TaskService) CreateTask(ctx context.Context, userID int64, req *models.Task) (*models.Task, error) {
	req.UserID = userID
	req.IsCompleted = false

	if err := s.tasks.Create(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int64) (*models.Task, error) {
	return s.tasks.FindByID(ctx, id)
}

func (s *TaskService) GetTasksForUser(ctx context.Context, userID int64) ([]models.Task, error) {
	return s.tasks.FindAllByUser(ctx, userID)
}

func (s *TaskService) UpdateTask(ctx context.Context, t *models.Task) error {
	return s.tasks.Update(ctx, t)
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	return s.tasks.Delete(ctx, id)
}
