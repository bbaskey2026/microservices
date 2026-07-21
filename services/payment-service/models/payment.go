package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	Reference     string         `json:"reference" gorm:"unique;not null"`
	Amount        float64        `json:"amount" gorm:"not null"`
	Currency      string         `json:"currency" gorm:"default:'NGN'"`
	Type          string         `json:"type"`                            // bill_payment, airtime, data, transfer
	Category      string         `json:"category"`                        // electricity, water, internet, etc
	Provider      string         `json:"provider"`                        // DSTV, IKEDC, etc
	AccountRef    string         `json:"account_ref"`                     // meter no, phone, etc
	Status        string         `json:"status" gorm:"default:'pending'"` // pending, success, failed
	Description   string         `json:"description"`
	FailureReason string         `json:"failure_reason,omitempty"`
	ProcessedAt   *time.Time     `json:"processed_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type BillPaymentRequest struct {
	Amount      float64 `json:"amount" validate:"required,min=100"`
	Category    string  `json:"category" validate:"required"` // electricity, water, cable, internet
	Provider    string  `json:"provider" validate:"required"`
	AccountRef  string  `json:"account_ref" validate:"required"`
	Description string  `json:"description"`
}

type AirtimeRequest struct {
	PhoneNumber string  `json:"phone_number" validate:"required,len=11"`
	Amount      float64 `json:"amount" validate:"required,min=50"`
	Network     string  `json:"network" validate:"required,oneof=MTN Airtel Glo 9mobile"`
}

type DataBundleRequest struct {
	PhoneNumber string  `json:"phone_number" validate:"required,len=11"`
	BundleCode  string  `json:"bundle_code" validate:"required"`
	Amount      float64 `json:"amount" validate:"required"`
	Network     string  `json:"network" validate:"required,oneof=MTN Airtel Glo 9mobile"`
}
