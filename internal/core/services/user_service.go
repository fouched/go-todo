package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/repositories"
	"github.com/fouched/toolkit/v2/faults"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users  repositories.UserRepository
	logger *slog.Logger
}

func NewUserService(users repositories.UserRepository, logger *slog.Logger) *UserService {
	return &UserService{
		users:  users,
		logger: logger,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, email, password string, role models.Role) (*models.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    email,
		Password: string(hashed),
		Role:     role,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, faults.Wrap(err, "failed to create user")
	}

	return user, nil
}

func (s *UserService) LoginUser(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, faults.Wrap(err, "failed to find user by email")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.users.DeleteByID(ctx, id)
}
