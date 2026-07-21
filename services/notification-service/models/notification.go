package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Title     string         `json:"title" gorm:"not null"`
	Body      string         `json:"body" gorm:"not null"`
	Type      string         `json:"type"`    // transaction, security, promotion, system
	Channel   string         `json:"channel"` // push, email, sms, in_app
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	ReadAt    *time.Time     `json:"read_at"`
	Data      string         `json:"data"` // JSON metadata
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type NotificationSetting struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID `json:"user_id" gorm:"type:uuid;unique;not null"`
	PushEnabled      bool      `json:"push_enabled" gorm:"default:true"`
	EmailEnabled     bool      `json:"email_enabled" gorm:"default:true"`
	SMSEnabled       bool      `json:"sms_enabled" gorm:"default:true"`
	TransactionAlert bool      `json:"transaction_alert" gorm:"default:true"`
	SecurityAlert    bool      `json:"security_alert" gorm:"default:true"`
	PromotionAlert   bool      `json:"promotion_alert" gorm:"default:false"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SendNotificationRequest struct {
	UserID  string `json:"user_id" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Body    string `json:"body" validate:"required"`
	Type    string `json:"type" validate:"required,oneof=transaction security promotion system"`
	Channel string `json:"channel" validate:"required,oneof=push email sms in_app"`
	Data    string `json:"data"`
}

type UpdateSettingsRequest struct {
	PushEnabled      *bool `json:"push_enabled"`
	EmailEnabled     *bool `json:"email_enabled"`
	SMSEnabled       *bool `json:"sms_enabled"`
	TransactionAlert *bool `json:"transaction_alert"`
	SecurityAlert    *bool `json:"security_alert"`
	PromotionAlert   *bool `json:"promotion_alert"`
}
