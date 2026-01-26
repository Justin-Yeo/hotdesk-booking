package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func HealthCheck(c *fiber.Ctx) error {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Services: map[string]string{
			"database": "not_connected",
			"redis":    "not_connected",
		},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
