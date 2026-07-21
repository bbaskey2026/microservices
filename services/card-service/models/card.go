package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Card struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	CardNumber     string         `json:"card_number" gorm:"unique;not null"`
	MaskedNumber   string         `json:"masked_number"`
	CardHolderName string         `json:"card_holder_name" gorm:"not null"`
	CardType       string         `json:"card_type" gorm:"not null"`            // visa, mastercard, verve
	CardCategory   string         `json:"card_category" gorm:"default:'debit'"` // debit, credit
	ExpiryMonth    int            `json:"expiry_month"`
	ExpiryYear     int            `json:"expiry_year"`
	CVV            string         `json:"-"`
	PIN            string         `json:"-"`
	DailyLimit     float64        `json:"daily_limit" gorm:"default:100000"`
	Status         string         `json:"status" gorm:"default:'active'"` // active, blocked, expired
	IsDefault      bool           `json:"is_default" gorm:"default:false"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateCardRequest struct {
	CardType       string `json:"card_type" validate:"required,oneof=visa mastercard verve"`
	CardCategory   string `json:"card_category" validate:"required,oneof=debit credit"`
	CardHolderName string `json:"card_holder_name" validate:"required"`
	PIN            string `json:"pin" validate:"required,len=4"`
}

type UpdateCardRequest struct {
	DailyLimit float64 `json:"daily_limit" validate:"omitempty,min=1000"`
	Status     string  `json:"status" validate:"omitempty,oneof=active blocked"`
	IsDefault  *bool   `json:"is_default"`
}

type ChangePINRequest struct {
	OldPIN string `json:"old_pin" validate:"required,len=4"`
	NewPIN string `json:"new_pin" validate:"required,len=4"`
}
