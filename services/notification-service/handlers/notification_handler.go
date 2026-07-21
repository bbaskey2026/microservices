package handlers

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"banking-app/services/notification-service/models"
	"banking-app/shared/utils"
)

type NotificationHandler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{
		DB:       db,
		Validate: validator.New(),
	}
}

func (h *NotificationHandler) SendNotification(c *fiber.Ctx) error {
	req := new(models.SendNotificationRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	userUUID, _ := uuid.Parse(req.UserID)
	notification := models.Notification{
		UserID:  userUUID,
		Title:   req.Title,
		Body:    req.Body,
		Type:    req.Type,
		Channel: req.Channel,
		Data:    req.Data,
	}

	if result := h.DB.Create(&notification); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to send notification", nil)
	}

	// TODO: Send via appropriate channel (Firebase, SendGrid, Twilio)

	return utils.SuccessResponse(c, fiber.StatusCreated, "Notification sent successfully", notification)
}

func (h *NotificationHandler) GetUserNotifications(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	isRead := c.Query("is_read", "")

	offset := (page - 1) * perPage
	query := h.DB.Where("user_id = ?", userID)

	if isRead == "true" {
		query = query.Where("is_read = ?", true)
	} else if isRead == "false" {
		query = query.Where("is_read = ?", false)
	}

	var total int64
	query.Model(&models.Notification{}).Count(&total)

	var notifications []models.Notification
	query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&notifications)

	return utils.PaginatedResponse(c, notifications, &utils.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	})
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	notifID := c.Params("id")

	var notification models.Notification
	if result := h.DB.Where("id = ? AND user_id = ?", notifID, userID).First(&notification); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Notification not found", nil)
	}

	now := time.Now()
	h.DB.Model(&notification).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": &now,
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Notification marked as read", nil)
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	now := time.Now()

	h.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})

	return utils.SuccessResponse(c, fiber.StatusOK, "All notifications marked as read", nil)
}

func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var count int64
	h.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count)

	return utils.SuccessResponse(c, fiber.StatusOK, "Unread count retrieved", fiber.Map{
		"unread_count": count,
	})
}

func (h *NotificationHandler) GetSettings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var settings models.NotificationSetting
	if result := h.DB.Where("user_id = ?", userID).First(&settings); result.Error != nil {
		// Create default settings
		userUUID, _ := uuid.Parse(userID)
		settings = models.NotificationSetting{
			UserID:           userUUID,
			PushEnabled:      true,
			EmailEnabled:     true,
			SMSEnabled:       true,
			TransactionAlert: true,
			SecurityAlert:    true,
			PromotionAlert:   false,
		}
		h.DB.Create(&settings)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Settings retrieved", settings)
}

func (h *NotificationHandler) UpdateSettings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.UpdateSettingsRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var settings models.NotificationSetting
	if result := h.DB.Where("user_id = ?", userID).First(&settings); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Settings not found", nil)
	}

	updates := map[string]interface{}{}
	if req.PushEnabled != nil {
		updates["push_enabled"] = *req.PushEnabled
	}
	if req.EmailEnabled != nil {
		updates["email_enabled"] = *req.EmailEnabled
	}
	if req.SMSEnabled != nil {
		updates["sms_enabled"] = *req.SMSEnabled
	}
	if req.TransactionAlert != nil {
		updates["transaction_alert"] = *req.TransactionAlert
	}
	if req.SecurityAlert != nil {
		updates["security_alert"] = *req.SecurityAlert
	}
	if req.PromotionAlert != nil {
		updates["promotion_alert"] = *req.PromotionAlert
	}

	h.DB.Model(&settings).Updates(updates)

	return utils.SuccessResponse(c, fiber.StatusOK, "Settings updated successfully", settings)
}

func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	notifID := c.Params("id")

	var notification models.Notification
	if result := h.DB.Where("id = ? AND user_id = ?", notifID, userID).First(&notification); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Notification not found", nil)
	}

	h.DB.Delete(&notification)

	return utils.SuccessResponse(c, fiber.StatusOK, "Notification deleted", nil)
}
