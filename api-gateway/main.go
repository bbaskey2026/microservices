package main

import (
	"log"
	"os"
	"time"

	"banking-app/shared/health"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

type ServiceConfig struct {
	Name string
	URL  string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	services := map[string]string{
		"auth":         os.Getenv("AUTH_SERVICE_URL"),
		"user":         os.Getenv("USER_SERVICE_URL"),
		"card":         os.Getenv("CARD_SERVICE_URL"),
		"wallet":       os.Getenv("WALLET_SERVICE_URL"),
		"payment":      os.Getenv("PAYMENT_SERVICE_URL"),
		"notification": os.Getenv("NOTIFICATION_SERVICE_URL"),
		"reports":      os.Getenv("REPORTS_SERVICE_URL"),
	}

	app := fiber.New(fiber.Config{
		AppName:      "Banking API Gateway v1.0",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Too many requests, please try again later",
			})
		},
	}))

	// Gateway liveness
	app.Get("/health", health.LivenessHandler("API Gateway"))
	app.Get("/healthz", health.LivenessHandler("API Gateway"))

	// Aggregate health of downstream services
	app.Get("/health/services", func(c *fiber.Ctx) error {
		type serviceStatus struct {
			Name   string `json:"name"`
			URL    string `json:"url"`
			Status string `json:"status"`
		}

		results := make([]serviceStatus, 0, len(services))
		overallStatus := "UP"

		for name, url := range services {
			status := "UP"
			if url == "" {
				status = "UNCONFIGURED"
				overallStatus = "DEGRADED"
			}
			results = append(results, serviceStatus{
				Name:   name,
				URL:    url,
				Status: status,
			})
		}

		httpCode := fiber.StatusOK
		if overallStatus != "UP" {
			httpCode = fiber.StatusServiceUnavailable
		}

		return c.Status(httpCode).JSON(fiber.Map{
			"success":  overallStatus == "UP",
			"status":   overallStatus,
			"gateway":  "API Gateway",
			"services": results,
		})
	})

	// Route proxying
	setupProxyRoutes(app, services)

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 API Gateway running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupProxyRoutes(app *fiber.App, services map[string]string) {
	// Auth routes
	app.All("/api/v1/auth/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, services["auth"]+c.OriginalURL())
	})

	// User routes
	app.All("/api/v1/users/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, services["user"]+c.OriginalURL())
	})

	// Card routes
	app.All("/api/v1/cards/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, services["card"]+c.OriginalURL())
	})

	// Wallet routes
	app.All("/api/v1/wallets/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, services["wallet"]+c.OriginalURL())
	})

	// Payment routes
	app.All("/api/v1/payments/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, services["payment"]+c.OriginalURL())
	})

	// Notification routes
	app.All("/api/v1/notifications/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, services["notification"]+c.OriginalURL())
	})

	// Reports routes
	app.All("/api/v1/reports/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, services["reports"]+c.OriginalURL())
	})
}
