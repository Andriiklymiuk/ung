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

// Business Models (shared with CLI)

// Company represents a business entity (the user's company)
type Company struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Name                string    `gorm:"not null" json:"name"`
	Email               string    `gorm:"not null" json:"email"`
	Phone               string    `json:"phone"`
	Address             string    `json:"address"`
	RegistrationAddress string    `gorm:"column:registration_address" json:"registration_address"`
	TaxID               string    `gorm:"column:tax_id" json:"tax_id"`
	BankName            string    `gorm:"column:bank_name" json:"bank_name"`
	BankAccount         string    `gorm:"column:bank_account" json:"bank_account"`
	BankSWIFT           string    `gorm:"column:bank_swift" json:"bank_swift"`
	LogoPath            string    `gorm:"column:logo_path" json:"logo_path"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Client represents a customer/client
type Client struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"not null" json:"email"`
	Address   string    `json:"address"`
	TaxID     string    `gorm:"column:tax_id" json:"tax_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContractType represents the type of contract
type ContractType string

const (
	ContractTypeHourly     ContractType = "hourly"
	ContractTypeFixedPrice ContractType = "fixed_price"
	ContractTypeRetainer   ContractType = "retainer"
)

// Contract represents a work contract with a client
type Contract struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	ContractNum  string       `gorm:"uniqueIndex;not null" json:"contract_num"`
	ClientID     uint         `gorm:"not null;index" json:"client_id"`
	Client       Client       `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Name         string       `gorm:"not null" json:"name"`
	ContractType ContractType `gorm:"not null" json:"contract_type"`
	HourlyRate   *float64     `json:"hourly_rate"`
	FixedPrice   *float64     `json:"fixed_price"`
	Currency     string       `gorm:"default:USD" json:"currency"`
	StartDate    time.Time    `json:"start_date"`
	EndDate      *time.Time   `json:"end_date"`
	Active       bool         `gorm:"default:true" json:"active"`
	Notes        string       `json:"notes"`
	PDFPath      string       `gorm:"column:pdf_path" json:"pdf_path"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	StatusPending InvoiceStatus = "pending"
	StatusSent    InvoiceStatus = "sent"
	StatusPaid    InvoiceStatus = "paid"
	StatusOverdue InvoiceStatus = "overdue"
)

// Invoice represents an invoice
type Invoice struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	InvoiceNum  string        `gorm:"uniqueIndex;not null" json:"invoice_num"`
	CompanyID   uint          `gorm:"not null;index" json:"company_id"`
	Company     Company       `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Amount      float64       `gorm:"not null" json:"amount"`
	Currency    string        `gorm:"default:USD" json:"currency"`
	Description string        `json:"description"`
	Status      InvoiceStatus `gorm:"default:pending" json:"status"`
	IssuedDate  time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"issued_date"`
	DueDate     time.Time     `json:"due_date"`
	PDFPath     string        `gorm:"column:pdf_path" json:"pdf_path"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// InvoiceRecipient links invoices to clients
type InvoiceRecipient struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	InvoiceID uint `gorm:"not null;index" json:"invoice_id"`
	ClientID  uint `gorm:"not null;index" json:"client_id"`
}

// InvoiceLineItem represents an item/service on an invoice
type InvoiceLineItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	InvoiceID   uint      `gorm:"not null;index" json:"invoice_id"`
	Invoice     Invoice   `gorm:"foreignKey:InvoiceID" json:"-"`
	ItemName    string    `gorm:"column:item_name;not null" json:"item_name"`
	Description string    `json:"description"`
	Quantity    float64   `gorm:"not null;default:1" json:"quantity"`
	Rate        float64   `gorm:"not null" json:"rate"`
	Amount      float64   `gorm:"not null" json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// TrackingSession represents a time tracking session
type TrackingSession struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ClientID    *uint          `gorm:"index" json:"client_id"`
	Client      *Client        `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	ContractID  *uint          `gorm:"index" json:"contract_id"`
	Contract    *Contract      `gorm:"foreignKey:ContractID" json:"contract,omitempty"`
	ProjectName string         `json:"project_name"`
	StartTime   time.Time      `gorm:"not null" json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	Duration    *int           `json:"duration"`
	Hours       *float64       `json:"hours"`
	Billable    bool           `gorm:"default:true" json:"billable"`
	Notes       string         `json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// ExpenseCategory represents the category of an expense
type ExpenseCategory string

const (
	ExpenseCategorySoftware       ExpenseCategory = "software"
	ExpenseCategoryHardware       ExpenseCategory = "hardware"
	ExpenseCategoryTravel         ExpenseCategory = "travel"
	ExpenseCategoryMeals          ExpenseCategory = "meals"
	ExpenseCategoryOfficeSupplies ExpenseCategory = "office_supplies"
	ExpenseCategoryUtilities      ExpenseCategory = "utilities"
	ExpenseCategoryMarketing      ExpenseCategory = "marketing"
	ExpenseCategoryOther          ExpenseCategory = "other"
)

// Expense represents a business expense
type Expense struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Description string          `gorm:"not null" json:"description"`
	Amount      float64         `gorm:"not null" json:"amount"`
	Currency    string          `gorm:"default:USD" json:"currency"`
	Category    ExpenseCategory `gorm:"not null" json:"category"`
	Date        time.Time       `gorm:"not null" json:"date"`
	Vendor      string          `json:"vendor"`
	ReceiptPath string          `gorm:"column:receipt_path" json:"receipt_path"`
	Notes       string          `json:"notes"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
