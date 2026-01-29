package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/justinyeo/hotdesk-booking/backend/internal/database"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func HealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check database connection
	dbStatus := "connected"
	if err := database.HealthCheck(ctx); err != nil {
		dbStatus = "not_connected"
	}

	// Determine overall health status
	overallStatus := "healthy"
	if dbStatus == "not_connected" {
		overallStatus = "degraded"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Services: map[string]string{
			"database": dbStatus,
			"redis":    "not_connected",
		},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
