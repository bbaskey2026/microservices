package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FirstName   string         `json:"first_name" gorm:"not null"`
	LastName    string         `json:"last_name" gorm:"not null"`
	Email       string         `json:"email" gorm:"unique;not null"`
	Password    string         `json:"-" gorm:"not null"`
	Phone       string         `json:"phone" gorm:"unique"`
	DateOfBirth *time.Time     `json:"date_of_birth"`
	Address     string         `json:"address"`
	Avatar      string         `json:"avatar"`
	Role        string         `json:"role" gorm:"default:'user'"`          // user, admin
	Status      string         `json:"status" gorm:"default:'active'"`      // active, inactive, suspended
	KYCStatus   string         `json:"kyc_status" gorm:"default:'pending'"` // pending, verified, rejected
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type UpdateUserRequest struct {
	FirstName   string     `json:"first_name" validate:"omitempty,min=2"`
	LastName    string     `json:"last_name" validate:"omitempty,min=2"`
	Phone       string     `json:"phone" validate:"omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Address     string     `json:"address"`
	Avatar      string     `json:"avatar"`
}

type KYCUpdateRequest struct {
	KYCStatus string `json:"kyc_status" validate:"required,oneof=pending verified rejected"`
}
