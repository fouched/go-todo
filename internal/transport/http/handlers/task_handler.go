package handlers

import (
	"strconv"

	"github.com/fouched/go-todo/internal/core/models"
	"github.com/fouched/go-todo/internal/core/services"
	"github.com/gofiber/fiber/v3"
)

type TaskHandler struct {
	service *services.TaskService
}

func NewTaskHandler(service *services.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/tasks")

	group.Post("/", h.CreateTask)
	group.Get("/:id", h.GetTaskByID)
	group.Get("/user/:userID", h.GetTasksForUser)
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

	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id is required",
		})
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
	}

	saved, err := h.service.CreateTask(c.Context(), userID, task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(TaskResponse{
		ID:          saved.ID,
		Title:       saved.Title,
		Description: saved.Description,
		Category:    saved.Category,
		IsCompleted: saved.IsCompleted,
	})
}

func (h *TaskHandler) GetTaskByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	task, err := h.service.GetTaskByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	return c.JSON(TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Category:    task.Category,
		IsCompleted: task.IsCompleted,
	})
}

func (h *TaskHandler) GetTasksForUser(c fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid userID",
		})
	}

	tasks, err := h.service.GetTasksForUser(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
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
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var req UpdateTaskRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON payload",
		})
	}

	task := &models.Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		IsCompleted: req.IsCompleted,
	}

	if err := h.service.UpdateTask(c.Context(), task); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
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
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	if err := h.service.DeleteTask(c.Context(), id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "task not found",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
