# UNG Go REST API

**Multi-tenant RESTful API for UNG - Your Next Gig, Simplified**

This document outlines the architecture and implementation of the UNG Go REST API, designed to provide a NestJS-like structure while adhering to Go idioms and using the Chi router for standard library compatibility.

## Overview

The UNG Go API provides a RESTful interface to the UNG billing system, supporting:
- Multi-tenant architecture (each user has their own SQLite database)
- JWT-based authentication
- All UNG CLI operations via HTTP endpoints
- Recurring invoice automation
- Gmail integration for email sending
- Webhook support
- Real-time updates via WebSockets
- Rate limiting and security

## Architecture

### NestJS-Inspired Structure

```
api/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration loading
│   │   └── env.go                  # Environment variables
│   ├── middleware/
│   │   ├── auth.go                 # JWT authentication
│   │   ├── cors.go                 # CORS handling
│   │   ├── logger.go               # Request logging
│   │   ├── ratelimit.go            # Rate limiting
│   │   └── tenant.go               # Multi-tenant context
│   ├── controllers/
│   │   ├── auth_controller.go      # Auth endpoints
│   │   ├── company_controller.go   # Company CRUD
│   │   ├── client_controller.go    # Client CRUD
│   │   ├── contract_controller.go  # Contract CRUD
│   │   ├── invoice_controller.go   # Invoice CRUD
│   │   ├── track_controller.go     # Time tracking
│   │   └── webhook_controller.go   # Webhook handlers
│   ├── services/
│   │   ├── auth_service.go         # Authentication logic
│   │   ├── company_service.go      # Business logic
│   │   ├── client_service.go
│   │   ├── contract_service.go
│   │   ├── invoice_service.go
│   │   ├── track_service.go
│   │   ├── email_service.go        # Gmail integration
│   │   ├── pdf_service.go          # PDF generation
│   │   └── scheduler_service.go    # Recurring invoices
│   ├── repository/
│   │   ├── user_repository.go      # User management
│   │   ├── company_repository.go   # Reuse from CLI
│   │   ├── client_repository.go
│   │   ├── contract_repository.go
│   │   ├── invoice_repository.go
│   │   └── track_repository.go
│   ├── models/
│   │   ├── user.go                 # API-specific models
│   │   ├── auth.go                 # JWT claims, tokens
│   │   └── response.go             # Standard responses
│   ├── database/
│   │   ├── connection.go           # DB connection pool
│   │   ├── tenant.go               # Per-user DB management
│   │   └── migrations.go           # Migration runner
│   └── router/
│       ├── router.go               # Chi router setup
│       └── routes.go               # Route definitions
├── migrations/                      # Same as CLI (/migrations/)
│   ├── 000001_initial_schema.up.sql
│   ├── 000002_add_invoice_fields.up.sql
│   └── ...
├── pkg/
│   └── utils/
│       ├── jwt.go                  # JWT utilities
│       ├── password.go             # Password hashing
│       └── validator.go            # Input validation
├── scripts/
│   ├── seed.sql                    # Sample data
│   └── setup.sh                    # Initial setup
├── .env.example
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## Key Technologies

- **Chi Router** - Lightweight, idiomatic router compatible with `net/http`
- **GORM** - ORM for database operations (reuse CLI models)
- **JWT-Go** - JSON Web Token authentication
- **Validator** - Request validation
- **Cron** - Scheduled tasks (recurring invoices)
- **Gmail API** - Email sending via Gmail
- **WebSockets** - Real-time updates (optional)
- **SQLite** - Per-user databases

## Multi-Tenant Architecture

### User Database Isolation

Each user has their own SQLite database file:

```
~/.ung/
├── users/
│   ├── user_1/
│   │   ├── ung.db                  # User 1's database
│   │   ├── invoices/               # User 1's PDFs
│   │   └── contracts/
│   ├── user_2/
│   │   ├── ung.db                  # User 2's database
│   │   ├── invoices/
│   │   └── contracts/
│   └── ...
└── api.db                           # API metadata (users, auth)
```

### Database Models

**API Database (api.db):**

```go
// internal/models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

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

type RefreshToken struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    Token     string    `gorm:"uniqueIndex;not null" json:"token"`
    ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
    CreatedAt time.Time `json:"created_at"`
}

type APIKey struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    Key       string    `gorm:"uniqueIndex;not null" json:"key"`
    Name      string    `gorm:"not null" json:"name"`
    LastUsed  *time.Time `json:"last_used"`
    ExpiresAt *time.Time `json:"expires_at"`
    Active    bool      `gorm:"default:true" json:"active"`
    CreatedAt time.Time `json:"created_at"`
}
```

**User Database (user_*/ung.db):**
- Reuses all models from CLI: Company, Client, Contract, Invoice, TimeEntry, InvoiceLineItem
- Same migration files as CLI

### Tenant Context Middleware

```go
// internal/middleware/tenant.go
package middleware

import (
    "context"
    "net/http"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
)

type contextKey string

const (
    UserContextKey   contextKey = "user"
    TenantDBKey      contextKey = "tenantDB"
)

func TenantMiddleware(apiDB *gorm.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // User already set by AuthMiddleware
            user, ok := r.Context().Value(UserContextKey).(*models.User)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Open user's specific database
            tenantDB, err := gorm.Open(sqlite.Open(user.DBPath), &gorm.Config{})
            if err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }

            // Add tenant DB to context
            ctx := context.WithValue(r.Context(), TenantDBKey, tenantDB)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Helper to get tenant DB from context
func GetTenantDB(r *http.Request) *gorm.DB {
    return r.Context().Value(TenantDBKey).(*gorm.DB)
}

// Helper to get user from context
func GetUser(r *http.Request) *models.User {
    return r.Context().Value(UserContextKey).(*models.User)
}
```

## Authentication

### JWT Implementation

```go
// pkg/utils/jwt.go
package utils

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, email string, secret string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "ung-api",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uint, email string, secret string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "ung-api",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, jwt.ErrSignatureInvalid
}
```

### Auth Middleware

```go
// internal/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
    "strings"
    "ung-api/pkg/utils"
    "ung-api/internal/repository"
)

func AuthMiddleware(userRepo *repository.UserRepository, jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
                return
            }

            // Validate JWT
            claims, err := utils.ValidateToken(tokenString, jwtSecret)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Get user from database
            user, err := userRepo.GetByID(claims.UserID)
            if err != nil {
                http.Error(w, "User not found", http.StatusUnauthorized)
                return
            }

            if !user.Active {
                http.Error(w, "Account disabled", http.StatusForbidden)
                return
            }

            // Add user to context
            ctx := context.WithValue(r.Context(), UserContextKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Controllers (NestJS-style)

### Auth Controller

```go
// internal/controllers/auth_controller.go
package controllers

import (
    "encoding/json"
    "net/http"
    "ung-api/internal/services"
    "ung-api/internal/models"
)

type AuthController struct {
    authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
    return &AuthController{authService: authService}
}

// POST /api/v1/auth/register
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"`
        Name     string `json:"name" validate:"required"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, accessToken, refreshToken, err := c.authService.Register(req.Email, req.Password, req.Name)
    if err != nil {
        respondError(w, err.Error(), http.StatusBadRequest)
        return
    }

    respondJSON(w, map[string]interface{}{
        "user":          user,
        "access_token":  accessToken,
        "refresh_token": refreshToken,
    }, http.StatusCreated)
}

// POST /api/v1/auth/login
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, accessToken, refreshToken, err := c.authService.Login(req.Email, req.Password)
    if err != nil {
        respondError(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    respondJSON(w, map[string]interface{}{
        "user":          user,
        "access_token":  accessToken,
        "refresh_token": refreshToken,
    }, http.StatusOK)
}

// POST /api/v1/auth/refresh
func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token" validate:"required"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    accessToken, refreshToken, err := c.authService.RefreshToken(req.RefreshToken)
    if err != nil {
        respondError(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }

    respondJSON(w, map[string]interface{}{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
    }, http.StatusOK)
}

// GET /api/v1/auth/me
func (c *AuthController) GetProfile(w http.ResponseWriter, r *http.Request) {
    user := middleware.GetUser(r)
    respondJSON(w, user, http.StatusOK)
}
```

### Invoice Controller

```go
// internal/controllers/invoice_controller.go
package controllers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "github.com/go-chi/chi/v5"
    "ung-api/internal/services"
    "ung-api/internal/middleware"
)

type InvoiceController struct {
    invoiceService *services.InvoiceService
}

func NewInvoiceController(invoiceService *services.InvoiceService) *InvoiceController {
    return &InvoiceController{invoiceService: invoiceService}
}

// GET /api/v1/invoices
func (c *InvoiceController) List(w http.ResponseWriter, r *http.Request) {
    tenantDB := middleware.GetTenantDB(r)

    // Parse query params for filtering
    status := r.URL.Query().Get("status")
    clientID := r.URL.Query().Get("client_id")

    invoices, err := c.invoiceService.List(tenantDB, status, clientID)
    if err != nil {
        respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    respondJSON(w, invoices, http.StatusOK)
}

// GET /api/v1/invoices/:id
func (c *InvoiceController) Get(w http.ResponseWriter, r *http.Request) {
    tenantDB := middleware.GetTenantDB(r)
    id, _ := strconv.Atoi(chi.URLParam(r, "id"))

    invoice, err := c.invoiceService.GetByID(tenantDB, uint(id))
    if err != nil {
        respondError(w, "Invoice not found", http.StatusNotFound)
        return
    }

    respondJSON(w, invoice, http.StatusOK)
}

// POST /api/v1/invoices
func (c *InvoiceController) Create(w http.ResponseWriter, r *http.Request) {
    tenantDB := middleware.GetTenantDB(r)

    var req struct {
        ClientID    uint    `json:"client_id" validate:"required"`
        Amount      float64 `json:"amount" validate:"required,gt=0"`
        Currency    string  `json:"currency" validate:"required"`
        DueDate     string  `json:"due_date" validate:"required"`
        Description string  `json:"description"`
        LineItems   []struct {
            ItemName    string  `json:"item_name" validate:"required"`
            Description string  `json:"description"`
            Quantity    float64 `json:"quantity" validate:"required,gt=0"`
            Rate        float64 `json:"rate" validate:"required,gt=0"`
        } `json:"line_items" validate:"required,min=1"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    invoice, err := c.invoiceService.Create(tenantDB, &req)
    if err != nil {
        respondError(w, err.Error(), http.StatusBadRequest)
        return
    }

    respondJSON(w, invoice, http.StatusCreated)
}

// PUT /api/v1/invoices/:id
func (c *InvoiceController) Update(w http.ResponseWriter, r *http.Request) {
    tenantDB := middleware.GetTenantDB(r)
    id, _ := strconv.Atoi(chi.URLParam(r, "id"))

    var updates map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
        respondError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    invoice, err := c.invoiceService.Update(tenantDB, uint(id), updates)
    if err != nil {
        respondError(w, err.Error(), http.StatusBadRequest)
        return
    }

    respondJSON(w, invoice, http.StatusOK)
}

// DELETE /api/v1/invoices/:id
func (c *InvoiceController) Delete(w http.ResponseWriter, r *http.Request) {
    tenantDB := middleware.GetTenantDB(r)
    id, _ := strconv.Atoi(chi.URLParam(r, "id"))

    if err := c.invoiceService.Delete(tenantDB, uint(id)); err != nil {
        respondError(w, err.Error(), http.StatusBadRequest)
        return
    }

    respondJSON(w, map[string]string{"message": "Invoice deleted"}, http.StatusOK)
}

// POST /api/v1/invoices/:id/pdf
func (c *InvoiceController) GeneratePDF(w http.ResponseWriter, r *http.Request) {
    tenantDB := middleware.GetTenantDB(r)
    user := middleware.GetUser(r)
    id, _ := strconv.Atoi(chi.URLParam(r, "id"))

    pdfPath, err := c.invoiceService.GeneratePDF(tenantDB, user, uint(id))
    if err != nil {
        respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    respondJSON(w, map[string]string{"pdf_path": pdfPath}, http.StatusOK)
}

// POST /api/v1/invoices/:id/send
func (c *InvoiceController) SendEmail(w http.ResponseWriter, r *http.Request) {
    tenantDB := middleware.GetTenantDB(r)
    user := middleware.GetUser(r)
    id, _ := strconv.Atoi(chi.URLParam(r, "id"))

    if err := c.invoiceService.SendEmail(tenantDB, user, uint(id)); err != nil {
        respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    respondJSON(w, map[string]string{"message": "Invoice sent"}, http.StatusOK)
}
```

## Services (Business Logic)

### Invoice Service

```go
// internal/services/invoice_service.go
package services

import (
    "gorm.io/gorm"
    "ung-api/internal/repository"
    "ung-api/internal/models"
    cliModels "ung/internal/models" // Reuse CLI models
    "ung/pkg/invoice" // Reuse CLI PDF generation
)

type InvoiceService struct {
    invoiceRepo  *repository.InvoiceRepository
    clientRepo   *repository.ClientRepository
    companyRepo  *repository.CompanyRepository
    emailService *EmailService
}

func NewInvoiceService(
    invoiceRepo *repository.InvoiceRepository,
    clientRepo *repository.ClientRepository,
    companyRepo *repository.CompanyRepository,
    emailService *EmailService,
) *InvoiceService {
    return &InvoiceService{
        invoiceRepo:  invoiceRepo,
        clientRepo:   clientRepo,
        companyRepo:  companyRepo,
        emailService: emailService,
    }
}

func (s *InvoiceService) List(db *gorm.DB, status, clientID string) ([]cliModels.Invoice, error) {
    repo := repository.NewInvoiceRepository(db)

    if status != "" {
        return repo.ListByStatus(status)
    }

    if clientID != "" {
        id, _ := strconv.Atoi(clientID)
        return repo.ListByClientID(uint(id))
    }

    return repo.List()
}

func (s *InvoiceService) GetByID(db *gorm.DB, id uint) (*cliModels.Invoice, error) {
    repo := repository.NewInvoiceRepository(db)
    return repo.GetByID(id)
}

func (s *InvoiceService) Create(db *gorm.DB, req interface{}) (*cliModels.Invoice, error) {
    repo := repository.NewInvoiceRepository(db)

    // Convert request to invoice model
    // Generate invoice number
    // Create invoice with line items
    // Return created invoice

    return repo.Create(invoice)
}

func (s *InvoiceService) GeneratePDF(db *gorm.DB, user *models.User, id uint) (string, error) {
    repo := repository.NewInvoiceRepository(db)
    companyRepo := repository.NewCompanyRepository(db)

    inv, err := repo.GetByID(id)
    if err != nil {
        return "", err
    }

    companies, err := companyRepo.List()
    if err != nil || len(companies) == 0 {
        return "", errors.New("no company configured")
    }

    // Reuse CLI PDF generation
    pdfPath, err := invoice.GeneratePDF(*inv, companies[0], inv.Client, inv.LineItems)
    if err != nil {
        return "", err
    }

    // Update invoice with PDF path
    inv.PDFPath = pdfPath
    repo.Update(inv)

    return pdfPath, nil
}

func (s *InvoiceService) SendEmail(db *gorm.DB, user *models.User, id uint) error {
    // Generate PDF if not exists
    pdfPath, err := s.GeneratePDF(db, user, id)
    if err != nil {
        return err
    }

    // Get invoice details
    repo := repository.NewInvoiceRepository(db)
    inv, err := repo.GetByID(id)
    if err != nil {
        return err
    }

    // Send via Gmail API
    return s.emailService.SendInvoice(user, inv, pdfPath)
}
```

### Email Service (Gmail Integration)

```go
// internal/services/email_service.go
package services

import (
    "context"
    "encoding/base64"
    "fmt"
    "io/ioutil"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/gmail/v1"
    "google.golang.org/api/option"
    "ung-api/internal/models"
    cliModels "ung/internal/models"
)

type EmailService struct {
    clientID     string
    clientSecret string
}

func NewEmailService(clientID, clientSecret string) *EmailService {
    return &EmailService{
        clientID:     clientID,
        clientSecret: clientSecret,
    }
}

func (s *EmailService) SendInvoice(user *models.User, invoice *cliModels.Invoice, pdfPath string) error {
    // Get Gmail token from user
    if user.GmailToken == nil {
        return errors.New("Gmail not connected. Please authorize access.")
    }

    // Decrypt token
    token := s.decryptToken(*user.GmailToken)

    // Create Gmail service
    ctx := context.Background()
    config := &oauth2.Config{
        ClientID:     s.clientID,
        ClientSecret: s.clientSecret,
        Endpoint:     google.Endpoint,
        Scopes:       []string{gmail.GmailSendScope},
    }

    client := config.Client(ctx, token)
    srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
        return err
    }

    // Read PDF file
    pdfData, err := ioutil.ReadFile(pdfPath)
    if err != nil {
        return err
    }

    // Create email message
    subject := fmt.Sprintf("Invoice %s from %s", invoice.InvoiceNum, invoice.Company.Name)
    body := fmt.Sprintf(`
Dear %s,

Please find attached invoice %s for the amount of %s %s.

Due Date: %s

Thank you for your business!

Best regards,
%s
`, invoice.Client.Name, invoice.InvoiceNum, invoice.Amount, invoice.Currency,
        invoice.DueDate.Format("January 2, 2006"), invoice.Company.Name)

    // Build multipart message with attachment
    message := s.buildEmailMessage(
        invoice.Client.Email,
        subject,
        body,
        pdfData,
        fmt.Sprintf("invoice_%s.pdf", invoice.InvoiceNum),
    )

    // Send email
    _, err = srv.Users.Messages.Send("me", &gmail.Message{
        Raw: message,
    }).Do()

    return err
}

func (s *EmailService) buildEmailMessage(to, subject, body string, attachment []byte, filename string) string {
    // Build RFC 2822 email with MIME multipart
    boundary := "boundary123"

    message := fmt.Sprintf(`From: me
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="%s"

--%s
Content-Type: text/plain; charset="UTF-8"

%s

--%s
Content-Type: application/pdf; name="%s"
Content-Disposition: attachment; filename="%s"
Content-Transfer-Encoding: base64

%s
--%s--
`, to, subject, boundary, boundary, body, boundary, filename, filename,
        base64.StdEncoding.EncodeToString(attachment), boundary)

    return base64.URLEncoding.EncodeToString([]byte(message))
}

// OAuth flow endpoints
func (s *EmailService) GetAuthURL(userID uint) string {
    config := &oauth2.Config{
        ClientID:     s.clientID,
        ClientSecret: s.clientSecret,
        Endpoint:     google.Endpoint,
        RedirectURL:  "http://localhost:8080/api/v1/gmail/callback",
        Scopes:       []string{gmail.GmailSendScope},
    }

    return config.AuthCodeURL(fmt.Sprintf("user_%d", userID), oauth2.AccessTypeOffline)
}

func (s *EmailService) HandleCallback(code string) (*oauth2.Token, error) {
    config := &oauth2.Config{
        ClientID:     s.clientID,
        ClientSecret: s.clientSecret,
        Endpoint:     google.Endpoint,
        RedirectURL:  "http://localhost:8080/api/v1/gmail/callback",
        Scopes:       []string{gmail.GmailSendScope},
    }

    return config.Exchange(context.Background(), code)
}
```

### Scheduler Service (Recurring Invoices)

```go
// internal/services/scheduler_service.go
package services

import (
    "log"
    "time"
    "github.com/robfig/cron/v3"
    "gorm.io/gorm"
    "ung-api/internal/repository"
)

type SchedulerService struct {
    cron         *cron.Cron
    apiDB        *gorm.DB
    emailService *EmailService
}

func NewSchedulerService(apiDB *gorm.DB, emailService *EmailService) *SchedulerService {
    return &SchedulerService{
        cron:         cron.New(),
        apiDB:        apiDB,
        emailService: emailService,
    }
}

func (s *SchedulerService) Start() {
    // Check for recurring invoices daily at 9 AM
    s.cron.AddFunc("0 9 * * *", s.processRecurringInvoices)

    // Check for overdue invoices daily at 10 AM
    s.cron.AddFunc("0 10 * * *", s.sendOverdueReminders)

    s.cron.Start()
    log.Println("Scheduler service started")
}

func (s *SchedulerService) Stop() {
    s.cron.Stop()
}

func (s *SchedulerService) processRecurringInvoices() {
    // Get all active users
    userRepo := repository.NewUserRepository(s.apiDB)
    users, err := userRepo.ListActive()
    if err != nil {
        log.Printf("Error getting users: %v", err)
        return
    }

    for _, user := range users {
        // Open user's database
        userDB, err := gorm.Open(sqlite.Open(user.DBPath), &gorm.Config{})
        if err != nil {
            log.Printf("Error opening DB for user %d: %v", user.ID, err)
            continue
        }

        // Find contracts with recurring invoices
        contractRepo := repository.NewContractRepository(userDB)
        contracts, err := contractRepo.ListRecurring()
        if err != nil {
            log.Printf("Error getting recurring contracts: %v", err)
            continue
        }

        for _, contract := range contracts {
            // Check if invoice is due
            if s.isInvoiceDue(contract) {
                // Generate and send invoice
                invoice := s.createInvoiceFromContract(userDB, contract)
                s.emailService.SendInvoice(&user, invoice, invoice.PDFPath)
                log.Printf("Sent recurring invoice for user %d, contract %d", user.ID, contract.ID)
            }
        }
    }
}

func (s *SchedulerService) sendOverdueReminders() {
    // Similar pattern - iterate users and send reminders for overdue invoices
}
```

## Router Configuration

```go
// internal/router/router.go
package router

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    "ung-api/internal/controllers"
    apiMiddleware "ung-api/internal/middleware"
)

type Router struct {
    chi              *chi.Mux
    authController   *controllers.AuthController
    companyController *controllers.CompanyController
    clientController *controllers.ClientController
    contractController *controllers.ContractController
    invoiceController *controllers.InvoiceController
    trackController  *controllers.TrackController
}

func NewRouter(
    authController *controllers.AuthController,
    companyController *controllers.CompanyController,
    clientController *controllers.ClientController,
    contractController *controllers.ContractController,
    invoiceController *controllers.InvoiceController,
    trackController *controllers.TrackController,
    authMiddleware func(http.Handler) http.Handler,
    tenantMiddleware func(http.Handler) http.Handler,
) *Router {
    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))

    // CORS
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"https://*", "http://*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    router := &Router{
        chi:                r,
        authController:     authController,
        companyController:  companyController,
        clientController:   clientController,
        contractController: contractController,
        invoiceController:  invoiceController,
        trackController:    trackController,
    }

    router.setupRoutes(authMiddleware, tenantMiddleware)

    return router
}

func (rt *Router) setupRoutes(authMiddleware, tenantMiddleware func(http.Handler) http.Handler) {
    r := rt.chi

    // Health check
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    // API v1
    r.Route("/api/v1", func(r chi.Router) {
        // Public routes
        r.Group(func(r chi.Router) {
            r.Post("/auth/register", rt.authController.Register)
            r.Post("/auth/login", rt.authController.Login)
            r.Post("/auth/refresh", rt.authController.RefreshToken)
        })

        // Protected routes
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware)

            // Auth
            r.Get("/auth/me", rt.authController.GetProfile)

            // Gmail integration
            r.Get("/gmail/auth", rt.authController.GetGmailAuthURL)
            r.Get("/gmail/callback", rt.authController.HandleGmailCallback)

            // Tenant-specific routes
            r.Group(func(r chi.Router) {
                r.Use(tenantMiddleware)

                // Companies
                r.Route("/companies", func(r chi.Router) {
                    r.Get("/", rt.companyController.List)
                    r.Post("/", rt.companyController.Create)
                    r.Get("/{id}", rt.companyController.Get)
                    r.Put("/{id}", rt.companyController.Update)
                    r.Delete("/{id}", rt.companyController.Delete)
                })

                // Clients
                r.Route("/clients", func(r chi.Router) {
                    r.Get("/", rt.clientController.List)
                    r.Post("/", rt.clientController.Create)
                    r.Get("/{id}", rt.clientController.Get)
                    r.Put("/{id}", rt.clientController.Update)
                    r.Delete("/{id}", rt.clientController.Delete)
                })

                // Contracts
                r.Route("/contracts", func(r chi.Router) {
                    r.Get("/", rt.contractController.List)
                    r.Post("/", rt.contractController.Create)
                    r.Get("/{id}", rt.contractController.Get)
                    r.Put("/{id}", rt.contractController.Update)
                    r.Delete("/{id}", rt.contractController.Delete)
                    r.Post("/{id}/pdf", rt.contractController.GeneratePDF)
                })

                // Invoices
                r.Route("/invoices", func(r chi.Router) {
                    r.Get("/", rt.invoiceController.List)
                    r.Post("/", rt.invoiceController.Create)
                    r.Get("/{id}", rt.invoiceController.Get)
                    r.Put("/{id}", rt.invoiceController.Update)
                    r.Delete("/{id}", rt.invoiceController.Delete)
                    r.Post("/{id}/pdf", rt.invoiceController.GeneratePDF)
                    r.Post("/{id}/send", rt.invoiceController.SendEmail)
                })

                // Time Tracking
                r.Route("/track", func(r chi.Router) {
                    r.Get("/", rt.trackController.List)
                    r.Post("/", rt.trackController.Create)
                    r.Get("/{id}", rt.trackController.Get)
                    r.Put("/{id}", rt.trackController.Update)
                    r.Delete("/{id}", rt.trackController.Delete)
                    r.Get("/report", rt.trackController.GetReport)
                })
            })
        })
    })
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    rt.chi.ServeHTTP(w, r)
}
```

## Main Application

```go
// cmd/server/main.go
package main

import (
    "log"
    "net/http"
    "os"
    "ung-api/internal/config"
    "ung-api/internal/controllers"
    "ung-api/internal/database"
    "ung-api/internal/middleware"
    "ung-api/internal/repository"
    "ung-api/internal/router"
    "ung-api/internal/services"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize API database (users, auth)
    apiDB, err := database.InitAPIDatabase(cfg.APIDatabasePath)
    if err != nil {
        log.Fatal("Failed to initialize API database:", err)
    }

    // Repositories
    userRepo := repository.NewUserRepository(apiDB)

    // Services
    emailService := services.NewEmailService(cfg.GmailClientID, cfg.GmailClientSecret)
    authService := services.NewAuthService(userRepo, cfg.JWTSecret)
    invoiceService := services.NewInvoiceService(nil, nil, nil, emailService)

    // Scheduler
    scheduler := services.NewSchedulerService(apiDB, emailService)
    scheduler.Start()
    defer scheduler.Stop()

    // Controllers
    authController := controllers.NewAuthController(authService)
    companyController := controllers.NewCompanyController()
    clientController := controllers.NewClientController()
    contractController := controllers.NewContractController()
    invoiceController := controllers.NewInvoiceController(invoiceService)
    trackController := controllers.NewTrackController()

    // Middleware
    authMiddleware := middleware.AuthMiddleware(userRepo, cfg.JWTSecret)
    tenantMiddleware := middleware.TenantMiddleware(apiDB)

    // Router
    r := router.NewRouter(
        authController,
        companyController,
        clientController,
        contractController,
        invoiceController,
        trackController,
        authMiddleware,
        tenantMiddleware,
    )

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Starting UNG API server on port %s", port)
    if err := http.ListenAndServe(":"+port, r); err != nil {
        log.Fatal("Server failed:", err)
    }
}
```

## Configuration

```yaml
# .env.example
# API Configuration
PORT=8080
ENV=development

# Database
API_DATABASE_PATH=/home/user/.ung/api.db
USER_DATA_DIR=/home/user/.ung/users

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Gmail API
GMAIL_CLIENT_ID=your-client-id.apps.googleusercontent.com
GMAIL_CLIENT_SECRET=your-client-secret

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# CORS
ALLOWED_ORIGINS=http://localhost:3000,https://ung.app

# Scheduler
ENABLE_SCHEDULER=true
RECURRING_INVOICE_TIME=09:00
OVERDUE_REMINDER_TIME=10:00
```

## Deployment

### Docker

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -o ung-api cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy binary
COPY --from=builder /app/ung-api .

# Copy migrations
COPY migrations ./migrations

EXPOSE 8080

CMD ["./ung-api"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/root/.ung
    environment:
      - PORT=8080
      - ENV=production
      - API_DATABASE_PATH=/root/.ung/api.db
      - USER_DATA_DIR=/root/.ung/users
      - JWT_SECRET=${JWT_SECRET}
      - GMAIL_CLIENT_ID=${GMAIL_CLIENT_ID}
      - GMAIL_CLIENT_SECRET=${GMAIL_CLIENT_SECRET}
    restart: unless-stopped
```

### Systemd Service

```ini
# /etc/systemd/system/ung-api.service
[Unit]
Description=UNG API Service
After=network.target

[Service]
Type=simple
User=ung
WorkingDirectory=/opt/ung-api
ExecStart=/opt/ung-api/ung-api
Restart=on-failure
RestartSec=5s

Environment="PORT=8080"
Environment="ENV=production"
EnvironmentFile=/opt/ung-api/.env

[Install]
WantedBy=multi-user.target
```

## API Documentation

### Endpoints Summary

#### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/auth/me` - Get current user profile

#### Gmail Integration
- `GET /api/v1/gmail/auth` - Get Gmail OAuth URL
- `GET /api/v1/gmail/callback` - Gmail OAuth callback

#### Companies
- `GET /api/v1/companies` - List companies
- `POST /api/v1/companies` - Create company
- `GET /api/v1/companies/:id` - Get company
- `PUT /api/v1/companies/:id` - Update company
- `DELETE /api/v1/companies/:id` - Delete company

#### Clients
- `GET /api/v1/clients` - List clients
- `POST /api/v1/clients` - Create client
- `GET /api/v1/clients/:id` - Get client
- `PUT /api/v1/clients/:id` - Update client
- `DELETE /api/v1/clients/:id` - Delete client

#### Contracts
- `GET /api/v1/contracts` - List contracts
- `POST /api/v1/contracts` - Create contract
- `GET /api/v1/contracts/:id` - Get contract
- `PUT /api/v1/contracts/:id` - Update contract
- `DELETE /api/v1/contracts/:id` - Delete contract
- `POST /api/v1/contracts/:id/pdf` - Generate PDF

#### Invoices
- `GET /api/v1/invoices` - List invoices
- `POST /api/v1/invoices` - Create invoice
- `GET /api/v1/invoices/:id` - Get invoice
- `PUT /api/v1/invoices/:id` - Update invoice
- `DELETE /api/v1/invoices/:id` - Delete invoice
- `POST /api/v1/invoices/:id/pdf` - Generate PDF
- `POST /api/v1/invoices/:id/send` - Send via email

#### Time Tracking
- `GET /api/v1/track` - List time entries
- `POST /api/v1/track` - Create entry
- `GET /api/v1/track/:id` - Get entry
- `PUT /api/v1/track/:id` - Update entry
- `DELETE /api/v1/track/:id` - Delete entry
- `GET /api/v1/track/report` - Get time report

## Implementation Timeline

### Phase 1: Foundation (2 weeks)
- [ ] Project setup with Chi router
- [ ] Database models and migrations
- [ ] Authentication (JWT, middleware)
- [ ] Multi-tenant infrastructure
- [ ] Basic CRUD controllers

### Phase 2: Core Features (3 weeks)
- [ ] All resource controllers (companies, clients, contracts, invoices, tracking)
- [ ] PDF generation service (reuse CLI code)
- [ ] Input validation
- [ ] Error handling
- [ ] API documentation (Swagger)

### Phase 3: Email Integration (2 weeks)
- [ ] Gmail OAuth flow
- [ ] Email service implementation
- [ ] Template system for emails
- [ ] Attachment handling
- [ ] Email queue for reliability

### Phase 4: Automation (2 weeks)
- [ ] Scheduler service with cron
- [ ] Recurring invoice generation
- [ ] Overdue reminders
- [ ] Webhook system
- [ ] Background jobs

### Phase 5: Polish (1 week)
- [ ] Rate limiting
- [ ] Logging and monitoring
- [ ] Performance optimization
- [ ] Docker deployment
- [ ] CI/CD pipeline

**Total: 10 weeks**

## Security Best Practices

1. **Password Storage**: bcrypt with cost factor 12
2. **JWT Secrets**: Strong random secrets, rotated regularly
3. **SQL Injection**: GORM prevents SQL injection automatically
4. **XSS**: Sanitize all outputs
5. **CSRF**: Use CSRF tokens for state-changing operations
6. **Rate Limiting**: Prevent abuse
7. **HTTPS Only**: Enforce TLS in production
8. **Token Encryption**: Encrypt sensitive tokens (Gmail) at rest
9. **Input Validation**: Validate all inputs
10. **Audit Logging**: Log all sensitive operations

## Monitoring

### Metrics to Track
- Request rate and latency
- Error rates by endpoint
- Authentication failures
- Database connection pool stats
- Background job success/failure rates
- Email delivery rates

### Tools
- Prometheus for metrics
- Grafana for dashboards
- Sentry for error tracking
- Log aggregation (ELK stack or similar)

## Future Enhancements

1. **GraphQL API** - Alternative to REST
2. **WebSocket Support** - Real-time updates
3. **File Upload** - Logo, attachments
4. **Multi-language** - i18n support
5. **Advanced Reports** - Financial analytics
6. **Export Options** - CSV, Excel, JSON
7. **API Versioning** - v2, v3, etc.
8. **Webhooks** - External integrations
9. **OAuth Providers** - Google, GitHub, etc.
10. **Admin Dashboard** - User management

## Conclusion

This Go API provides a robust, scalable backend for the UNG ecosystem, featuring:
- **NestJS-like architecture** with clear separation of concerns
- **Chi router** for standard library compatibility
- **Multi-tenant** with isolated user databases
- **Gmail integration** for automated email sending
- **Recurring invoices** via scheduler
- **Reuses CLI code** for business logic and PDF generation
- **Production-ready** with Docker, monitoring, and security best practices

The API serves as the backbone for the Expo mobile app, Swift macOS app, and Telegram bot, providing a unified data layer while maintaining data isolation per user.
