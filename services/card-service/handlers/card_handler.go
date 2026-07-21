package handlers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"banking-app/services/card-service/models"
	"banking-app/shared/utils"
)

type CardHandler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewCardHandler(db *gorm.DB) *CardHandler {
	return &CardHandler{
		DB:       db,
		Validate: validator.New(),
	}
}

func (h *CardHandler) CreateCard(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.CreateCardRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}

	hashedPIN, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to process PIN", nil)
	}

	cardNumber, maskedNumber := generateCardNumber()
	cvv := fmt.Sprintf("%03d", rand.Intn(900)+100)
	now := time.Now()

	card := models.Card{
		UserID:         userUUID,
		CardNumber:     cardNumber,
		MaskedNumber:   maskedNumber,
		CardHolderName: req.CardHolderName,
		CardType:       req.CardType,
		CardCategory:   req.CardCategory,
		ExpiryMonth:    int(now.Month()),
		ExpiryYear:     now.Year() + 3,
		CVV:            cvv,
		PIN:            string(hashedPIN),
		DailyLimit:     100000,
		Status:         "active",
	}

	if result := h.DB.Create(&card); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create card", result.Error.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Card created successfully", card)
}

func (h *CardHandler) GetUserCards(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var cards []models.Card
	if result := h.DB.Where("user_id = ?", userID).Find(&cards); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve cards", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Cards retrieved successfully", cards)
}

func (h *CardHandler) GetCardByID(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	cardID := c.Params("id")

	var card models.Card
	if result := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Card not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Card retrieved successfully", card)
}

func (h *CardHandler) UpdateCard(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	cardID := c.Params("id")
	req := new(models.UpdateCardRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var card models.Card
	if result := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Card not found", nil)
	}

	updates := map[string]interface{}{}
	if req.DailyLimit > 0 {
		updates["daily_limit"] = req.DailyLimit
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	h.DB.Model(&card).Updates(updates)

	return utils.SuccessResponse(c, fiber.StatusOK, "Card updated successfully", card)
}

func (h *CardHandler) ChangePIN(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	cardID := c.Params("id")
	req := new(models.ChangePINRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var card models.Card
	if result := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Card not found", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(card.PIN), []byte(req.OldPIN)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Incorrect current PIN", nil)
	}

	newHashedPIN, err := bcrypt.GenerateFromPassword([]byte(req.NewPIN), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update PIN", nil)
	}

	h.DB.Model(&card).Update("pin", string(newHashedPIN))

	return utils.SuccessResponse(c, fiber.StatusOK, "PIN changed successfully", nil)
}

func (h *CardHandler) BlockCard(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	cardID := c.Params("id")

	var card models.Card
	if result := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Card not found", nil)
	}

	h.DB.Model(&card).Update("status", "blocked")

	return utils.SuccessResponse(c, fiber.StatusOK, "Card blocked successfully", nil)
}

func (h *CardHandler) UnblockCard(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	cardID := c.Params("id")

	var card models.Card
	if result := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Card not found", nil)
	}

	h.DB.Model(&card).Update("status", "active")

	return utils.SuccessResponse(c, fiber.StatusOK, "Card unblocked successfully", nil)
}

func (h *CardHandler) DeleteCard(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	cardID := c.Params("id")

	var card models.Card
	if result := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Card not found", nil)
	}

	h.DB.Delete(&card)

	return utils.SuccessResponse(c, fiber.StatusOK, "Card deleted successfully", nil)
}

func generateCardNumber() (string, string) {
	prefix := "4532"
	middle := fmt.Sprintf("%08d", rand.Intn(100000000))
	last4 := fmt.Sprintf("%04d", rand.Intn(10000))
	full := prefix + middle + last4
	masked := fmt.Sprintf("**** **** **** %s", last4)
	return full, masked
}
