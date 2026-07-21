package routes

import (
	"banking-app/services/payment-service/handlers"
	"banking-app/shared/health"
	"banking-app/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.PaymentHandler) {
	app.Get("/api/v1/payments/health", health.LivenessHandler("Payment Service"))

	payments := app.Group("/api/v1/payments", middleware.AuthMiddleware())

	payments.Post("/bill", handler.PayBill)
	payments.Post("/airtime", handler.BuyAirtime)
	payments.Post("/data", handler.BuyDataBundle)
	payments.Get("/history", handler.GetPaymentHistory)
	payments.Get("/:ref", handler.GetPaymentByRef)
}
