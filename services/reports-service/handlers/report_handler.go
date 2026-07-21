package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	paymentModels "banking-app/services/payment-service/models"
	walletModels "banking-app/services/wallet-service/models"
	"banking-app/shared/utils"
)

type ReportHandler struct {
	DB *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{DB: db}
}

type TransactionSummary struct {
	TotalCredit       float64 `json:"total_credit"`
	TotalDebit        float64 `json:"total_debit"`
	TotalTransactions int64   `json:"total_transactions"`
	NetBalance        float64 `json:"net_balance"`
}

type MonthlyReport struct {
	Month        string                           `json:"month"`
	Summary      TransactionSummary               `json:"summary"`
	Transactions []walletModels.WalletTransaction `json:"transactions"`
}

func (h *ReportHandler) GetTransactionSummary(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	startDate := c.Query("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.Query("end_date", time.Now().Format("2006-01-02"))

	var wallet walletModels.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}

	var summary struct {
		TotalCredit float64
		TotalDebit  float64
		Count       int64
	}

	h.DB.Model(&walletModels.WalletTransaction{}).
		Where("wallet_id = ? AND created_at BETWEEN ? AND ?", wallet.ID, startDate, endDate+" 23:59:59").
		Select("SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END) as total_credit, SUM(CASE WHEN type = 'debit' THEN amount ELSE 0 END) as total_debit, COUNT(*) as count").
		Scan(&summary)

	return utils.SuccessResponse(c, fiber.StatusOK, "Summary retrieved", fiber.Map{
		"period": fiber.Map{
			"start": startDate,
			"end":   endDate,
		},
		"total_credit":       summary.TotalCredit,
		"total_debit":        summary.TotalDebit,
		"total_transactions": summary.Count,
		"net":                summary.TotalCredit - summary.TotalDebit,
	})
}

func (h *ReportHandler) GetMonthlyReport(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	year := c.Query("year", fmt.Sprintf("%d", time.Now().Year()))
	monthQuery := c.Query("month", fmt.Sprintf("%d", int(time.Now().Month())))

	monthInt, _ := strconv.Atoi(monthQuery)
	if monthInt < 1 || monthInt > 12 {
		monthInt = int(time.Now().Month())
	}

	startDate := fmt.Sprintf("%s-%02d-01", year, monthInt)
	endDate := fmt.Sprintf("%s-%02d-31", year, monthInt)

	var wallet walletModels.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}

	var transactions []walletModels.WalletTransaction
	h.DB.Where("wallet_id = ? AND created_at BETWEEN ? AND ?", wallet.ID, startDate, endDate+" 23:59:59").
		Order("created_at DESC").
		Find(&transactions)

	var totalCredit, totalDebit float64
	for _, t := range transactions {
		if t.Type == "credit" {
			totalCredit += t.Amount
		} else {
			totalDebit += t.Amount
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Monthly report retrieved", fiber.Map{
		"period":       fmt.Sprintf("%02d/%s", monthInt, year),
		"total_credit": totalCredit,
		"total_debit":  totalDebit,
		"net":          totalCredit - totalDebit,
		"transactions": transactions,
	})
}

func (h *ReportHandler) GetPaymentSummary(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	startDate := c.Query("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDate := c.Query("end_date", time.Now().Format("2006-01-02"))

	type CategorySummary struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
		Count    int64   `json:"count"`
	}

	var categorySummaries []CategorySummary
	h.DB.Model(&paymentModels.Payment{}).
		Where("user_id = ? AND status = ? AND created_at BETWEEN ? AND ?",
			userID, "success", startDate, endDate+" 23:59:59").
		Select("category, SUM(amount) as total, COUNT(*) as count").
		Group("category").
		Scan(&categorySummaries)

	return utils.SuccessResponse(c, fiber.StatusOK, "Payment summary retrieved", fiber.Map{
		"period":     fiber.Map{"start": startDate, "end": endDate},
		"categories": categorySummaries,
	})
}

func (h *ReportHandler) GetSpendingAnalytics(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	type MonthlySpending struct {
		Month  string  `json:"month"`
		Amount float64 `json:"amount"`
	}

	var monthlySpending []MonthlySpending
	h.DB.Model(&walletModels.WalletTransaction{}).
		Joins("JOIN wallets ON wallets.id = wallet_transactions.wallet_id").
		Where("wallets.user_id = ? AND wallet_transactions.type = ?", userID, "debit").
		Select("TO_CHAR(wallet_transactions.created_at, 'YYYY-MM') as month, SUM(wallet_transactions.amount) as amount").
		Group("month").
		Order("month DESC").
		Limit(12).
		Scan(&monthlySpending)

	return utils.SuccessResponse(c, fiber.StatusOK, "Spending analytics retrieved", monthlySpending)
}

func (h *ReportHandler) GetAdminDashboard(c *fiber.Ctx) error {
	type DashboardStats struct {
		TotalUsers        int64   `json:"total_users"`
		ActiveUsers       int64   `json:"active_users"`
		TotalTransactions int64   `json:"total_transactions"`
		TotalVolume       float64 `json:"total_volume"`
		TotalPayments     int64   `json:"total_payments"`
	}

	var stats DashboardStats
	h.DB.Raw("SELECT COUNT(*) as total_users FROM users WHERE deleted_at IS NULL").Scan(&stats.TotalUsers)
	h.DB.Raw("SELECT COUNT(*) as active_users FROM users WHERE status = 'active' AND deleted_at IS NULL").Scan(&stats.ActiveUsers)
	h.DB.Raw("SELECT COUNT(*) as total_transactions FROM wallet_transactions").Scan(&stats.TotalTransactions)
	h.DB.Raw("SELECT COALESCE(SUM(amount), 0) as total_volume FROM wallet_transactions WHERE type = 'credit'").Scan(&stats.TotalVolume)
	h.DB.Raw("SELECT COUNT(*) as total_payments FROM payments WHERE status = 'success'").Scan(&stats.TotalPayments)

	return utils.SuccessResponse(c, fiber.StatusOK, "Dashboard stats retrieved", stats)
}
