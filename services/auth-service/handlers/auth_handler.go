package handlers

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"banking-app/services/auth-service/models"
	userModels "banking-app/services/user-service/models"
	"banking-app/shared/utils"
)

type AuthHandler struct {
	DB        *gorm.DB
	Redis     *redis.Client
	JWTSecret string
	Validate  *validator.Validate
}

func NewAuthHandler(db *gorm.DB, redisClient *redis.Client, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		DB:        db,
		Redis:     redisClient,
		JWTSecret: jwtSecret,
		Validate:  validator.New(),
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(models.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var existingUser userModels.User
	if result := h.DB.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		return utils.ErrorResponse(c, fiber.StatusConflict, "Email already registered", nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", nil)
	}

	user := userModels.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Phone:     req.Phone,
		Role:      "user",
		Status:    "active",
		KYCStatus: "pending",
	}

	if result := h.DB.Create(&user); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", result.Error.Error())
	}

	tokenResp, err := h.generateTokens(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate tokens", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User registered successfully", tokenResp)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(models.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var user userModels.User
	if result := h.DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials", nil)
	}

	if user.Status != "active" {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Account is not active", nil)
	}

	tokenResp, err := h.generateTokens(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate tokens", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", tokenResp)
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	req := new(models.RefreshRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var refreshToken models.RefreshToken
	if result := h.DB.Where("token = ? AND is_revoked = ? AND expires_at > ?", req.RefreshToken, false, time.Now()).First(&refreshToken); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired refresh token", nil)
	}

	var user userModels.User
	if result := h.DB.Where("id = ?", refreshToken.UserID).First(&user); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Revoke old refresh token
	h.DB.Model(&refreshToken).Update("is_revoked", true)

	tokenResp, err := h.generateTokens(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate tokens", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Token refreshed successfully", tokenResp)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if ok && userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			h.DB.Model(&models.RefreshToken{}).Where("user_id = ?", userUUID).Update("is_revoked", true)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Logged out successfully", nil)
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	req := new(models.ChangePasswordRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.Validate.Struct(req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Validation failed", err.Error())
	}

	var user userModels.User
	if result := h.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Incorrect current password", nil)
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash new password", nil)
	}

	h.DB.Model(&user).Update("password", string(newHashedPassword))

	return utils.SuccessResponse(c, fiber.StatusOK, "Password changed successfully", nil)
}

func (h *AuthHandler) generateTokens(userID, email, role string) (*models.TokenResponse, error) {
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := &utils.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenStr := uuid.New().String()

	userUUID, _ := uuid.Parse(userID)
	refreshToken := models.RefreshToken{
		UserID:    userUUID,
		Token:     refreshTokenStr,
		ExpiresAt: refreshExpiry,
	}
	h.DB.Create(&refreshToken)

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}
