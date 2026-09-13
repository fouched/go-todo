package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/fouched/go-todo/internal/core/services"
	"github.com/fouched/go-todo/internal/transport/http/handlers"
	"github.com/fouched/go-todo/platform/postgres"
	"github.com/gofiber/fiber/v3"
)

func main() {
	ctx := context.Background()

	// Load config from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnvInt("DB_PORT", 5432)
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "todo")

	// Connect to Postgres
	db, err := postgres.NewDB(ctx, host, user, password, dbname, port)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.Pool)
	taskRepo := postgres.NewTaskRepository(db.Pool)

	// Initialize services
	userService := services.NewUserService(userRepo)
	taskService := services.NewTaskService(taskRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Fiber v3 app
	app := fiber.New()

	// Register routes
	userHandler.RegisterRoutes(app)
	taskHandler.RegisterRoutes(app)

	// Start server
	log.Println("Starting server on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}
