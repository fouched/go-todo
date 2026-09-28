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

func (h *UserHandler) RegisterPublicRoutes(app *fiber.App, jwtSecret string) {
	group := app.Group("/api/users")

	// Pass the secret via an anonymous wrapper so the Login method can use it
	group.Post("/register", h.Register)
	group.Post("/login", func(c fiber.Ctx) error {
		return h.Login(c, jwtSecret)
	})
}

func (h *UserHandler) RegisterProtectedRoutes(app *fiber.App, jwtSecret string) {
	// This subgroup is now explicitly protected
	group := app.Group("/api/users", security.JWTMiddleware(jwtSecret))

	group.Get("/:id", h.GetUserByID)
	group.Delete("/:id", security.RequireRole(models.RoleAdmin), h.DeleteUser)
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
	return c.JSON(fiber.Map{"message": "logged out"})
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

	if claims.UserID != id && claims.Role != models.RoleAdmin {
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
