package models

import "time"

// User represents a Telegram user
type User struct {
	TelegramID int64
	APIToken   string
	UserID     uint
	Name       string
	Email      string
	CreatedAt  time.Time
}

// Session represents a conversation state
type Session struct {
	TelegramID int64
	State      string
	Data       map[string]interface{}
	UpdatedAt  time.Time
}

// Command represents bot command data
type Command string

const (
	CommandStart    Command = "/start"
	CommandHelp     Command = "/help"
	CommandInvoice  Command = "/invoice"
	CommandClient   Command = "/client"
	CommandContract Command = "/contract"
	CommandTrack    Command = "/track"
	CommandReport   Command = "/report"
	CommandStatus   Command = "/status"
	CommandSettings Command = "/settings"
)

// SessionState represents conversation states
type SessionState string

const (
	StateNone                SessionState = ""
	// Invoice states
	StateInvoiceSelectClient SessionState = "invoice_select_client"
	StateInvoiceAmount       SessionState = "invoice_amount"
	StateInvoiceDescription  SessionState = "invoice_description"
	StateInvoiceDueDate      SessionState = "invoice_due_date"
	// Client states
	StateClientCreateName    SessionState = "client_create_name"
	StateClientCreateEmail   SessionState = "client_create_email"
	StateClientCreateAddress SessionState = "client_create_address"
	StateClientCreateTaxID   SessionState = "client_create_tax_id"
	// Company states
	StateCompanyCreateName    SessionState = "company_create_name"
	StateCompanyCreateEmail   SessionState = "company_create_email"
	StateCompanyCreatePhone   SessionState = "company_create_phone"
	StateCompanyCreateAddress SessionState = "company_create_address"
	StateCompanyCreateTaxID   SessionState = "company_create_tax_id"
	// Contract states
	StateContractSelectClient SessionState = "contract_select_client"
	StateContractName         SessionState = "contract_name"
	StateContractType         SessionState = "contract_type"
	StateContractRate         SessionState = "contract_rate"
	// Expense states
	StateExpenseDescription SessionState = "expense_description"
	StateExpenseAmount      SessionState = "expense_amount"
	StateExpenseCategory    SessionState = "expense_category"
	StateExpenseVendor      SessionState = "expense_vendor"
)
