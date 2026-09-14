package postgres

import (
	"context"
	"errors"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repositories.UserRepository {
	return &UserRepository{db: db}
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
				return repositories.ErrDuplicateEmail
			}
		}
	}

	return err
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
		return nil, repositories.ErrNotFound
	}

	return &u, err
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
		return nil, repositories.ErrNotFound
	}

	return &u, err
}

func (r *UserRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return repositories.ErrNotFound
	}

	return nil
}
