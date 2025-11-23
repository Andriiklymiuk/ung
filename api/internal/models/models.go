package models

import (
	"time"
	"gorm.io/gorm"
)

// User represents an API user account
type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Email           string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash    string         `gorm:"not null" json:"-"`
	Name            string         `gorm:"not null" json:"name"`
	DBPath          string         `gorm:"not null" json:"-"` // Path to user's ung.db
	SubscriptionID  *string        `json:"subscription_id"`   // RevenueCat/Stripe
	PlanType        string         `gorm:"default:free" json:"plan_type"` // free, pro, business
	Active          bool           `gorm:"default:true" json:"active"`
	EmailVerified   bool           `gorm:"default:false" json:"email_verified"`
	GmailToken      *string        `json:"-"` // Encrypted Gmail OAuth token
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// RefreshToken represents a JWT refresh token
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"user_id"`
	Key       string     `gorm:"uniqueIndex;not null" json:"key"`
	Name      string     `gorm:"not null" json:"name"`
	LastUsed  *time.Time `json:"last_used"`
	ExpiresAt *time.Time `json:"expires_at"`
	Active    bool       `gorm:"default:true" json:"active"`
	CreatedAt time.Time  `json:"created_at"`
}

// StandardResponse is the standard API response format
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse adds pagination metadata
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}
