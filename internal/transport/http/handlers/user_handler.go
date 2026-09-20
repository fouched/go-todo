package handlers

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/repositories"
	"github.com/fouched/go-todo/internal/core/services"
	"github.com/fouched/go-todo/platform/security"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	service   *services.UserService
	jwtSecret string
	logger    *slog.Logger
}

func NewUserHandler(service *services.UserService, jwtSecret string, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		service:   service,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

func (h *UserHandler) RegisterPublicRoutes(app *fiber.App) {
	group := app.Group("/api/users")
	group.Post("/register", h.Register)
	group.Post("/login", h.Login)
}

func (h *UserHandler) RegisterProtectedRoutes(app *fiber.App) {
	group := app.Group("/api/users")
	group.Get("/:id", h.GetUserByID)
	group.Delete("/:id", security.RequireRole(models.RoleAdmin), h.DeleteUser)
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	h.logger.Debug("In Register")
	var req AuthRequest

	if err := c.Bind().Body(&req); err != nil {
		h.logger.Warn("registration failed: invalid JSON payload", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON payload",
		})
	}

	user, err := h.service.RegisterUser(c.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		h.logger.Error("registration failed: unknown error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

func (h *UserHandler) Login(c fiber.Ctx) error {
	var req AuthRequest

	if err := c.Bind().Body(&req); err != nil {
		h.logger.Warn("registration failed: invalid JSON payload", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON payload",
		})
	}

	user, err := h.service.LoginUser(c.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("registration failed: invalid credentials", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	token, err := security.GenerateToken(user, h.jwtSecret)
	if err != nil {
		h.logger.Error("registration failed: error generating token", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
		"token": token,
	})
}

func (h *UserHandler) Logout(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "logged out"})
}

func (h *UserHandler) GetUserByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		h.logger.Error("GetUserByID: invalid id", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	user, err := h.service.GetUserByID(c.Context(), id)
	if err != nil {
		h.logger.Error("GetUserByID: user not found", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		h.logger.Error("GetUserByID: invalid id", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	err = h.service.DeleteUser(c.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			h.logger.Error("DeleteUser: user not found", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		h.logger.Error("DeleteUser: unknown error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "user deleted",
	})
}
