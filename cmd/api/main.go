package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/fouched/go-todo/internal/core/services"
	"github.com/fouched/go-todo/internal/transport/http/handlers"
	"github.com/fouched/go-todo/internal/transport/http/middleware"
	"github.com/fouched/go-todo/platform/config"
	"github.com/fouched/go-todo/platform/postgres"
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

	baseLogger := initLogging(cfg)

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
	userRepo := postgres.NewUserRepository(db.Pool, baseLogger)
	taskRepo := postgres.NewTaskRepository(db.Pool, baseLogger)

	// Initialize services
	userService := services.NewUserService(userRepo, baseLogger)
	taskService := services.NewTaskService(taskRepo, baseLogger)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, baseLogger)
	taskHandler := handlers.NewTaskHandler(taskService, baseLogger)

	// Fiber v3 app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.NewErrorHandler(baseLogger),
	})
	app.Use(cors.New())

	// Public routes
	userHandler.RegisterPublicRoutes(app, cfg.JWT.Secret)

	// Protected paths - pass the secret explicitly so handlers can manage their own scopes
	userHandler.RegisterProtectedRoutes(app, cfg.JWT.Secret)
	taskHandler.RegisterProtectedRoutes(app, cfg.JWT.Secret)

	// Start server
	baseLogger.Info("Starting server on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func initLogging(cfg *config.Config) *slog.Logger {
	var encoder slog.Handler
	isDevelopment := cfg.App.Env == "development"
	if isDevelopment {
		encoder = logging.NewPrettyDevHandler()
	} else {
		lvl, err := logging.ParseLevel(cfg.Logging.Level)
		if err != nil {
			// Fallback to INFO but log the issue
			fmt.Printf("Invalid log level %q, defaulting to INFO\n", cfg.Logging.Level)
			lvl = slog.LevelInfo
		}

		logging.ProdLevel.Set(lvl)
		jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logging.ProdLevel, // <-- dynamic!
		})
		encoder = &logging.ProdHandler{
			Handler: jsonHandler,
		}
	}

	// Create the core base logger
	baseLogger := slog.New(encoder)

	// Optional: Set as global just in case external libraries rely on it
	slog.SetDefault(baseLogger)
	return baseLogger
}
