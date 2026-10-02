package handlers

import (
	"log/slog"
	"strconv"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/platform/security"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	service UserService
	logger  *slog.Logger
}

func NewUserHandler(service UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) RegisterRoutes(app *fiber.App, jwtSecret string) {
	// 1. Define the base path exactly once
	baseGroup := app.Group("/api/users")

	// 2. Register completely public endpoints
	baseGroup.Post("/register", h.Register)
	baseGroup.Post("/login", func(c fiber.Ctx) error {
		return h.Login(c, jwtSecret)
	})

	// 3. Admin-only endpoints group
	// Register static sub-paths directly to isolate the role check
	baseGroup.Get("/all", security.JWTMiddleware(jwtSecret), security.RequireRole(models.RoleAdmin), h.GetAllUsers)

	// 4. Standard User/Authenticated endpoints group
	// Place specific static routes BEFORE parameter wildcards
	baseGroup.Post("/logout", security.JWTMiddleware(jwtSecret), h.Logout)

	// Wildcard parameter paths go last
	baseGroup.Get("/:id", security.JWTMiddleware(jwtSecret), h.GetUserByID)
	baseGroup.Delete("/:id", security.JWTMiddleware(jwtSecret), security.RequireRole(models.RoleAdmin), h.DeleteUser)
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	var req AuthRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON payload"})
	}

	user, err := h.service.RegisterUser(c.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		// Central error middleware intercepts ErrDuplicateEmail automatically and issues a 409!
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

func (h *UserHandler) Login(c fiber.Ctx, jwtSecret string) error {
	var req AuthRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON payload"})
	}

	user, err := h.service.LoginUser(c.Context(), req.Email, req.Password)
	if err != nil {
		// Tip: map an internal authentication failure error to StatusUnauthorized (401) in your central middleware
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := security.GenerateToken(user, jwtSecret)
	if err != nil {
		return err
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
	// If you want to track which user logged out in your logging system:
	claims, ok := security.GetClaims(c)
	if ok {
		h.logger.Info("User logged out successfully", slog.Int64("user_id", claims.UserID))
	}

	// Instruct the client to wipe its auth storage state
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

func (h *UserHandler) GetAllUsers(c fiber.Ctx) error {
	users, err := h.service.GetAllUsers(c.Context())
	if err != nil {
		return err
	}

	var response []UserResponse
	for _, u := range users {
		response = append(response, UserResponse{
			ID:    u.ID,
			Email: u.Email,
			Role:  u.Role,
		})
	}

	return c.JSON(response)
}

func (h *UserHandler) GetUserByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// Implicit Ownership Check: Ensure users can only look up their own profile details
	claims, ok := security.GetClaims(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if claims.UserID != id {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	user, err := h.service.GetUserByID(c.Context(), id)
	if err != nil {
		return err // Automatically yields 404 via ErrNotFound mapping
	}

	return c.JSON(UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	// This path is already locked to Admin by the routing group wrapper
	if err := h.service.DeleteUser(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}
