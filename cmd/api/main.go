package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/fouched/go-todo/internal/core/services"
	"github.com/fouched/go-todo/internal/transport/http/handlers"
	"github.com/fouched/go-todo/platform/config"
	"github.com/fouched/go-todo/platform/postgres"
	"github.com/fouched/go-todo/platform/security"
	"github.com/fouched/toolkit/v2/logging"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(logging.NewPrettyDevHandler())

	// Connect to Postgres
	db, err := postgres.NewDB(ctx,
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.Pool, logger)
	taskRepo := postgres.NewTaskRepository(db.Pool)

	// Initialize services
	userService := services.NewUserService(userRepo, logger)
	taskService := services.NewTaskService(taskRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, cfg.JWT.Secret, logger)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Fiber v3 app
	app := fiber.New()

	app.Use(cors.New())

	// Public routes
	userHandler.RegisterPublicRoutes(app)

	// Protected routes
	app.Use(security.JWTMiddleware(cfg.JWT.Secret))
	userHandler.RegisterProtectedRoutes(app)
	taskHandler.RegisterRoutes(app)

	// Start server
	logger.Info("Starting server on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
