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

// UserSettings represents user preferences for rate calculations
type UserSettings struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	HoursPerWeek      float64   `gorm:"default:40" json:"hours_per_week"`       // Billable hours per week
	WeeksPerYear      int       `gorm:"default:48" json:"weeks_per_year"`       // Working weeks per year
	DefaultTaxPercent float64   `gorm:"default:25" json:"default_tax_percent"`  // Default tax rate
	DefaultMargin     float64   `gorm:"default:20" json:"default_margin"`       // Default profit margin
	AnnualExpenses    float64   `gorm:"default:0" json:"annual_expenses"`       // Annual business expenses
	DefaultCurrency   string    `gorm:"default:USD" json:"default_currency"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// IncomeGoal represents an income goal for a specific period
type IncomeGoal struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Amount      float64   `gorm:"not null" json:"amount"`
	Period      string    `gorm:"not null" json:"period"` // monthly, quarterly, yearly
	Year        int       `gorm:"not null" json:"year"`
	Month       int       `json:"month"`   // for monthly goals (1-12)
	Quarter     int       `json:"quarter"` // for quarterly goals (1-4)
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RevenueCatEntitlement represents a subscription entitlement from RevenueCat
type RevenueCatEntitlement struct {
	IsActive          bool       `json:"is_active"`
	ProductIdentifier string     `json:"product_identifier"`
	ExpiresDate       *time.Time `json:"expires_date"`
	PurchaseDate      time.Time  `json:"purchase_date"`
	WillRenew         bool       `json:"will_renew"`
}

// SubscriptionInfo represents the user's subscription status
type SubscriptionInfo struct {
	UserID       uint                             `json:"user_id"`
	PlanType     string                           `json:"plan_type"`
	IsActive     bool                             `json:"is_active"`
	Entitlements map[string]RevenueCatEntitlement `json:"entitlements"`
	ExpiresAt    *time.Time                       `json:"expires_at"`
}

// RecurringFrequency represents how often an invoice recurs
type RecurringFrequency string

const (
	FrequencyWeekly    RecurringFrequency = "weekly"
	FrequencyBiweekly  RecurringFrequency = "biweekly"
	FrequencyMonthly   RecurringFrequency = "monthly"
	FrequencyQuarterly RecurringFrequency = "quarterly"
	FrequencyYearly    RecurringFrequency = "yearly"
)

// RecurringInvoice represents a recurring invoice template
type RecurringInvoice struct {
	ID            uint               `gorm:"primaryKey" json:"id"`
	ClientID      uint               `gorm:"not null;index" json:"client_id"`
	Client        Client             `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	CompanyID     uint               `gorm:"not null;index" json:"company_id"`
	Company       Company            `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Amount        float64            `gorm:"not null" json:"amount"`
	Currency      string             `gorm:"default:USD" json:"currency"`
	Description   string             `json:"description"`
	Frequency     RecurringFrequency `gorm:"not null" json:"frequency"`
	DayOfMonth    int                `gorm:"default:1" json:"day_of_month"`
	DayOfWeek     int                `json:"day_of_week"`
	NextRunDate   time.Time          `json:"next_run_date"`
	LastRunDate   *time.Time         `json:"last_run_date"`
	Active        bool               `gorm:"default:true" json:"active"`
	InvoicePrefix string             `gorm:"default:REC" json:"invoice_prefix"`
	TotalGenerated int               `gorm:"default:0" json:"total_generated"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// PomodoroSession represents a pomodoro timer session
type PomodoroSession struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ContractID  *uint          `gorm:"index" json:"contract_id"`
	Contract    *Contract      `gorm:"foreignKey:ContractID" json:"contract,omitempty"`
	ClientID    *uint          `gorm:"index" json:"client_id"`
	Client      *Client        `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	ProjectName string         `json:"project_name"`
	Duration    int            `gorm:"default:25" json:"duration"` // Duration in minutes
	BreakTime   int            `gorm:"default:5" json:"break_time"`
	StartTime   time.Time      `gorm:"not null" json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	Completed   bool           `gorm:"default:false" json:"completed"`
	Notes       string         `json:"notes"`
	SessionType string         `gorm:"default:work" json:"session_type"` // work, short_break, long_break
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// InvoiceTemplate represents a reusable invoice template
type InvoiceTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Content     string    `gorm:"type:text" json:"content"` // HTML/Markdown template content
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	Variables   string    `gorm:"type:text" json:"variables"` // JSON list of template variables
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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
	Languages  string    `json:"languages"`  // JSON array of languages
	Education  string    `json:"education"`  // JSON array of education
	Projects   string    `json:"projects"`   // JSON array of notable projects
	Links      string    `json:"links"`      // JSON: github, linkedin, portfolio
	PDFPath    string    `gorm:"column:pdf_path" json:"pdf_path"`
	PDFContent string    `json:"pdf_content"` // Extracted text from PDF
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
	SourceID    string     `json:"source_id"`
	SourceURL   string     `json:"source_url"`
	Title       string     `gorm:"not null" json:"title"`
	Company     string     `json:"company"`
	Description string     `json:"description"`
	Skills      string     `json:"skills"`    // JSON array of required skills
	RateMin     float64    `json:"rate_min"`
	RateMax     float64    `json:"rate_max"`
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
	Job         Job               `gorm:"foreignKey:JobID" json:"job,omitempty"`
	ProfileID   uint              `gorm:"index" json:"profile_id"`
	Profile     Profile           `gorm:"foreignKey:ProfileID" json:"-"`
	Proposal    string            `json:"proposal"`
	ProposalPDF string            `json:"proposal_pdf"`
	CoverLetter string            `json:"cover_letter"`
	Status      ApplicationStatus `gorm:"default:draft" json:"status"`
	Notes       string            `json:"notes"`
	AppliedAt   *time.Time        `json:"applied_at"`
	ResponseAt  *time.Time        `json:"response_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// =====================================
// Dig Models - Idea Analysis & Incubation
// =====================================

// DigSessionStatus represents the status of a dig session
type DigSessionStatus string

const (
	DigStatusPending   DigSessionStatus = "pending"
	DigStatusAnalyzing DigSessionStatus = "analyzing"
	DigStatusCompleted DigSessionStatus = "completed"
	DigStatusFailed    DigSessionStatus = "failed"
)

// DigRecommendation represents the final recommendation for an idea
type DigRecommendation string

const (
	DigRecommendProceed DigRecommendation = "proceed"
	DigRecommendPivot   DigRecommendation = "pivot"
	DigRecommendRefine  DigRecommendation = "refine"
	DigRecommendAbandon DigRecommendation = "abandon"
)

// DigPerspective represents analysis perspectives
type DigPerspective string

const (
	DigPerspectiveFirstPrinciples DigPerspective = "first_principles"
	DigPerspectiveDesigner        DigPerspective = "designer"
	DigPerspectiveMarketing       DigPerspective = "marketing"
	DigPerspectiveTechnical       DigPerspective = "technical"
	DigPerspectiveFinancial       DigPerspective = "financial"
)

// DigSession represents an idea analysis session
type DigSession struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	Title           string            `json:"title"`
	RawIdea         string            `gorm:"not null" json:"raw_idea"`
	RefinedIdea     string            `json:"refined_idea"`
	Status          DigSessionStatus  `gorm:"default:pending" json:"status"`
	OverallScore    *float64          `json:"overall_score"`
	Recommendation  DigRecommendation `json:"recommendation"`
	CurrentStage    string            `gorm:"default:first_principles" json:"current_stage"`
	StagesCompleted string            `json:"stages_completed"` // JSON array
	StartedAt       *time.Time        `json:"started_at"`
	CompletedAt     *time.Time        `json:"completed_at"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`

	// Relationships
	Analyses          []DigAnalysis          `gorm:"foreignKey:SessionID" json:"analyses,omitempty"`
	ExecutionPlan     *DigExecutionPlan      `gorm:"foreignKey:SessionID" json:"execution_plan,omitempty"`
	Marketing         *DigMarketing          `gorm:"foreignKey:SessionID" json:"marketing,omitempty"`
	RevenueProjection *DigRevenueProjection  `gorm:"foreignKey:SessionID" json:"revenue_projection,omitempty"`
	Alternatives      []DigAlternative       `gorm:"foreignKey:SessionID" json:"alternatives,omitempty"`
}

// DigAnalysis represents analysis from one perspective
type DigAnalysis struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	SessionID        uint           `gorm:"not null;index" json:"session_id"`
	Perspective      DigPerspective `gorm:"not null" json:"perspective"`
	Summary          string         `json:"summary"`
	Strengths        string         `json:"strengths"`        // JSON array
	Weaknesses       string         `json:"weaknesses"`       // JSON array
	Opportunities    string         `json:"opportunities"`    // JSON array
	Threats          string         `json:"threats"`          // JSON array
	Recommendations  string         `json:"recommendations"`  // JSON array
	Score            *float64       `json:"score"`
	DetailedAnalysis string         `json:"detailed_analysis"` // JSON
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// DigExecutionPlan represents the implementation roadmap
type DigExecutionPlan struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	SessionID        uint      `gorm:"not null;index" json:"session_id"`
	Summary          string    `json:"summary"`
	MVPScope         string    `json:"mvp_scope"`
	FullScope        string    `json:"full_scope"`
	Architecture     string    `json:"architecture"`       // JSON
	TechStack        string    `json:"tech_stack"`         // JSON
	Integrations     string    `json:"integrations"`       // JSON
	Phases           string    `json:"phases"`             // JSON array
	Milestones       string    `json:"milestones"`         // JSON array
	TeamRequirements string    `json:"team_requirements"`  // JSON
	EstimatedCost    string    `json:"estimated_cost"`     // JSON
	LLMPrompt        string    `json:"llm_prompt"`         // Ready-to-use prompt
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DigMarketing represents generated marketing materials
type DigMarketing struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	SessionID            uint      `gorm:"not null;index" json:"session_id"`
	ValueProposition     string    `json:"value_proposition"`
	TargetAudience       string    `json:"target_audience"`        // JSON
	PositioningStatement string    `json:"positioning_statement"`
	Taglines             string    `json:"taglines"`               // JSON array
	ElevatorPitch        string    `json:"elevator_pitch"`
	Headlines            string    `json:"headlines"`              // JSON array
	Descriptions         string    `json:"descriptions"`           // JSON array
	ColorSuggestions     string    `json:"color_suggestions"`      // JSON
	ImageryPrompts       string    `json:"imagery_prompts"`        // JSON array
	GeneratedImages      string    `json:"generated_images"`       // JSON array
	ChannelStrategy      string    `json:"channel_strategy"`       // JSON
	LaunchStrategy       string    `json:"launch_strategy"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// DigRevenueProjection represents financial projections
type DigRevenueProjection struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	SessionID         uint      `gorm:"not null;index" json:"session_id"`
	MarketSize        string    `json:"market_size"`         // JSON: TAM, SAM, SOM
	MarketGrowth      string    `json:"market_growth"`
	Competitors       string    `json:"competitors"`         // JSON array
	PricingModels     string    `json:"pricing_models"`      // JSON array
	RecommendedPrice  string    `json:"recommended_price"`
	PricingRationale  string    `json:"pricing_rationale"`
	Year1Revenue      string    `json:"year1_revenue"`       // JSON monthly
	Year2Revenue      string    `json:"year2_revenue"`       // JSON monthly
	Year3Revenue      string    `json:"year3_revenue"`       // JSON yearly
	KeyMetrics        string    `json:"key_metrics"`         // JSON
	BreakEvenAnalysis string    `json:"break_even_analysis"`
	Assumptions       string    `json:"assumptions"`         // JSON
	Risks             string    `json:"risks"`               // JSON
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// DigAlternative represents a suggested pivot or alternative idea
type DigAlternative struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SessionID       uint      `gorm:"not null;index" json:"session_id"`
	AlternativeIdea string    `gorm:"not null" json:"alternative_idea"`
	Rationale       string    `json:"rationale"`
	Comparison      string    `json:"comparison"`
	ViabilityScore  *float64  `json:"viability_score"`
	EffortLevel     string    `json:"effort_level"` // low, medium, high
	Potential       string    `json:"potential"`    // low, medium, high, very_high
	CreatedAt       time.Time `json:"created_at"`
}

// DigStartRequest represents the request to start a dig session
type DigStartRequest struct {
	Idea     string `json:"idea" validate:"required"`
	UseOwnAI bool   `json:"use_own_ai"` // If true, use user's API key
}

// DigProgressResponse represents the progress of a dig session
type DigProgressResponse struct {
	SessionID       uint             `json:"session_id"`
	Status          DigSessionStatus `json:"status"`
	CurrentStage    string           `json:"current_stage"`
	StagesCompleted []string         `json:"stages_completed"`
	Progress        int              `json:"progress"` // 0-100
	Message         string           `json:"message"`
}
