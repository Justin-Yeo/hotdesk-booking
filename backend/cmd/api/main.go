package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"

	"github.com/justinyeo/hotdesk-booking/backend/internal/config"
	"github.com/justinyeo/hotdesk-booking/backend/internal/database"
	"github.com/justinyeo/hotdesk-booking/backend/internal/handlers"
	customMiddleware "github.com/justinyeo/hotdesk-booking/backend/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize structured logger
	logger, err := initLogger(cfg.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting hotdesk booking API server",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.ServerPort),
	)

	// Initialize database connection pool
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		dbConfig := database.DefaultConfig(cfg.DatabaseURL)
		_, err := database.Connect(ctx, dbConfig)
		if err != nil {
			logger.Warn("Failed to connect to database", zap.Error(err))
		} else {
			logger.Info("Database connection pool initialized",
				zap.Int32("max_connections", dbConfig.MaxConnections),
				zap.Int32("min_connections", dbConfig.MinConnections),
			)
			defer database.Close()
		}
	} else {
		logger.Warn("DATABASE_URL not set, skipping database connection")
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Hotdesk Booking API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			logger.Error("Request error",
				zap.Error(err),
				zap.Int("status", code),
				zap.String("path", c.Path()),
			)

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(customMiddleware.Logger(logger))

	// Routes
	api := app.Group("/api")
	api.Get("/health", handlers.HealthCheck)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down server...")
		app.Shutdown()
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	logger.Info("Server listening", zap.String("address", addr))

	if err := app.Listen(addr); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func initLogger(environment string) (*zap.Logger, error) {
	if environment == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
