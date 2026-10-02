package handlers

import (
	"log/slog"
	"strconv"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/platform/security"
	"github.com/gofiber/fiber/v3"
)

type TaskHandler struct {
	service TaskService
	logger  *slog.Logger
}

func NewTaskHandler(service TaskService, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		service: service,
		logger:  logger,
	}
}

func (h *TaskHandler) RegisterRoutes(app *fiber.App, jwtSecret string) {
	// 1. Establish the clean, protected routing tree root
	group := app.Group("/api/tasks", security.JWTMiddleware(jwtSecret))

	// 2. Base Resource Paths (Using explicit empty strings for Fiber v3 routing consistency)
	group.Post("", h.CreateTask)
	group.Get("", h.GetTasksByUserAndCategory)

	// 3. Specific Sub-action modifiers (Placed logically before general wildcard actions if expanded later)
	group.Put("/:id/completed", h.ToggleTaskCompletion)

	// 4. Wildcard Parameter Resource Paths
	group.Put("/:id", h.UpdateTask)
	group.Delete("/:id", h.DeleteTask)
}

func (h *TaskHandler) CreateTask(c fiber.Ctx) error {
	var req CreateTaskRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON payload",
		})
	}

	claims, ok := security.GetClaims(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
	}

	saved, err := h.service.CreateTask(c.Context(), claims.UserID, task)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(TaskResponse{
		ID:          saved.ID,
		Title:       saved.Title,
		Description: saved.Description,
		Category:    saved.Category,
		IsCompleted: saved.IsCompleted,
	})
}

func (h *TaskHandler) GetTasksByUserAndCategory(c fiber.Ctx) error {
	claims, ok := security.GetClaims(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Read from query parameters
	category := c.Query("category", "")

	tasks, err := h.service.GetTasksByUserAndCategory(c.Context(), claims.UserID, category)
	if err != nil {
		return err
	}

	responses := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		responses = append(responses, TaskResponse{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Category:    t.Category,
			IsCompleted: t.IsCompleted,
		})
	}

	return c.JSON(responses)
}

func (h *TaskHandler) UpdateTask(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid task id",
		})
	}

	claims, ok := security.GetClaims(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req UpdateTaskRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON payload",
		})
	}

	task := &models.Task{
		ID:          id,
		UserID:      claims.UserID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		IsCompleted: req.IsCompleted,
	}

	if err := h.service.UpdateTask(c.Context(), task); err != nil {
		return err
	}

	return c.JSON(TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Category:    task.Category,
		IsCompleted: task.IsCompleted,
	})
}

func (h *TaskHandler) DeleteTask(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid task id",
		})
	}

	claims, ok := security.GetClaims(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.service.DeleteTask(c.Context(), claims.UserID, id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TaskHandler) ToggleTaskCompletion(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid task id",
		})
	}

	claims, ok := security.GetClaims(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Read target state from query params: ?status=true (defaults to true if omitted)
	isCompleted := c.Query("status", "true") == "true"

	// Invoke service layer contract
	updatedTask, err := h.service.ToggleTaskCompletion(c.Context(), claims.UserID, id, isCompleted)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(TaskResponse{
		ID:          updatedTask.ID,
		Title:       updatedTask.Title,
		Description: updatedTask.Description,
		Category:    updatedTask.Category,
		IsCompleted: updatedTask.IsCompleted,
	})
}
