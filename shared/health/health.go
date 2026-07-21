package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var startTime = time.Now()

type HealthResponse struct {
	Success      bool                         `json:"success"`
	Status       string                       `json:"status"`
	Service      string                       `json:"service"`
	Timestamp    time.Time                    `json:"timestamp"`
	Uptime       string                       `json:"uptime,omitempty"`
	Dependencies map[string]DependencyDetails `json:"dependencies,omitempty"`
}

type DependencyDetails struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// LivenessHandler returns basic HTTP 200 liveness check
func LivenessHandler(serviceName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(HealthResponse{
			Success:   true,
			Status:    "UP",
			Service:   serviceName,
			Timestamp: time.Now().UTC(),
			Uptime:    time.Since(startTime).String(),
		})
	}
}

// ReadinessHandler checks database and redis connectivity
func ReadinessHandler(serviceName string, db *gorm.DB, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dependencies := make(map[string]DependencyDetails)
		overallStatus := "UP"
		httpCode := fiber.StatusOK

		if db != nil {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				dependencies["database"] = DependencyDetails{
					Status:  "DOWN",
					Message: "PostgreSQL database connection failed",
				}
				overallStatus = "DOWN"
				httpCode = fiber.StatusServiceUnavailable
			} else {
				dependencies["database"] = DependencyDetails{
					Status:  "UP",
					Message: "Connected to PostgreSQL",
				}
			}
		}

		if redisClient != nil {
			ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
			defer cancel()

			if err := redisClient.Ping(ctx).Err(); err != nil {
				dependencies["redis"] = DependencyDetails{
					Status:  "DOWN",
					Message: "Redis connection failed",
				}
				overallStatus = "DOWN"
				httpCode = fiber.StatusServiceUnavailable
			} else {
				dependencies["redis"] = DependencyDetails{
					Status:  "UP",
					Message: "Connected to Redis",
				}
			}
		}

		return c.Status(httpCode).JSON(HealthResponse{
			Success:      overallStatus == "UP",
			Status:       overallStatus,
			Service:      serviceName,
			Timestamp:    time.Now().UTC(),
			Uptime:       time.Since(startTime).String(),
			Dependencies: dependencies,
		})
	}
}

// RegisterHealthRoutes helper to bind /health, /healthz and /ready
func RegisterHealthRoutes(app *fiber.App, serviceName string, db *gorm.DB, redisClient *redis.Client) {
	app.Get("/health", LivenessHandler(serviceName))
	app.Get("/healthz", LivenessHandler(serviceName))
	app.Get("/ready", ReadinessHandler(serviceName, db, redisClient))
}
