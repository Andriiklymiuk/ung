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
	// Tracking log states (manual time entry)
	StateTrackLogSelectContract SessionState = "track_log_select_contract"
	StateTrackLogHours          SessionState = "track_log_hours"
	StateTrackLogProject        SessionState = "track_log_project"
	StateTrackLogNotes          SessionState = "track_log_notes"
	// Search states
	StateSearchQuery            SessionState = "search_query"
	// Pomodoro states
	StatePomodoroProject        SessionState = "pomodoro_project"
	// Goal states
	StateGoalAmount             SessionState = "goal_amount"
	// Hunter states
	StateHunterProfileName      SessionState = "hunter_profile_name"
	StateHunterProfileTitle     SessionState = "hunter_profile_title"
	StateHunterProfileSkills    SessionState = "hunter_profile_skills"
	StateHunterProfileRate      SessionState = "hunter_profile_rate"
	StateHunterAwaitingPDF      SessionState = "hunter_awaiting_pdf"
	// Gig states
	StateGigCreateName          SessionState = "gig_create_name"
	StateGigCreateSelectClient  SessionState = "gig_create_select_client"
	StateGigTaskAdd             SessionState = "gig_task_add"
)
