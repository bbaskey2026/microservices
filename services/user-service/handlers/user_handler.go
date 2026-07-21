package handlers

import (
	"math"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"banking-app/services/user-service/models"
	"banking-app/shared/utils"
)

type UserHandler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		DB:       db,
		Validate: validator.New(),
	}
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var user models.User
	if result := h.DB.First(&user, "id = ?", userID); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile retrieved successfully", user)
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.UpdateUserRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var user models.User
	if result := h.DB.First(&user, "id = ?", userID); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	updates := map[string]interface{}{}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.DateOfBirth != nil {
		updates["date_of_birth"] = req.DateOfBirth
	}

	h.DB.Model(&user).Updates(updates)

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile updated successfully", user)
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	search := c.Query("search", "")
	status := c.Query("status", "")

	offset := (page - 1) * perPage
	query := h.DB.Model(&models.User{})

	if search != "" {
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var users []models.User
	query.Offset(offset).Limit(perPage).Find(&users)

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return utils.PaginatedResponse(c, users, &utils.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User

	if result := h.DB.First(&user, "id = ?", id); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User retrieved successfully", user)
}

func (h *UserHandler) UpdateUserStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive suspended"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var user models.User
	if result := h.DB.First(&user, "id = ?", id); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	h.DB.Model(&user).Update("status", req.Status)

	return utils.SuccessResponse(c, fiber.StatusOK, "User status updated successfully", user)
}

func (h *UserHandler) UpdateKYCStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	req := new(models.KYCUpdateRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var user models.User
	if result := h.DB.First(&user, "id = ?", id); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	h.DB.Model(&user).Update("kyc_status", req.KYCStatus)

	return utils.SuccessResponse(c, fiber.StatusOK, "KYC status updated successfully", user)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var user models.User
	if result := h.DB.First(&user, "id = ?", id); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	h.DB.Delete(&user)

	return utils.SuccessResponse(c, fiber.StatusOK, "User deleted successfully", nil)
}
