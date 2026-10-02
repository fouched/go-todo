package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/repositories"
	"github.com/fouched/toolkit/v2/faults"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger *slog.Logger) repositories.UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	query := `
        INSERT INTO users (email, password, role)
        VALUES ($1, $2, $3)
        RETURNING id
    `

	err := r.db.QueryRow(ctx, query,
		u.Email,
		u.Password,
		u.Role,
	).Scan(&u.ID)

	if err != nil {
		// Unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return faults.Wrap(repositories.ErrDuplicateEmail, "failed to create user")
			}
		}
	}

	return faults.Wrap(err, "failed to create user")
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, password, role
		FROM users
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, faults.Wrap(err, "failed to find all users")
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.Role); err != nil {
			return nil, faults.Wrap(err, "failed to scan user")
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, faults.Wrap(err, "failed to iterate over users")
	}

	return users, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
        SELECT id, email, password, role
        FROM users
        WHERE id = $1
    `

	var u models.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Password,
		&u.Role,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, faults.Wrap(repositories.ErrNotFound, "failed to find user by id")
	}

	if err != nil {
		return nil, faults.Wrap(err, "unknown error - failed to find user by id")
	}

	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password, role
        FROM users
        WHERE email = $1
    `

	var u models.User

	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Password,
		&u.Role,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, faults.Wrap(repositories.ErrNotFound, "failed to find user by email")
	}

	if err != nil {
		return nil, faults.Wrap(err, "unknown error - failed to find user by email")
	}

	return &u, nil
}

func (r *UserRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return faults.Wrap(err, "failed to delete user")
	}

	if cmdTag.RowsAffected() == 0 {
		return faults.Wrap(repositories.ErrNotFound, "failed to delete user")
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
	query := `UPDATE users SET email = $1, password = $2, role = $3 WHERE id = $4`

	cmdTag, err := r.db.Exec(ctx, query, u.Email, u.Password, u.Role, u.ID)
	if err != nil {
		return faults.Wrap(err, "failed to update user")
	}

	if cmdTag.RowsAffected() == 0 {
		return faults.Wrap(repositories.ErrNotFound, "failed to update user")
	}

	return nil
}
