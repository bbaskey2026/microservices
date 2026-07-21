package routes

import (
	"banking-app/services/card-service/handlers"
	"banking-app/shared/health"
	"banking-app/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.CardHandler) {
	app.Get("/api/v1/cards/health", health.LivenessHandler("Card Service"))

	cards := app.Group("/api/v1/cards", middleware.AuthMiddleware())

	cards.Post("/", handler.CreateCard)
	cards.Get("/", handler.GetUserCards)
	cards.Get("/:id", handler.GetCardByID)
	cards.Put("/:id", handler.UpdateCard)
	cards.Put("/:id/pin", handler.ChangePIN)
	cards.Put("/:id/block", handler.BlockCard)
	cards.Put("/:id/unblock", handler.UnblockCard)
	cards.Delete("/:id", handler.DeleteCard)
}
