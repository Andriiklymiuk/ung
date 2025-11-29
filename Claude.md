# UNG - Universal Next-Gen Billing & Tracking

**Complete Project Architecture & Feature Documentation**

*Last Updated: November 29, 2025*
*Version: 1.0.54*
*License: MIT*

---

## Executive Summary

UNG is a comprehensive, open-source billing and time tracking system designed for freelancers. It includes:
- **CLI Tool** (Go) - Core application with offline-first SQLite database
- **REST API** (Go) - Multi-tenant server for cloud sync and team features
- **macOS App** (SwiftUI) - Native application with real-time tracking
- **VSCode Extension** (TypeScript) - IDE integration for seamless workflow
- **Telegram Bot** (Go) - Conversational billing assistant
- **Free forever** - No SaaS subscriptions, data stays local

Key philosophy: No cloud lock-in, no subscriptions, your data stays on your machine.

---

## Project Structure

```
ung/
├── cmd/                          # CLI Commands (42 command files)
│   ├── root.go                   # Root command definition
│   ├── client.go                 # Client management commands
│   ├── contract.go               # Contract management
│   ├── invoice.go                # Invoice generation & management
│   ├── track.go                  # Time tracking commands
│   ├── company.go                # Company settings
│   ├── dashboard.go              # Revenue dashboard
│   ├── export.go                 # Data export/import
│   ├── report.go                 # Financial reports
│   ├── goal.go                   # Income goals
│   ├── recurring.go              # Recurring invoices
│   ├── expense.go                # Expense tracking
│   ├── sync.go                   # Backup/restore functionality
│   ├── security.go               # Database encryption
│   ├── email.go                  # Email integration
│   ├── pomodoro.go               # Pomodoro timer
│   ├── rate.go                   # Rate calculations
│   ├── profit.go                 # Profit analysis
│   ├── template.go               # Invoice templates
│   ├── upgrade.go                # Version upgrades
│   ├── config.go                 # Configuration management
│   ├── database.go               # Database operations
│   ├── cloud.go                  # iCloud sync (future)
│   └── version.go                # Version info
│
├── api/                          # REST API Server (Go)
│   ├── cmd/server/main.go        # API entry point
│   ├── internal/
│   │   ├── controllers/          # 17 controllers for API endpoints
│   │   │   ├── auth_controller.go
│   │   │   ├── invoice_controller.go
│   │   │   ├── client_controller.go
│   │   │   ├── company_controller.go
│   │   │   ├── contract_controller.go
│   │   │   ├── expense_controller.go
│   │   │   ├── tracking_controller.go
│   │   │   ├── dashboard_controller.go
│   │   │   ├── recurring_controller.go
│   │   │   ├── goal_controller.go
│   │   │   ├── rate_controller.go
│   │   │   ├── report_controller.go
│   │   │   ├── pomodoro_controller.go
│   │   │   ├── template_controller.go
│   │   │   ├── search_controller.go
│   │   │   ├── export_controller.go
│   │   │   └── settings_controller.go
│   │   ├── middleware/
│   │   │   ├── auth.go           # JWT authentication
│   │   │   ├── tenant.go         # Multi-tenant isolation
│   │   │   └── subscription.go   # RevenueCat integration
│   │   ├── router/router.go      # Chi route definitions
│   │   ├── services/             # Business logic
│   │   ├── repository/           # Data access layer
│   │   ├── database/             # Database initialization
│   │   ├── models/               # API data models
│   │   └── config/               # API configuration
│   ├── .env.example              # API environment variables
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── telegram/                     # Telegram Bot (Go)
│   ├── cmd/bot/main.go           # Bot entry point
│   ├── internal/
│   │   ├── handlers/             # Command handlers
│   │   │   ├── start.go
│   │   │   ├── invoice.go
│   │   │   ├── client.go
│   │   │   ├── contract.go
│   │   │   ├── tracking.go
│   │   │   ├── report.go
│   │   │   ├── dashboard.go
│   │   │   ├── pomodoro.go
│   │   │   └── help.go
│   │   ├── services/
│   │   │   ├── api_client.go     # API communication
│   │   │   └── session_manager.go # User session state
│   │   ├── models/models.go      # Data models
│   │   └── config/config.go      # Configuration
│   ├── .env.example
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── macos-ung/                    # macOS App (SwiftUI)
│   ├── ung/
│   │   ├── ungApp.swift          # App entry point
│   │   ├── Views/
│   │   │   ├── MainWindow/       # Main app windows
│   │   │   ├── MenuBarView.swift # Menu bar integration
│   │   │   ├── Settings/         # Settings screens
│   │   │   ├── Sheets/           # Modal sheets
│   │   │   ├── Components/       # Reusable components
│   │   │   └── OnboardingView.swift
│   │   ├── Services/
│   │   │   ├── CLIService.swift  # Execute UNG CLI
│   │   │   ├── AppState.swift    # State management
│   │   │   ├── WatchConnectivityService.swift
│   │   │   ├── NotificationService.swift
│   │   │   ├── PDFGenerator.swift
│   │   │   ├── LiveActivityService.swift
│   │   │   ├── SiriShortcutsService.swift
│   │   │   ├── SmartGoalsService.swift
│   │   │   ├── SmartSearchService.swift
│   │   │   ├── PerformanceOptimizer.swift
│   │   │   └── WidgetDataProvider.swift
│   │   ├── DesignSystem/
│   │   │   └── DesignTokens.swift # Design tokens
│   │   ├── Database/              # Local data models
│   │   ├── Helpers/               # Utility functions
│   │   └── Assets.xcassets/
│   ├── ungTests/                 # Unit tests
│   ├── ungWidgets/               # iOS/macOS widgets
│   ├── ung Watch App/            # watchOS app
│   └── ung.xcodeproj/            # Xcode project
│
├── vscode-ung/                   # VSCode Extension (TypeScript)
│   ├── src/
│   │   ├── extension.ts          # Main entry point (167 lines)
│   │   ├── cli/
│   │   │   └── ungCli.ts         # CLI wrapper (321 lines)
│   │   ├── commands/             # Command handlers (6 files)
│   │   │   ├── company.ts
│   │   │   ├── client.ts
│   │   │   ├── contract.ts
│   │   │   ├── invoice.ts
│   │   │   ├── expense.ts
│   │   │   └── tracking.ts
│   │   ├── views/                # Tree view providers (5 files)
│   │   │   ├── invoiceProvider.ts
│   │   │   ├── contractProvider.ts
│   │   │   ├── clientProvider.ts
│   │   │   ├── expenseProvider.ts
│   │   │   └── trackingProvider.ts
│   │   ├── webview/              # Webview panels (2 files)
│   │   │   ├── invoicePanel.ts
│   │   │   └── exportPanel.ts
│   │   └── utils/                # Utilities (3 files)
│   │       ├── config.ts
│   │       ├── formatting.ts
│   │       └── statusBar.ts
│   ├── test/                     # Tests (Mocha + VSCode)
│   ├── package.json              # Extension manifest
│   ├── tsconfig.json
│   ├── .eslintrc.json
│   └── .vscodeignore
│
├── internal/                     # CLI internal packages
│   ├── config/config.go          # Configuration management
│   ├── db/                       # Database operations
│   │   ├── db.go                 # Database initialization & schema
│   │   ├── encryption.go         # AES-256-GCM encryption
│   │   ├── password.go           # Password hashing
│   │   └── keychain.go           # OS keychain integration
│   ├── models/models.go          # Data models (11 types)
│   ├── repository/               # Data access layer
│   │   ├── client.go
│   │   ├── company.go
│   │   ├── contract.go
│   │   ├── invoice.go
│   │   └── tracking.go
│   └── cloud/icloud.go           # iCloud integration (future)
│
├── pkg/                          # Shared packages
│   ├── contract/pdf.go           # Contract PDF generation
│   ├── invoice/pdf.go            # Invoice PDF generation
│   ├── email/smtp.go             # SMTP email sending
│   ├── template/renderer.go      # Template rendering
│   ├── format/format.go          # Date/number formatting
│   └── idgen/idgen.go            # ID generation
│
├── migrations/                   # Database migrations (7 versions)
│   ├── 000001_init_schema.sql
│   ├── 000002_add_contracts.sql
│   ├── 000003_enhance_company_and_invoices.sql
│   ├── 000004_add_contract_pdf_path.sql
│   ├── 000005_add_contract_num.sql
│   ├── 000006_add_expenses.sql
│   └── 000007_add_tracking_deleted_at.sql
│
├── templates/                    # Email/PDF templates
├── ung-docs/                     # Documentation site (Docusaurus)
│   ├── docs/                     # Markdown documentation
│   ├── src/                      # React components
│   └── static/                   # Assets
│
├── scripts/                      # Build/setup scripts
├── .github/workflows/            # CI/CD pipelines
├── Makefile                      # Build targets (30+ targets)
├── main.go                       # CLI entry point
├── go.mod                        # Go dependencies
├── sqlc.yaml                     # SQL code generation config
├── corgi-compose.yml             # Container orchestration
├── lefthook.yml                  # Git hooks
├── .goreleaser.yaml              # Release configuration
├── .env.example                  # Environment variables template
├── .ung.yaml.example             # CLI configuration template
├── CONFIGURATION_GUIDE.md        # Setup instructions
├── README.md                     # Main README
└── LICENSE                       # MIT License

```

---

## Database Models & Relationships

### Core Models

```go
// Company - The user's business entity
type Company struct {
  ID                  uint
  Name                string      // required
  Email               string      // required
  Phone               string
  Address             string
  RegistrationAddress string      // for invoices
  TaxID               string      // for compliance
  BankName            string
  BankAccount         string
  BankSWIFT           string
  LogoPath            string      // for PDF invoices
  CreatedAt           time.Time
  UpdatedAt           time.Time
}

// Client - Customer/client
type Client struct {
  ID        uint
  Name      string      // required
  Email     string      // required
  Address   string
  TaxID     string
  CreatedAt time.Time
  UpdatedAt time.Time
}

// Contract - Work agreement with client
type Contract struct {
  ID           uint          // primary key
  ContractNum  string        // unique: "contract.acme.jan.2025"
  ClientID     uint          // foreign key
  Client       Client        // relationship
  Name         string        // e.g., "Website Development Q1 2025"
  ContractType ContractType  // hourly | fixed_price | retainer
  HourlyRate   *float64      // optional, for hourly contracts
  FixedPrice   *float64      // optional, for fixed price
  Currency     string        // default: USD
  StartDate    time.Time
  EndDate      *time.Time    // optional, open-ended if null
  Active       bool          // default: true
  Notes        string
  PDFPath      string        // generated contract PDF
  CreatedAt    time.Time
  UpdatedAt    time.Time
}

// Invoice - Billing document
type Invoice struct {
  ID          uint          // primary key
  InvoiceNum  string        // unique: "INV-001" format
  CompanyID   uint          // foreign key
  Company     Company       // relationship
  Amount      float64       // total invoice amount
  Currency    string        // default: USD
  Description string
  Status      InvoiceStatus // pending | sent | paid | overdue
  IssuedDate  time.Time     // date invoice created
  DueDate     time.Time     // payment deadline
  PDFPath     string        // generated PDF file
  CreatedAt   time.Time
  UpdatedAt   time.Time
}

// InvoiceLineItem - Individual line on invoice
type InvoiceLineItem struct {
  ID          uint
  InvoiceID   uint        // foreign key
  Invoice     Invoice     // relationship
  ItemName    string      // service name
  Description string
  Quantity    float64     // default: 1
  Rate        float64     // unit price
  Amount      float64     // calculated: quantity * rate
  Discount    float64     // discount amount
  DiscountPct float64     // discount percentage (0-100)
  TaxRate     float64     // tax rate for this item (0-1)
  TaxAmount   float64     // calculated tax
  CreatedAt   time.Time
}

// TrackingSession - Time entry
type TrackingSession struct {
  ID          uint            // primary key
  ClientID    *uint           // optional foreign key
  Client      *Client         // optional relationship
  ContractID  *uint           // optional foreign key
  Contract    *Contract       // optional relationship
  ProjectName string          // task description
  StartTime   time.Time       // required
  EndTime     *time.Time      // optional, null while active
  Duration    *int            // in seconds, null while active
  Hours       *float64        // calculated for display
  Billable    bool            // default: true
  Notes       string          // additional info
  CreatedAt   time.Time
  UpdatedAt   time.Time
  DeletedAt   gorm.DeletedAt  // soft delete (for archive)
}

// Expense - Business expense
type Expense struct {
  ID          uint             // primary key
  Description string           // required
  Amount      float64          // required
  Currency    string           // default: USD
  Category    ExpenseCategory  // software|hardware|travel|meals|office_supplies|utilities|marketing|other
  Date        time.Time        // expense date
  Vendor      string           // who sold it
  ReceiptPath string           // receipt PDF/image
  Notes       string
  CreatedAt   time.Time
  UpdatedAt   time.Time
}

// RecurringInvoice - Template for automatic invoices
type RecurringInvoice struct {
  ID                 uint                  // primary key
  ClientID           uint                  // required foreign key
  Client             Client                // relationship
  ContractID         *uint                 // optional
  Contract           *Contract             // optional
  Amount             float64               // fixed amount
  Currency           string                // default: USD
  Description        string
  Frequency          RecurringFrequency    // weekly|biweekly|monthly|quarterly|yearly
  DayOfMonth         int                   // 1-28 for monthly
  DayOfWeek          int                   // 1-7 for weekly (1=Monday)
  NextGenerationDate time.Time
  LastGeneratedDate  *time.Time
  LastInvoiceID      *uint                 // link to last generated invoice
  Active             bool                  // default: true
  AutoSend           bool                  // send email automatically
  AutoPDF            bool                  // generate PDF automatically
  EmailApp           string                // apple|outlook|gmail
  GeneratedCount     int                   // how many generated
  Notes              string
  CreatedAt          time.Time
  UpdatedAt          time.Time
}

// UserSettings - Key-value store for preferences
type UserSettings struct {
  ID        uint
  Key       string      // unique: "weekly_hours_target", etc.
  Value     string      // JSON string
  CreatedAt time.Time
  UpdatedAt time.Time
}
```

### Enumerations

```go
// ContractType
const (
  ContractTypeHourly     = "hourly"
  ContractTypeFixedPrice = "fixed_price"
  ContractTypeRetainer   = "retainer"
)

// InvoiceStatus
const (
  StatusPending = "pending"
  StatusSent    = "sent"
  StatusPaid    = "paid"
  StatusOverdue = "overdue"
)

// ExpenseCategory
const (
  ExpenseCategorySoftware       = "software"
  ExpenseCategoryHardware       = "hardware"
  ExpenseCategoryTravel         = "travel"
  ExpenseCategoryMeals          = "meals"
  ExpenseCategoryOfficeSupplies = "office_supplies"
  ExpenseCategoryUtilities      = "utilities"
  ExpenseCategoryMarketing      = "marketing"
  ExpenseCategoryOther          = "other"
)

// RecurringFrequency
const (
  FrequencyWeekly    = "weekly"
  FrequencyBiweekly  = "biweekly"
  FrequencyMonthly   = "monthly"
  FrequencyQuarterly = "quarterly"
  FrequencyYearly    = "yearly"
)
```

### Database Schema Location

- **SQLite Database**: `~/.ung/ung.db` (local) or `.ung/ung.db` (project-specific)
- **Migrations**: `/migrations/` directory with 7 versions
- **Schema**: Managed by GORM models, migrations in SQL files
- **Encryption**: Optional AES-256-GCM with PBKDF2 key derivation

---

## CLI Commands

### Configuration Commands

```bash
ung config init              # Initialize local workspace config
ung config init --global     # Initialize global config (~/.ung/)
ung config show              # Display current configuration
ung config path              # Show config file location
ung config migrate           # Migrate from old config format
```

### Company Commands

```bash
ung company add              # Create company profile (interactive)
ung company ls               # List all companies
ung company edit [id]        # Edit company details
```

### Client Commands

```bash
ung client add               # Add new client (interactive)
ung client ls                # List all clients
ung client edit [id]         # Edit client information
ung client delete <id>       # Delete client
```

### Contract Commands

```bash
ung contract add             # Create new contract (interactive)
ung contract ls              # List all contracts
ung contract edit [id]       # Edit contract
ung contract pdf <id>        # Generate contract PDF
ung contract email <id>      # Send contract via email
ung contract delete <id>     # Delete contract
```

### Invoice Commands

```bash
ung invoice create           # Create new invoice (interactive)
ung invoice ls               # List all invoices
ung invoice view <id>        # View invoice details
ung invoice from-time        # Generate from tracked time (interactive)
ung invoice pdf <id>         # Export invoice to PDF
ung invoice email <id>       # Send invoice via email
ung invoice duplicate <id>   # Create invoice from template
ung invoice mark-paid <id>   # Mark invoice as paid
ung invoice mark-sent <id>   # Mark invoice as sent
```

### Time Tracking Commands

```bash
ung track start              # Start new tracking session (interactive)
ung track stop               # Stop current tracking session
ung track now                # Show current/active session
ung track ls                 # List all tracking sessions
ung track log                # Log time manually (interactive)
ung track edit <id>          # Edit tracking session
ung track delete <id>        # Delete tracking session
```

### Expense Commands

```bash
ung expense add              # Log new expense (interactive)
ung expense ls               # List all expenses
ung expense edit <id>        # Edit expense
ung expense delete <id>      # Delete expense
ung expense report           # Generate expense report
```

### Reports & Analysis

```bash
ung dashboard               # Revenue & profit overview
ung report                  # Generate financial reports
ung profit                  # Profit analysis
ung rate analyze            # Analyze hourly rates
ung rate calculate          # Calculate required rate
ung goal                    # Manage income goals
```

### Recurring Invoices

```bash
ung recurring create        # Create recurring invoice template
ung recurring ls            # List recurring invoices
ung recurring generate      # Generate all due invoices
ung recurring pause <id>    # Pause recurring invoice
ung recurring resume <id>   # Resume recurring invoice
```

### Import/Export

```bash
ung export all              # Export all data to JSON
ung export invoices         # Export invoices to CSV
ung export expenses         # Export expenses to CSV
ung export tracking         # Export time tracking to CSV
ung import all              # Import data from JSON
ung import invoices         # Import from CSV
ung import expenses         # Import expenses from CSV
```

### Backup & Sync

```bash
ung sync backup              # Create backup of all data
ung sync backup --output ~/  # Save to custom location
ung sync restore [file]      # Restore from backup
ung sync ls                  # List available backups
```

### Security Commands

```bash
ung security status          # Show encryption status
ung security enable          # Enable database encryption
ung security disable         # Disable encryption
ung security change-password # Change encryption password
ung security save-password   # Save password to OS keychain
ung security forget-password # Remove password from keychain
```

### Other Commands

```bash
ung version                 # Display version information
ung upgrade                 # Upgrade to latest version
ung update                  # Check for updates
ung create                  # Interactive setup wizard
ung pomodoro                # Pomodoro timer
ung template                # Manage invoice templates
ung doctor                  # Health check & diagnostics
ung cloud status            # Check iCloud sync status (future)
ung help                    # Show help information
```

---

## API Endpoints

### Base URL
- **Local**: `http://localhost:8080`
- **Production**: Configured via environment variables

### Authentication Endpoints

```
POST   /api/v1/auth/register        # Register new user
       { "email": "...", "password": "...", "name": "..." }

POST   /api/v1/auth/login           # Login
       { "email": "...", "password": "..." }
       Returns: { "access_token": "...", "refresh_token": "..." }

POST   /api/v1/auth/refresh         # Refresh access token
       { "refresh_token": "..." }

GET    /api/v1/auth/me              # Get user profile
       Headers: { "Authorization": "Bearer {token}" }
```

### Client Endpoints

```
GET    /api/v1/clients              # List all clients
POST   /api/v1/clients              # Create client
GET    /api/v1/clients/{id}         # Get client details
PUT    /api/v1/clients/{id}         # Update client
DELETE /api/v1/clients/{id}         # Delete client
```

### Invoice Endpoints

```
GET    /api/v1/invoices             # List all invoices
POST   /api/v1/invoices             # Create invoice
GET    /api/v1/invoices/{id}        # Get invoice details
PUT    /api/v1/invoices/{id}        # Update invoice
DELETE /api/v1/invoices/{id}        # Delete invoice
PATCH  /api/v1/invoices/{id}/status # Update invoice status
```

### Company Endpoints

```
GET    /api/v1/companies            # List companies
POST   /api/v1/companies            # Create company
GET    /api/v1/companies/{id}       # Get company
PUT    /api/v1/companies/{id}       # Update company
DELETE /api/v1/companies/{id}       # Delete company
```

### Contract Endpoints

```
GET    /api/v1/contracts            # List contracts
POST   /api/v1/contracts            # Create contract
GET    /api/v1/contracts/{id}       # Get contract
PUT    /api/v1/contracts/{id}       # Update contract
DELETE /api/v1/contracts/{id}       # Delete contract
```

### Expense Endpoints

```
GET    /api/v1/expenses             # List expenses
POST   /api/v1/expenses             # Create expense
GET    /api/v1/expenses/{id}        # Get expense
PUT    /api/v1/expenses/{id}        # Update expense
DELETE /api/v1/expenses/{id}        # Delete expense
```

### Time Tracking Endpoints

```
GET    /api/v1/tracking             # List all sessions
POST   /api/v1/tracking             # Create session
GET    /api/v1/tracking/active      # Get active session
POST   /api/v1/tracking/start       # Start new session
GET    /api/v1/tracking/{id}        # Get session
POST   /api/v1/tracking/{id}/stop   # Stop session
PUT    /api/v1/tracking/{id}        # Update session
DELETE /api/v1/tracking/{id}        # Delete session
```

### Dashboard Endpoints

```
GET    /api/v1/dashboard/revenue    # Revenue summary
GET    /api/v1/dashboard/summary    # Overall summary
GET    /api/v1/dashboard/profit     # Profit analysis
```

### Settings Endpoints

```
GET    /api/v1/settings             # Get settings
PUT    /api/v1/settings             # Update settings
GET    /api/v1/settings/working-hours # Get working hours
```

### Rate Endpoints

```
POST   /api/v1/rate/calculate       # Calculate required rate
GET    /api/v1/rate/analyze         # Analyze rates
GET    /api/v1/rate/compare         # Compare rates
```

### Income Goals Endpoints

```
GET    /api/v1/goals                # List goals
POST   /api/v1/goals                # Create goal
GET    /api/v1/goals/status         # Goal status
GET    /api/v1/goals/{id}           # Get goal
PUT    /api/v1/goals/{id}           # Update goal
DELETE /api/v1/goals/{id}           # Delete goal
```

### Recurring Invoice Endpoints

```
GET    /api/v1/recurring            # List recurring invoices
POST   /api/v1/recurring            # Create recurring
POST   /api/v1/recurring/generate   # Generate all due
GET    /api/v1/recurring/{id}       # Get recurring
PUT    /api/v1/recurring/{id}       # Update recurring
DELETE /api/v1/recurring/{id}       # Delete recurring
POST   /api/v1/recurring/{id}/pause    # Pause
POST   /api/v1/recurring/{id}/resume   # Resume
POST   /api/v1/recurring/{id}/generate # Generate single
```

### Report Endpoints

```
GET    /api/v1/reports/weekly       # Weekly report
GET    /api/v1/reports/monthly      # Monthly report
GET    /api/v1/reports/revenue      # Revenue report
GET    /api/v1/reports/clients      # Client report
GET    /api/v1/reports/overdue      # Overdue invoices
GET    /api/v1/reports/unpaid       # Unpaid invoices
```

### Pomodoro Endpoints

```
GET    /api/v1/pomodoro             # List pomodoros
GET    /api/v1/pomodoro/active      # Get active
GET    /api/v1/pomodoro/stats       # Get statistics
POST   /api/v1/pomodoro/start       # Start pomodoro
GET    /api/v1/pomodoro/{id}        # Get pomodoro
POST   /api/v1/pomodoro/{id}/stop   # Stop pomodoro
POST   /api/v1/pomodoro/{id}/complete # Mark complete
DELETE /api/v1/pomodoro/{id}        # Delete pomodoro
```

### Template Endpoints

```
GET    /api/v1/templates            # List templates
POST   /api/v1/templates            # Create template
GET    /api/v1/templates/default    # Get default
GET    /api/v1/templates/{id}       # Get template
PUT    /api/v1/templates/{id}       # Update template
DELETE /api/v1/templates/{id}       # Delete template
POST   /api/v1/templates/{id}/preview # Preview template
```

### Search Endpoints

```
GET    /api/v1/search               # Search all
GET    /api/v1/search/invoices      # Search invoices
GET    /api/v1/search/clients       # Search clients
GET    /api/v1/search/contracts     # Search contracts
```

### Export/Import Endpoints

```
GET    /api/v1/export/all           # Export all JSON
GET    /api/v1/export/invoices/csv  # Export invoices CSV
GET    /api/v1/export/expenses/csv  # Export expenses CSV
GET    /api/v1/export/tracking/csv  # Export tracking CSV
GET    /api/v1/export/clients/csv   # Export clients CSV
POST   /api/v1/import/all           # Import JSON
POST   /api/v1/import/clients/csv   # Import clients CSV
POST   /api/v1/import/expenses/csv  # Import expenses CSV
```

### Subscription Endpoints

```
GET    /api/v1/subscription         # Get subscription status
POST   /api/v1/subscription/verify  # Verify purchase (RevenueCat)
```

### Health Check

```
GET    /health                      # Health status (returns "OK")
```

---

## Key Technologies & Dependencies

### CLI (Go)

**Core Dependencies:**
- `github.com/spf13/cobra` v1.10.1 - Command-line interface framework
- `gorm.io/gorm` v1.31.1 - ORM for database operations
- `gorm.io/driver/sqlite` v1.6.0 - SQLite driver
- `github.com/mattn/go-sqlite3` v1.14.32 - SQLite bindings
- `github.com/charmbracelet/huh` v0.8.0 - Interactive prompts
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework
- `github.com/charmbracelet/lipgloss` v1.1.0 - Terminal styling
- `github.com/jung-kurt/gofpdf` v1.16.2 - PDF generation
- `golang.org/x/crypto` v0.45.0 - Encryption & hashing
- `golang.org/x/term` v0.37.0 - Terminal control
- `gopkg.in/yaml.v3` v3.0.1 - YAML parsing
- `github.com/zalando/go-keyring` v0.2.6 - OS keychain access

**Go Version:** 1.24.0+

### API (Go)

**Additional Dependencies:**
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/go-chi/cors` - CORS middleware

**Database:** SQLite with multi-tenant isolation
**Architecture:** REST API with JWT authentication

### macOS App (Swift)

**Target:** macOS 13.0+
**Framework:** SwiftUI
**Features:**
- Native menu bar integration
- watchOS companion app
- Live activities (real-time notifications)
- Widgets
- Siri Shortcuts integration
- iCloud sync (future)

**Key Services:**
- `CLIService` - Execute UNG CLI commands
- `AppState` - SwiftUI state management
- `NotificationService` - Local notifications
- `WatchConnectivityService` - watchOS sync
- `PDFGenerator` - Generate PDFs
- `SmartGoalsService` - Goal tracking
- `SmartSearchService` - Search functionality
- `PerformanceOptimizer` - Performance monitoring

**Version:** 1.0.54
**Bundle ID:** com.ung.ung

### VSCode Extension (TypeScript)

**Language:** TypeScript 5.0+ (strict mode)
**VSCode Engine:** ^1.80.0
**Node.js:** 18+
**Tools:**
- `npm` - Package manager
- `Mocha` - Testing framework
- `ESLint` - Linting
- `@typescript-eslint` - TypeScript linting

**Features:**
- Tree view providers for data visualization
- Webview panels for detailed views
- Status bar integration
- Command palette commands (30+)
- Keyboard shortcuts
- Configuration system

### Telegram Bot (Go)

**Framework:** `github.com/go-telegram-bot-api/telegram-bot-api`
**Features:**
- Conversational interface with inline keyboards
- Session state management (in-memory)
- API integration with UNG REST API
- JWT authentication

---

## Authentication & Subscription System

### CLI Authentication

**Type:** Local Database-based
**Security:** AES-256-GCM encryption with PBKDF2 key derivation
**Keychain Integration:** Automatic OS keychain support (macOS, Windows, Linux)

```bash
# Enable encryption
ung security enable

# Save password to keychain
ung security save-password

# Change password
ung security change-password
```

### API Authentication

**Type:** JWT (JSON Web Tokens)
**Token Lifecycle:**
- Access Token: 15 minutes expiry
- Refresh Token: 7 days expiry

**Endpoints:**
- `POST /api/v1/auth/register` - Create new user account
- `POST /api/v1/auth/login` - Get tokens
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/me` - Get current user profile

**Multi-Tenant Architecture:**
- Each user has isolated SQLite database
- Databases stored in: `~/.ung/users/{user_id}/ung.db`
- API database: `~/.ung/api.db` (users & auth)

### Subscription Management

**System:** RevenueCat integration (optional)
**Configuration Variables:**
- `REVENUE_CAT_API_KEY` - RevenueCat API key
- `REVENUE_CAT_ENABLED` - Enable/disable subscription checks

**Middleware:** `subscriptionMiddleware` validates subscription status on protected endpoints

**Current Status:** Optional feature, free tier available

**Note:** UNG philosophy is "free forever" - subscriptions are optional for cloud features

---

## Configuration Files & Environment Variables

### CLI Configuration

**File Location:**
- Global: `~/.ung/config.yaml` or `~/.ung/.ung.yaml`
- Project: `.ung/config.yaml` or `.ung.yaml`

**Example Configuration:**
```yaml
# Database and storage paths
database_path: "~/.ung/ung.db"
invoices_dir: "~/.ung/invoices"

# Language code (en, uk, de, fr, es, etc.)
language: "en"

# PDF Configuration
pdf:
  # Color theme (RGB values 0-255)
  primary_color:
    r: 232
    g: 119
    b: 34
  secondary_color:
    r: 80
    g: 80
    b: 80
  text_color:
    r: 60
    g: 60
    b: 60

  # Watermark settings
  show_watermark: true
  show_logo: true
  show_qr_code: false
  show_page_number: true
  show_tax_breakdown: false

  # Tax/VAT settings
  tax_rate: 0.0
  tax_label: "VAT"
  tax_inclusive: false

  # Labels
  subtotal_label: "Subtotal"
  discount_label: "Discount"
  tax_amount_label: "VAT"
  balance_due_label: "Balance Due"
  paid_label: "PAID"
  draft_label: "DRAFT"
  overdue_label: "OVERDUE"

# Invoice configuration
invoice:
  terms: "Please make the payment by the due date."
  payment_note: "Payment is due within the specified term."
  notes_label: "Notes"
  terms_label: "Terms & Conditions"
  invoice_label: "INVOICE"
  # ... more labels
```

### API Environment Variables

**File:** `api/.env` (required for API server)

```bash
# Server Configuration
PORT=8080
ENV=development

# Database Paths (SQLite)
API_DATABASE_PATH=/home/user/.ung/api.db
USER_DATA_DIR=/home/user/.ung/users

# Security - MUST CHANGE IN PRODUCTION
JWT_SECRET=your-super-secret-jwt-key-minimum-32-chars-recommended

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000,https://ung.app

# Email/SMTP Configuration (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@ung.app
SMTP_FROM_NAME=UNG Billing
SMTP_USE_TLS=true

# Scheduler Configuration
ENABLE_SCHEDULER=true
SCHEDULER_INVOICE_REMINDERS=true
SCHEDULER_OVERDUE_NOTIFICATIONS=true
SCHEDULER_CONTRACT_REMINDERS=true
SCHEDULER_WEEKLY_SUMMARY=true
SCHEDULER_MONTHLY_REPORTS=true

# RevenueCat Integration (optional)
REVENUE_CAT_API_KEY=your-api-key
REVENUE_CAT_ENABLED=false
```

### Telegram Bot Environment Variables

**File:** `telegram/.env` (required for bot)

```bash
# Bot Configuration
TELEGRAM_BOT_TOKEN=your-telegram-bot-token-from-botfather

# API Configuration
UNG_API_URL=http://localhost:8080
WEB_APP_URL=https://ung.app

# Security
JWT_SECRET=your-super-secret-jwt-key

# Debug mode
DEBUG=false
```

### VSCode Extension Configuration

**File:** VSCode `settings.json`

```json
{
  "ung.cliPath": "ung",
  "ung.defaultCurrency": "USD",
  "ung.autoRefresh": true,
  "ung.dateFormat": "YYYY-MM-DD",
  "ung.showStatusBar": true
}
```

---

## Build/Test/Run Instructions

### CLI Build & Test

```bash
# Build binary
make build

# Install to GOPATH/bin
make install

# Run tests
make test

# Development build with race detector
make dev

# Cross-compile for all platforms
make build-all

# Format code
make fmt

# Run linter
make lint
```

### API Build & Run

```bash
# Development
cd api
cp .env.example .env
go mod download
go run cmd/server/main.go

# Using Docker
docker-compose up --build

# Tests
go test -v ./...
```

### macOS App Build

```bash
# Build in Release configuration
make macosBuild

# Create archive for distribution
make macosArchive

# Export for App Store
make macosExport

# Run tests
make testMac
```

### VSCode Extension Build

```bash
# Compile TypeScript
cd vscode-ung
npm install
npm run compile

# Watch mode
npm run watch

# Run tests
npm test

# Lint code
npm run lint

# Package for distribution
npm run vsce package
```

### Telegram Bot Build & Run

```bash
# Development
cd telegram
cp .env.example .env
go mod download
go run cmd/bot/main.go

# Using Docker
docker-compose up --build
```

---

## Release & Versioning

### Version Management

**Source of Truth:** `cmd/version.go`
**Current Version:** 1.0.54
**Semantic Versioning:** MAJOR.MINOR.PATCH

```bash
# Auto-bump patch version
make release

# Bump to specific version
make release v=1.2.3

# Auto-bump minor
make increaseVersion v=1.1.0

# Manual bump
make increaseVersion
```

### Release Process

1. Version is bumped in:
   - `cmd/version.go` (CLI)
   - `vscode-ung/package.json` (VSCode)
   - `macos-ung/ung.xcodeproj/project.pbxproj` (macOS)

2. Changes committed and tagged with git

3. GitHub Actions workflows triggered

### Cross-Platform Support

- **CLI:** Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- **macOS App:** macOS 13.0+
- **VSCode:** All platforms supporting VSCode
- **Telegram Bot:** Any platform running Go

---

## Installation & Setup

### CLI Installation

```bash
# Using Homebrew (macOS)
brew install andriiklymiuk/tools/ung

# From source
git clone https://github.com/Andriiklymiuk/ung.git
cd ung
make build
./ung

# Go install
go install github.com/Andriiklymiuk/ung@latest
```

### Initial Setup

```bash
# Initialize configuration
ung config init --global

# Add company
ung company add --name "Your Name" --email "you@email.com"

# Add client
ung client add --name "Acme Corp" --email "billing@acme.com"

# Create contract
ung contract add --client 1 --name "Development" --type hourly --rate 100

# Start tracking time
ung track start --contract 1

# Create invoice
ung invoice create --client "Acme Corp"
```

### API Server Setup

```bash
# Navigate to API directory
cd api

# Create environment file
cp .env.example .env

# Edit configuration
nano .env

# Run database migrations (automatic on first run)
go run cmd/server/main.go

# Server runs on http://localhost:8080
```

### Telegram Bot Setup

```bash
# Create bot with @BotFather on Telegram
# Get your bot token

# Setup bot
cd telegram
cp .env.example .env

# Edit with bot token
nano .env

# Run bot
go run cmd/bot/main.go
```

### VSCode Extension Setup

```bash
# Install CLI first
brew install andriiklymiuk/tools/ung

# Install extension from VSCode Marketplace
# Or build from source

cd vscode-ung
npm install
npm run compile

# Package extension
vsce package

# Install locally
code --install-extension ung-1.0.54.vsix
```

---

## Project Statistics

- **Total Lines of Code:** ~50,000+ (across all components)
  - CLI: ~20,000 lines
  - API: ~15,000 lines
  - macOS App: ~10,000 lines
  - VSCode Extension: ~2,500 lines
  - Telegram Bot: ~5,000 lines

- **Database Models:** 11 main entities
- **API Endpoints:** 60+ endpoints
- **CLI Commands:** 40+ commands
- **Migrations:** 7 database versions
- **Test Files:** 40+ test files

---

## Key Features Summary

### Universal

- Multi-platform support (CLI, macOS, Windows, Linux, Telegram)
- Offline-first design with local SQLite database
- End-to-end encryption support
- Backup & restore functionality
- Data import/export (JSON, CSV)

### Billing

- Invoice generation with customizable templates
- Multiple contract types (hourly, fixed-price, retainer)
- Recurring invoices with automation
- Tax/VAT support
- Multi-currency support (USD, EUR, GBP, etc.)
- PDF generation with custom branding

### Time Tracking

- Real-time timer with live activity updates
- Manual time logging
- Billable/non-billable indicators
- Pomodoro timer integration
- Time tracking by project/client/contract

### Financial Analysis

- Revenue dashboards
- Profit calculations
- Expense tracking by category
- Income goal management
- Rate analysis and comparison
- Financial reports (weekly, monthly)

### Client Management

- Client database with contact info
- Contract tracking
- Payment status monitoring
- Overdue invoice alerts

### Security & Privacy

- Local-first data storage
- AES-256-GCM encryption at rest
- OS keychain integration
- No cloud lock-in
- No subscriptions required

---

## Future Development

**Planned Features:**
- iCloud sync for macOS
- Team collaboration
- Advanced AI insights
- QuickBooks/Xero integration
- Mobile apps (iOS)
- Webhook mode for Telegram bot
- Redis session persistence
- Advanced reporting & analytics

**Architecture Considerations:**
- Modular design for easy feature addition
- Clean separation of concerns
- Comprehensive test coverage
- Automated CI/CD pipelines
- Docker containerization support

---

## Contributing

UNG is open source (MIT License). The project structure supports:
- Adding new CLI commands
- Creating new API endpoints
- Extending macOS app features
- Building new VSCode extension features
- Enhancing Telegram bot capabilities

Each component is modular and well-documented for easy contribution.

---

## Resources

- **Main Repository:** https://github.com/Andriiklymiuk/ung
- **Documentation:** https://andriiklymiuk.github.io/ung
- **Issues:** https://github.com/Andriiklymiuk/ung/issues
- **CLI Installation:** https://brew.sh/ (via Homebrew)
- **VSCode Marketplace:** Search "UNG"

---

## License

MIT License - Free forever, no restrictions

---

*This documentation was auto-generated from codebase analysis*
*Project Architecture Version: 1.0.54*
*Last Updated: November 29, 2025*
