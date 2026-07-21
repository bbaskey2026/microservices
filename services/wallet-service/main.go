package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"banking-app/services/wallet-service/handlers"
	"banking-app/services/wallet-service/models"
	"banking-app/services/wallet-service/routes"
	"banking-app/shared/database"
	"banking-app/shared/health"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := database.ConnectPostgres("WALLET")

	// Auto migrate
	db.AutoMigrate(&models.Wallet{}, &models.WalletTransaction{})

	app := fiber.New(fiber.Config{
		AppName: "Banking Wallet Service v1.0",
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

	health.RegisterHealthRoutes(app, "Wallet Service", db, nil)

	handler := handlers.NewWalletHandler(db)
	routes.SetupRoutes(app, handler)

	port := os.Getenv("WALLET_SERVICE_PORT")
	if port == "" {
		port = "3004"
	}

	log.Printf("🚀 Wallet Service running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
