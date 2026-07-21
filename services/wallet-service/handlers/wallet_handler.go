package handlers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"banking-app/services/wallet-service/models"
	"banking-app/shared/utils"
)

type WalletHandler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewWalletHandler(db *gorm.DB) *WalletHandler {
	return &WalletHandler{
		DB:       db,
		Validate: validator.New(),
	}
}

func (h *WalletHandler) CreateWallet(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}

	// Check if wallet already exists
	var existing models.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&existing); result.Error == nil {
		return utils.ErrorResponse(c, fiber.StatusConflict, "Wallet already exists", nil)
	}

	wallet := models.Wallet{
		UserID:        userUUID,
		AccountNumber: generateAccountNumber(),
		Currency:      "NGN",
		Status:        "active",
		DailyLimit:    500000,
	}

	if result := h.DB.Create(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create wallet", result.Error.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Wallet created successfully", wallet)
}

func (h *WalletHandler) GetWallet(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var wallet models.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Wallet retrieved successfully", wallet)
}

func (h *WalletHandler) FundWallet(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.FundWalletRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var wallet models.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}

	if wallet.Status != "active" {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Wallet is not active", nil)
	}

	// Use transaction to ensure atomicity
	tx := h.DB.Begin()

	balanceBefore := wallet.Balance
	wallet.Balance += req.Amount
	tx.Save(&wallet)

	transaction := models.WalletTransaction{
		WalletID:      wallet.ID,
		Type:          "credit",
		Amount:        req.Amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  wallet.Balance,
		Description:   req.Description,
		Reference:     generateReference(),
		Status:        "success",
	}
	tx.Create(&transaction)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Transaction failed", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Wallet funded successfully", fiber.Map{
		"wallet":      wallet,
		"transaction": transaction,
	})
}

func (h *WalletHandler) Transfer(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.TransferRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var senderWallet models.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&senderWallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Sender wallet not found", nil)
	}

	if senderWallet.Balance < req.Amount {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Insufficient balance", nil)
	}

	var receiverWallet models.Wallet
	if result := h.DB.Where("account_number = ?", req.ToAccountNumber).First(&receiverWallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Receiver account not found", nil)
	}

	if senderWallet.ID == receiverWallet.ID {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Cannot transfer to same account", nil)
	}

	tx := h.DB.Begin()
	reference := generateReference()

	// Debit sender
	senderBalanceBefore := senderWallet.Balance
	senderWallet.Balance -= req.Amount
	tx.Save(&senderWallet)

	tx.Create(&models.WalletTransaction{
		WalletID:      senderWallet.ID,
		Type:          "debit",
		Amount:        req.Amount,
		BalanceBefore: senderBalanceBefore,
		BalanceAfter:  senderWallet.Balance,
		Description:   fmt.Sprintf("Transfer to %s: %s", receiverWallet.AccountNumber, req.Description),
		Reference:     reference,
	})

	// Credit receiver
	receiverBalanceBefore := receiverWallet.Balance
	receiverWallet.Balance += req.Amount
	tx.Save(&receiverWallet)

	tx.Create(&models.WalletTransaction{
		WalletID:      receiverWallet.ID,
		Type:          "credit",
		Amount:        req.Amount,
		BalanceBefore: receiverBalanceBefore,
		BalanceAfter:  receiverWallet.Balance,
		Description:   fmt.Sprintf("Transfer from %s: %s", senderWallet.AccountNumber, req.Description),
		Reference:     reference + "_IN",
	})

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Transfer failed", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Transfer successful", fiber.Map{
		"amount":      req.Amount,
		"reference":   reference,
		"new_balance": senderWallet.Balance,
	})
}

func (h *WalletHandler) Withdraw(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.WithdrawRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var wallet models.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}

	if wallet.Status != "active" {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Wallet is not active", nil)
	}

	if wallet.Balance < req.Amount {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Insufficient balance", nil)
	}

	tx := h.DB.Begin()
	balanceBefore := wallet.Balance
	wallet.Balance -= req.Amount
	tx.Save(&wallet)

	transaction := models.WalletTransaction{
		WalletID:      wallet.ID,
		Type:          "debit",
		Amount:        req.Amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  wallet.Balance,
		Description:   fmt.Sprintf("Withdrawal: %s", req.Description),
		Reference:     generateReference(),
		Status:        "success",
	}
	tx.Create(&transaction)

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Withdrawal failed", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Withdrawal successful", fiber.Map{
		"wallet":      wallet,
		"transaction": transaction,
	})
}

func (h *WalletHandler) GetTransactions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var wallet models.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}

	var transactions []models.WalletTransaction
	h.DB.Where("wallet_id = ?", wallet.ID).
		Order("created_at DESC").
		Find(&transactions)

	return utils.SuccessResponse(c, fiber.StatusOK, "Transactions retrieved", transactions)
}

func (h *WalletHandler) GetBalance(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var wallet models.Wallet
	if result := h.DB.Where("user_id = ?", userID).First(&wallet); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Balance retrieved", fiber.Map{
		"balance":  wallet.Balance,
		"currency": wallet.Currency,
	})
}

func generateAccountNumber() string {
	return fmt.Sprintf("%010d", rand.Intn(9000000000)+1000000000)
}

func generateReference() string {
	return fmt.Sprintf("TXN%d%s", time.Now().Unix(), uuid.New().String()[:8])
}
