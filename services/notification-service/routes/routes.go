package routes

import (
	"banking-app/services/notification-service/handlers"
	"banking-app/shared/health"
	"banking-app/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.NotificationHandler) {
	app.Get("/api/v1/notifications/health", health.LivenessHandler("Notification Service"))

	notifs := app.Group("/api/v1/notifications", middleware.AuthMiddleware())

	notifs.Post("/", handler.SendNotification)
	notifs.Get("/", handler.GetUserNotifications)
	notifs.Put("/read-all", handler.MarkAllAsRead)
	notifs.Get("/unread-count", handler.GetUnreadCount)
	notifs.Get("/settings", handler.GetSettings)
	notifs.Put("/settings", handler.UpdateSettings)
	notifs.Put("/:id/read", handler.MarkAsRead)
	notifs.Delete("/:id", handler.DeleteNotification)
}
