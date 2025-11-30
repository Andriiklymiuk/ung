package models

import (
	"time"

	"gorm.io/gorm"
)

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
	ContractNum  string       `gorm:"uniqueIndex;not null" json:"contract_num"` // e.g., "contract.acme.jan.2025"
	ClientID     uint         `gorm:"not null;index" json:"client_id"`
	Client       Client       `gorm:"foreignKey:ClientID" json:"-"`
	Name         string       `gorm:"not null" json:"name"` // e.g., "Website Development Q1 2025"
	ContractType ContractType `gorm:"not null" json:"contract_type"`
	HourlyRate   *float64     `json:"hourly_rate"` // For hourly contracts
	FixedPrice   *float64     `json:"fixed_price"` // For fixed price contracts
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
	Company     Company       `gorm:"foreignKey:CompanyID" json:"-"`
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
	Discount    float64   `gorm:"default:0" json:"discount"`     // Discount amount for this line item
	DiscountPct float64   `gorm:"default:0" json:"discount_pct"` // Discount percentage (0-100)
	TaxRate     float64   `gorm:"default:0" json:"tax_rate"`     // Tax rate for this item (0-1)
	TaxAmount   float64   `gorm:"default:0" json:"tax_amount"`   // Calculated tax amount
	CreatedAt   time.Time `json:"created_at"`
}

// TrackingSession represents a time tracking session
type TrackingSession struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ClientID    *uint          `gorm:"index" json:"client_id"`
	Client      *Client        `gorm:"foreignKey:ClientID" json:"-"`
	ContractID  *uint          `gorm:"index" json:"contract_id"`
	Contract    *Contract      `gorm:"foreignKey:ContractID" json:"-"`
	ProjectName string         `json:"project_name"`
	StartTime   time.Time      `gorm:"not null" json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	Duration    *int           `json:"duration"` // in seconds
	Hours       *float64       `json:"hours"`    // calculated hours for easier display
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

// RecurringFrequency represents how often a recurring invoice is generated
type RecurringFrequency string

const (
	FrequencyWeekly    RecurringFrequency = "weekly"
	FrequencyBiweekly  RecurringFrequency = "biweekly"
	FrequencyMonthly   RecurringFrequency = "monthly"
	FrequencyQuarterly RecurringFrequency = "quarterly"
	FrequencyYearly    RecurringFrequency = "yearly"
)

// RecurringInvoice represents a template for automatically generating invoices
type RecurringInvoice struct {
	ID                 uint               `gorm:"primaryKey" json:"id"`
	ClientID           uint               `gorm:"not null;index" json:"client_id"`
	Client             Client             `gorm:"foreignKey:ClientID" json:"-"`
	ContractID         *uint              `gorm:"index" json:"contract_id"`
	Contract           *Contract          `gorm:"foreignKey:ContractID" json:"-"`
	Amount             float64            `gorm:"not null" json:"amount"`
	Currency           string             `gorm:"default:USD" json:"currency"`
	Description        string             `json:"description"`
	Frequency          RecurringFrequency `gorm:"not null" json:"frequency"`
	DayOfMonth         int                `gorm:"default:1" json:"day_of_month"` // 1-28 for monthly
	DayOfWeek          int                `gorm:"default:1" json:"day_of_week"`  // 1-7 for weekly (1=Monday)
	NextGenerationDate time.Time          `json:"next_generation_date"`
	LastGeneratedDate  *time.Time         `json:"last_generated_date"`
	LastInvoiceID      *uint              `json:"last_invoice_id"`
	Active             bool               `gorm:"default:true" json:"active"`
	AutoSend           bool               `gorm:"default:false" json:"auto_send"`    // Auto-send email when generated
	AutoPDF            bool               `gorm:"default:true" json:"auto_pdf"`      // Auto-generate PDF
	EmailApp           string             `gorm:"column:email_app" json:"email_app"` // apple, outlook, gmail
	GeneratedCount     int                `gorm:"default:0" json:"generated_count"`  // How many invoices generated
	Notes              string             `json:"notes"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// UserSettings stores user preferences and dashboard settings
type UserSettings struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Settings keys
const (
	SettingWeeklyHoursTarget = "weekly_hours_target"
)

// =====================================
// Job Hunter Models
// =====================================

// Profile represents the user's professional profile extracted from CV/resume
type Profile struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `json:"name"`
	Title      string    `json:"title"`      // e.g., "Senior Go Developer"
	Bio        string    `json:"bio"`        // Professional summary
	Skills     string    `json:"skills"`     // JSON array of skills
	Experience int       `json:"experience"` // Years of experience
	Rate       float64   `json:"rate"`       // Desired hourly rate
	Currency   string    `gorm:"default:USD" json:"currency"`
	Location   string    `json:"location"`
	Remote     bool      `gorm:"default:true" json:"remote"`
	Languages  string    `json:"languages"`                       // JSON array of languages
	Education  string    `json:"education"`                       // JSON array of education
	Projects   string    `json:"projects"`                        // JSON array of notable projects
	Links      string    `json:"links"`                           // JSON: github, linkedin, portfolio
	PDFPath    string    `gorm:"column:pdf_path" json:"pdf_path"` // Original CV path
	PDFContent string    `json:"pdf_content"`                     // Extracted text from PDF
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// JobSource represents where jobs are scraped from
type JobSource string

const (
	JobSourceHN             JobSource = "hackernews"
	JobSourceRemoteOK       JobSource = "remoteok"
	JobSourceWeWorkRemotely JobSource = "weworkremotely"
	JobSourceJobicy         JobSource = "jobicy"
	JobSourceArbeitnow      JobSource = "arbeitnow"
	JobSourceDjinni         JobSource = "djinni"      // Ukrainian IT jobs
	JobSourceDOU            JobSource = "dou"         // Ukrainian IT community
	JobSourceNetherlands    JobSource = "netherlands" // Dutch job boards
	JobSourceEuroJobs       JobSource = "eurojobs"    // European job boards
	JobSourceUpwork         JobSource = "upwork"
	JobSourceLinkedIn       JobSource = "linkedin"
	JobSourceManual         JobSource = "manual"
)

// Job represents a scraped job posting
type Job struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Source      JobSource  `gorm:"not null" json:"source"`
	SourceID    string     `json:"source_id"`  // External ID from source
	SourceURL   string     `json:"source_url"` // Link to original posting
	Title       string     `gorm:"not null" json:"title"`
	Company     string     `json:"company"`
	Description string     `json:"description"`
	Skills      string     `json:"skills"`    // JSON array of required skills
	RateMin     float64    `json:"rate_min"`  // Minimum rate/salary
	RateMax     float64    `json:"rate_max"`  // Maximum rate/salary
	RateType    string     `json:"rate_type"` // hourly, monthly, yearly
	Currency    string     `gorm:"default:USD" json:"currency"`
	Remote      bool       `gorm:"default:true" json:"remote"`
	Location    string     `json:"location"`
	JobType     string     `json:"job_type"`    // contract, fulltime, parttime
	MatchScore  float64    `json:"match_score"` // 0-100 match with profile
	PostedAt    time.Time  `json:"posted_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ApplicationStatus represents the status of a job application
type ApplicationStatus string

const (
	AppStatusDraft     ApplicationStatus = "draft"
	AppStatusApplied   ApplicationStatus = "applied"
	AppStatusViewed    ApplicationStatus = "viewed"
	AppStatusResponse  ApplicationStatus = "response"
	AppStatusInterview ApplicationStatus = "interview"
	AppStatusOffer     ApplicationStatus = "offer"
	AppStatusRejected  ApplicationStatus = "rejected"
	AppStatusWithdrawn ApplicationStatus = "withdrawn"
)

// Application represents a job application
type Application struct {
	ID          uint              `gorm:"primaryKey" json:"id"`
	JobID       uint              `gorm:"not null;index" json:"job_id"`
	Job         Job               `gorm:"foreignKey:JobID" json:"-"`
	ProfileID   uint              `gorm:"index" json:"profile_id"`
	Profile     Profile           `gorm:"foreignKey:ProfileID" json:"-"`
	Proposal    string            `json:"proposal"`     // Generated proposal text
	ProposalPDF string            `json:"proposal_pdf"` // Path to PDF proposal
	CoverLetter string            `json:"cover_letter"`
	Status      ApplicationStatus `gorm:"default:draft" json:"status"`
	Notes       string            `json:"notes"`
	AppliedAt   *time.Time        `json:"applied_at"`
	ResponseAt  *time.Time        `json:"response_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// =====================================
// Gig Models - Central Work Unit
// =====================================

// GigStatus represents the workflow status of a gig
// Simplified flow: todo → in_progress → sent → done
type GigStatus string

const (
	GigStatusTodo       GigStatus = "todo"        // Queued, not started
	GigStatusInProgress GigStatus = "in_progress" // Actively working
	GigStatusSent       GigStatus = "sent"        // Delivered, awaiting payment
	GigStatusDone       GigStatus = "done"        // Completed & paid
	GigStatusOnHold     GigStatus = "on_hold"     // Paused
	GigStatusCancelled  GigStatus = "cancelled"   // Cancelled
)

// GigType represents the type of work arrangement
type GigType string

const (
	GigTypeHourly   GigType = "hourly"
	GigTypeFixed    GigType = "fixed"
	GigTypeRetainer GigType = "retainer"
)

// GigPriority represents urgency level
type GigPriority int

const (
	GigPriorityNormal GigPriority = 0
	GigPriorityHigh   GigPriority = 1
	GigPriorityUrgent GigPriority = 2
)

// Gig represents a work project/engagement - the central unit connecting hunter → work → invoice
type Gig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"not null" json:"name"`
	ClientID      *uint     `gorm:"index" json:"client_id"`
	Client        *Client   `gorm:"foreignKey:ClientID" json:"-"`
	ContractID    *uint     `gorm:"index" json:"contract_id"`
	Contract      *Contract `gorm:"foreignKey:ContractID" json:"-"`
	ApplicationID *uint     `gorm:"index" json:"application_id"` // From hunter

	Status   GigStatus   `gorm:"not null;default:todo" json:"status"`
	GigType  GigType     `gorm:"default:hourly" json:"gig_type"`
	Priority GigPriority `gorm:"default:0" json:"priority"`
	Project  string      `gorm:"index" json:"project"` // Optional project grouping

	// Financial
	EstimatedHours  *float64 `json:"estimated_hours"`
	EstimatedAmount *float64 `json:"estimated_amount"`
	HourlyRate      *float64 `json:"hourly_rate"`
	Currency        string   `gorm:"default:USD" json:"currency"`

	// Tracking aggregation
	TotalHoursTracked float64    `gorm:"default:0" json:"total_hours_tracked"`
	LastTrackedAt     *time.Time `json:"last_tracked_at"`

	// Billing
	TotalInvoiced  float64    `gorm:"default:0" json:"total_invoiced"`
	LastInvoicedAt *time.Time `json:"last_invoiced_at"`

	// Dates
	StartDate   *time.Time `json:"start_date"`
	DueDate     *time.Time `json:"due_date"`
	CompletedAt *time.Time `json:"completed_at"`

	// Metadata
	Description string `json:"description"`
	Notes       string `json:"notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WorkLogType represents the type of work log entry
type WorkLogType string

const (
	WorkLogTypeNote      WorkLogType = "note"
	WorkLogTypeDecision  WorkLogType = "decision"
	WorkLogTypeMeeting   WorkLogType = "meeting"
	WorkLogTypeBlocker   WorkLogType = "blocker"
	WorkLogTypeMilestone WorkLogType = "milestone"
)

// WorkLog represents a note/log entry attached to work
type WorkLog struct {
	ID                uint             `gorm:"primaryKey" json:"id"`
	GigID             *uint            `gorm:"index" json:"gig_id"`
	Gig               *Gig             `gorm:"foreignKey:GigID" json:"-"`
	ClientID          *uint            `gorm:"index" json:"client_id"`
	Client            *Client          `gorm:"foreignKey:ClientID" json:"-"`
	TrackingSessionID *uint            `gorm:"index" json:"tracking_session_id"`
	TrackingSession   *TrackingSession `gorm:"foreignKey:TrackingSessionID" json:"-"`

	Content string      `gorm:"not null" json:"content"`
	LogType WorkLogType `gorm:"default:note" json:"log_type"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GigTask represents an actionable task within a gig
type GigTask struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	GigID       uint       `gorm:"not null;index" json:"gig_id"`
	Gig         *Gig       `gorm:"foreignKey:GigID" json:"-"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	Completed   bool       `gorm:"default:false" json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
	DueDate     *time.Time `json:"due_date"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// =====================================
// Enhanced Goals
// =====================================

// GoalType represents different types of goals
type GoalType string

const (
	GoalTypeIncome  GoalType = "income"
	GoalTypeHours   GoalType = "hours"
	GoalTypeClients GoalType = "clients"
	GoalTypeSavings GoalType = "savings"
)

// IncomeGoal represents an enhanced income/hours/client goal
type IncomeGoal struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	Amount      float64  `gorm:"not null" json:"amount"`
	Period      string   `gorm:"not null" json:"period"` // monthly, quarterly, yearly
	Year        int      `gorm:"not null" json:"year"`
	Month       int      `json:"month"`   // for monthly goals
	Quarter     int      `json:"quarter"` // for quarterly goals
	Description string   `json:"description"`
	GoalType    GoalType `gorm:"default:income" json:"goal_type"`

	// For different goal types
	TargetHours    *float64 `json:"target_hours"`    // hours goal
	TargetClients  *int     `json:"target_clients"`  // clients goal
	SavingsPercent *float64 `json:"savings_percent"` // savings goal (e.g., 30%)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
