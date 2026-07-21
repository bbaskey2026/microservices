package routes

import (
	"banking-app/services/reports-service/handlers"
	"banking-app/shared/health"
	"banking-app/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.ReportHandler) {
	app.Get("/api/v1/reports/health", health.LivenessHandler("Reports Service"))

	reports := app.Group("/api/v1/reports", middleware.AuthMiddleware())

	reports.Get("/summary", handler.GetTransactionSummary)
	reports.Get("/monthly", handler.GetMonthlyReport)
	reports.Get("/payments", handler.GetPaymentSummary)
	reports.Get("/spending", handler.GetSpendingAnalytics)
	reports.Get("/admin/dashboard", handler.GetAdminDashboard)
}
