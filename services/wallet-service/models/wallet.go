package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wallet struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID      `json:"user_id" gorm:"type:uuid;unique;not null"`
	AccountNumber string         `json:"account_number" gorm:"unique;not null"`
	Balance       float64        `json:"balance" gorm:"default:0"`
	Currency      string         `json:"currency" gorm:"default:'NGN'"`
	Status        string         `json:"status" gorm:"default:'active'"` // active, frozen, closed
	DailyLimit    float64        `json:"daily_limit" gorm:"default:500000"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type WalletTransaction struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WalletID      uuid.UUID `json:"wallet_id" gorm:"type:uuid;not null;index"`
	Type          string    `json:"type"` // credit, debit
	Amount        float64   `json:"amount"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	Description   string    `json:"description"`
	Reference     string    `json:"reference" gorm:"unique"`
	Status        string    `json:"status" gorm:"default:'success'"`
	CreatedAt     time.Time `json:"created_at"`
}

type FundWalletRequest struct {
	Amount      float64 `json:"amount" validate:"required,min=100"`
	Description string  `json:"description"`
}

type TransferRequest struct {
	ToAccountNumber string  `json:"to_account_number" validate:"required"`
	Amount          float64 `json:"amount" validate:"required,min=100"`
	Description     string  `json:"description"`
	PIN             string  `json:"pin" validate:"required,len=4"`
}

type WithdrawRequest struct {
	Amount      float64 `json:"amount" validate:"required,min=100"`
	Description string  `json:"description"`
	PIN         string  `json:"pin" validate:"required,len=4"`
}
