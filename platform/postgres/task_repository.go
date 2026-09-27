package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/repositories"
	"github.com/fouched/toolkit/v2/faults"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewTaskRepository(db *pgxpool.Pool, logger *slog.Logger) *TaskRepository {
	return &TaskRepository{
		db:     db,
		logger: logger,
	}
}

func (r *TaskRepository) Create(ctx context.Context, t *models.Task) error {
	query := `
        INSERT INTO tasks (title, description, category, is_completed, user_id)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

	err := r.db.QueryRow(ctx, query,
		t.Title,
		t.Description,
		t.Category,
		t.IsCompleted,
		t.UserID,
	).Scan(&t.ID)

	return faults.Wrap(err, "failed to create task")
}

func (r *TaskRepository) FindByID(ctx context.Context, id int64) (*models.Task, error) {
	query := `
        SELECT id, title, description, category, is_completed, user_id
        FROM tasks
        WHERE id = $1
    `

	var t models.Task

	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Category,
		&t.IsCompleted,
		&t.UserID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Annotate gives context without wiping out the core Sentinel identity!
			return nil, faults.Annotate(repositories.ErrNotFound, fmt.Sprintf("task %d row not found", id))
		}
		// Wrap captures the stack track immediately for real, unexpected DB faults
		return nil, faults.Wrap(err, "database query execution failed")
	}

	return &t, nil
}

func (r *TaskRepository) FindAllByUser(ctx context.Context, userID int64) ([]models.Task, error) {
	query := `
        SELECT id, title, description, category, is_completed, user_id
        FROM tasks
        WHERE user_id = $1
        ORDER BY id
    `

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, faults.Wrap(err, "failed to find all tasks for user")
	}
	defer rows.Close()

	tasks := []models.Task{}

	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Category,
			&t.IsCompleted,
			&t.UserID,
		); err != nil {
			return nil, faults.Wrap(err, "failed to scan all tasks for user")
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) FindAllByUserAndCategory(ctx context.Context, userID int64, category string) ([]models.Task, error) {
	query := `
        SELECT id, title, description, category, is_completed, user_id
        FROM tasks
        WHERE user_id = $1 AND category = $2
        ORDER BY id
    `

	rows, err := r.db.Query(ctx, query, userID, category)
	if err != nil {
		return nil, faults.Wrap(err, "failed to find all tasks for user by category")
	}
	defer rows.Close()

	tasks := []models.Task{}

	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Category,
			&t.IsCompleted,
			&t.UserID,
		); err != nil {
			return nil, faults.Wrap(err, "failed to scan all tasks for user by category")
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, t *models.Task) error {
	query := `
        UPDATE tasks
        SET title = $1,
            description = $2,
            category = $3,
            is_completed = $4
        WHERE id = $5
    `

	cmd, err := r.db.Exec(ctx, query,
		t.Title,
		t.Description,
		t.Category,
		t.IsCompleted,
		t.ID,
	)

	if err != nil {
		return faults.Wrap(err, "failed to update task")
	}

	if cmd.RowsAffected() == 0 {
		return faults.Annotate(repositories.ErrNotFound, "failed to update task")
	}

	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1`

	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return faults.Wrap(err, "failed to delete task")
	}

	if cmd.RowsAffected() == 0 {
		return faults.Annotate(repositories.ErrNotFound, fmt.Sprintf("failed to delete task %d row not found", id))
	}

	return nil
}
