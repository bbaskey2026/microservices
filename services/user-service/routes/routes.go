package routes

import (
	"banking-app/services/user-service/handlers"
	"banking-app/shared/health"
	"banking-app/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.UserHandler) {
	app.Get("/api/v1/users/health", health.LivenessHandler("User Service"))

	users := app.Group("/api/v1/users", middleware.AuthMiddleware())

	// User profile
	users.Get("/profile", handler.GetProfile)
	users.Put("/profile", handler.UpdateProfile)

	// Admin / User operations
	users.Get("/", handler.GetAllUsers)
	users.Get("/:id", handler.GetUserByID)
	users.Put("/:id/status", handler.UpdateUserStatus)
	users.Put("/:id/kyc", handler.UpdateKYCStatus)
	users.Delete("/:id", handler.DeleteUser)
}
