package handlers

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"banking-app/services/payment-service/models"
	"banking-app/shared/utils"
)

type PaymentHandler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewPaymentHandler(db *gorm.DB) *PaymentHandler {
	return &PaymentHandler{
		DB:       db,
		Validate: validator.New(),
	}
}

func (h *PaymentHandler) PayBill(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.BillPaymentRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	userUUID, _ := uuid.Parse(userID)
	now := time.Now()
	payment := models.Payment{
		UserID:      userUUID,
		Reference:   generatePaymentRef(),
		Amount:      req.Amount,
		Type:        "bill_payment",
		Category:    req.Category,
		Provider:    req.Provider,
		AccountRef:  req.AccountRef,
		Description: req.Description,
		Status:      "success",
		ProcessedAt: &now,
	}

	if result := h.DB.Create(&payment); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Payment failed", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Bill payment successful", payment)
}

func (h *PaymentHandler) BuyAirtime(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.AirtimeRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	userUUID, _ := uuid.Parse(userID)
	now := time.Now()
	payment := models.Payment{
		UserID:      userUUID,
		Reference:   generatePaymentRef(),
		Amount:      req.Amount,
		Type:        "airtime",
		Category:    "airtime",
		Provider:    req.Network,
		AccountRef:  req.PhoneNumber,
		Description: fmt.Sprintf("%s Airtime for %s", req.Network, req.PhoneNumber),
		Status:      "success",
		ProcessedAt: &now,
	}

	if result := h.DB.Create(&payment); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Airtime purchase failed", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Airtime purchased successfully", payment)
}

func (h *PaymentHandler) BuyDataBundle(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.DataBundleRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	userUUID, _ := uuid.Parse(userID)
	now := time.Now()
	payment := models.Payment{
		UserID:      userUUID,
		Reference:   generatePaymentRef(),
		Amount:      req.Amount,
		Type:        "data",
		Category:    "data",
		Provider:    req.Network,
		AccountRef:  req.PhoneNumber,
		Description: fmt.Sprintf("%s Data Bundle (%s) for %s", req.Network, req.BundleCode, req.PhoneNumber),
		Status:      "success",
		ProcessedAt: &now,
	}

	if result := h.DB.Create(&payment); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Data purchase failed", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Data bundle purchased successfully", payment)
}

func (h *PaymentHandler) GetPaymentHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	paymentType := c.Query("type", "")
	status := c.Query("status", "")

	query := h.DB.Where("user_id = ?", userID)
	if paymentType != "" {
		query = query.Where("type = ?", paymentType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var payments []models.Payment
	query.Order("created_at DESC").Find(&payments)

	return utils.SuccessResponse(c, fiber.StatusOK, "Payment history retrieved", payments)
}

func (h *PaymentHandler) GetPaymentByRef(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	ref := c.Params("ref")

	var payment models.Payment
	if result := h.DB.Where("reference = ? AND user_id = ?", ref, userID).First(&payment); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Payment not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Payment retrieved", payment)
}

func generatePaymentRef() string {
	return fmt.Sprintf("PAY%d%s", time.Now().Unix(), uuid.New().String()[:6])
}
