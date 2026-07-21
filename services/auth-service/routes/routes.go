package routes

import (
	"banking-app/services/auth-service/handlers"
	"banking-app/shared/health"
	"banking-app/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.AuthHandler) {
	auth := app.Group("/api/v1/auth")

	// Public routes
	auth.Get("/health", health.LivenessHandler("Auth Service"))
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Post("/refresh-token", handler.RefreshToken)

	// Protected routes
	auth.Post("/logout", middleware.AuthMiddleware(), handler.Logout)
	auth.Put("/change-password", middleware.AuthMiddleware(), handler.ChangePassword)
}

