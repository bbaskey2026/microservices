package routes

import (
	"banking-app/services/wallet-service/handlers"
	"banking-app/shared/health"
	"banking-app/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, handler *handlers.WalletHandler) {
	app.Get("/api/v1/wallets/health", health.LivenessHandler("Wallet Service"))

	wallets := app.Group("/api/v1/wallets", middleware.AuthMiddleware())

	wallets.Post("/", handler.CreateWallet)
	wallets.Get("/", handler.GetWallet)
	wallets.Get("/balance", handler.GetBalance)
	wallets.Post("/fund", handler.FundWallet)
	wallets.Post("/transfer", handler.Transfer)
	wallets.Post("/withdraw", handler.Withdraw)
	wallets.Get("/transactions", handler.GetTransactions)
}
