package middleware

import (
	"errors"
	"log/slog"

	"github.com/fouched/go-todo/internal/core/repositories"
	"github.com/fouched/toolkit/v2/faults"
	"github.com/gofiber/fiber/v3"
)

// NewErrorHandler configures Fiber's global error processing block.
func NewErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		// 1. Handle explicit Fiber framework HTTP errors (e.g., 404 Route Not Found, 405 Method Not Allowed)
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(fiber.Map{
				"error": fiberErr.Message,
			})
		}

		// 2. Handle your application's Core Domain Sentinels
		if faults.Is(err, repositories.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "the requested resource could not be found",
			})
		}

		if faults.Is(err, repositories.ErrDuplicateEmail) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "a user with this email address already exists",
			})
		}

		// 3. Handle Unexpected Server System Faults (500 Internal Server Errors)
		// Capture standard request metadata for context tracking
		logger.ErrorContext(c.Context(), "unhandled application error caught by middleware",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Any("error", err),
		)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "an unexpected internal error occurred",
		})
	}
}
