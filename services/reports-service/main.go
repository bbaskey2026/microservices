package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"banking-app/services/reports-service/handlers"
	"banking-app/services/reports-service/routes"
	"banking-app/shared/database"
	"banking-app/shared/health"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := database.ConnectPostgres("REPORTS")

	app := fiber.New(fiber.Config{
		AppName: "Banking Reports Service v1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH",
	}))

	health.RegisterHealthRoutes(app, "Reports Service", db, nil)

	handler := handlers.NewReportHandler(db)
	routes.SetupRoutes(app, handler)

	port := os.Getenv("REPORTS_SERVICE_PORT")
	if port == "" {
		port = "3007"
	}

	log.Printf("🚀 Reports Service running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
