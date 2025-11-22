# UNG Telegram Bot

**Conversational billing assistant for Telegram - Your Next Gig, in your pocket**

This document outlines the architecture and implementation of the UNG Telegram bot, a subscription-based conversational interface that connects to the UNG Go API.

## Overview

The UNG Telegram bot provides a natural language interface to the UNG billing system, allowing users to:
- Create and manage invoices via chat
- Track time with simple commands
- View reports and analytics
- Receive notifications for overdue invoices
- Generate and send PDFs directly from Telegram
- Quick actions with inline keyboards
- Voice message support for time tracking

## Why Telegram?

- **700M+ active users** worldwide
- **Rich bot API** with inline keyboards, file uploads, voice messages
- **Cross-platform** - works on mobile, desktop, web
- **Push notifications** built-in
- **No app installation** required
- **File sharing** up to 2GB
- **End-to-end encryption** available
- **Global reach** especially popular with freelancers and businesses

## Architecture

```
telegram-bot/
├── cmd/
│   └── bot/
│       └── main.go                 # Bot entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration
│   ├── handlers/
│   │   ├── start.go                # /start command
│   │   ├── auth.go                 # Authentication flow
│   │   ├── invoice.go              # Invoice commands
│   │   ├── client.go               # Client commands
│   │   ├── contract.go             # Contract commands
│   │   ├── track.go                # Time tracking commands
│   │   ├── report.go               # Reports and analytics
│   │   └── callback.go             # Inline button callbacks
│   ├── middleware/
│   │   ├── auth.go                 # Auth middleware
│   │   └── logger.go               # Logging middleware
│   ├── services/
│   │   ├── api_client.go           # UNG API client
│   │   ├── subscription.go         # RevenueCat integration
│   │   ├── nlp.go                  # Natural language processing
│   │   └── notification.go         # Notification service
│   ├── models/
│   │   ├── user.go                 # Telegram user model
│   │   ├── session.go              # Conversation state
│   │   └── subscription.go         # Subscription model
│   └── database/
│       ├── db.go                   # Redis for session state
│       └── repository.go           # User repository
├── pkg/
│   └── utils/
│       ├── keyboard.go             # Inline keyboard builders
│       ├── validator.go            # Input validation
│       └── formatter.go            # Message formatting
├── .env.example
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## Key Technologies

- **Go** - Programming language
- **telegram-bot-api** - Official Go bindings for Telegram Bot API
- **Redis** - Session state and conversation context
- **HTTP Client** - Communication with UNG Go API
- **RevenueCat** - Subscription management
- **Natural Language Processing** - Intent recognition (optional)

## Bot Commands

### Essential Commands

```
/start - Start the bot and authenticate
/help - Show available commands
/invoice - Create new invoice
/client - Manage clients
/contract - Manage contracts
/track - Log time or start timer
/report - View reports
/status - View account status
/settings - Configure preferences
/subscribe - Manage subscription
```

### Quick Actions

```
/quick_invoice - Fast invoice creation from template
/today - View today's tracked time
/week - View this week's summary
/unpaid - List unpaid invoices
/overdue - List overdue invoices
```

## Authentication Flow

### Initial Setup

```go
// internal/handlers/start.go
package handlers

import (
    "fmt"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartHandler struct {
    bot        *tgbotapi.BotAPI
    apiClient  *services.APIClient
    userRepo   *repository.UserRepository
}

func (h *StartHandler) Handle(update tgbotapi.Update) error {
    chatID := update.Message.Chat.ID
    telegramUser := update.Message.From

    // Check if user exists
    user, err := h.userRepo.GetByTelegramID(telegramUser.ID)

    if err != nil {
        // New user - start authentication
        msg := tgbotapi.NewMessage(chatID,
            "👋 Welcome to UNG - Your Next Gig, Simplified!\n\n"+
            "To get started, please authenticate with your UNG account.\n\n"+
            "If you don't have an account yet, you can create one at https://ung.app")

        keyboard := tgbotapi.NewInlineKeyboardMarkup(
            tgbotapi.NewInlineKeyboardRow(
                tgbotapi.NewInlineKeyboardButtonURL("Create Account", "https://ung.app/register"),
            ),
            tgbotapi.NewInlineKeyboardRow(
                tgbotapi.NewInlineKeyboardButtonData("I have an account", "auth_login"),
            ),
        )
        msg.ReplyMarkup = keyboard

        _, err = h.bot.Send(msg)
        return err
    }

    // Existing user
    return h.sendMainMenu(chatID, user)
}

func (h *StartHandler) sendMainMenu(chatID int64, user *models.User) error {
    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("Welcome back, %s! 👋\n\nWhat would you like to do?", user.Name))

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📄 Create Invoice", "action_invoice"),
            tgbotapi.NewInlineKeyboardButtonData("👥 Clients", "action_clients"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("⏱️ Track Time", "action_track"),
            tgbotapi.NewInlineKeyboardButtonData("📊 Reports", "action_reports"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "action_settings"),
        ),
    )
    msg.ReplyMarkup = keyboard

    _, err := h.bot.Send(msg)
    return err
}
```

### Authentication with API

```go
// internal/handlers/auth.go
package handlers

import (
    "fmt"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AuthHandler struct {
    bot        *tgbotapi.BotAPI
    apiClient  *services.APIClient
    sessionMgr *services.SessionManager
    userRepo   *repository.UserRepository
}

func (h *AuthHandler) HandleLogin(chatID int64, telegramUserID int64) error {
    // Generate auth token for this Telegram user
    authToken := generateRandomToken()

    // Store in session (expires in 5 minutes)
    h.sessionMgr.SetAuthToken(telegramUserID, authToken)

    // Send auth link
    authURL := fmt.Sprintf("https://ung.app/telegram/auth?token=%s", authToken)

    msg := tgbotapi.NewMessage(chatID,
        "🔐 To authenticate, please click the link below and login to your UNG account:\n\n"+
        "This link expires in 5 minutes.")

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonURL("Authenticate", authURL),
        ),
    )
    msg.ReplyMarkup = keyboard

    _, err := h.bot.Send(msg)
    return err
}

// Called by webhook when user completes web auth
func (h *AuthHandler) CompleteAuth(telegramUserID int64, apiToken string, userID uint) error {
    // Store API token for this Telegram user
    user := &models.User{
        TelegramID: telegramUserID,
        APIToken:   apiToken,
        UserID:     userID,
    }

    err := h.userRepo.Create(user)
    if err != nil {
        return err
    }

    // Notify user
    msg := tgbotapi.NewMessage(int64(telegramUserID),
        "✅ Authentication successful!\n\nYou can now use all UNG features via Telegram.")

    h.bot.Send(msg)
    return nil
}
```

## Conversation Flow - Invoice Creation

```go
// internal/handlers/invoice.go
package handlers

import (
    "fmt"
    "strconv"
    "strings"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type InvoiceHandler struct {
    bot        *tgbotapi.BotAPI
    apiClient  *services.APIClient
    sessionMgr *services.SessionManager
}

func (h *InvoiceHandler) HandleCreate(update tgbotapi.Update) error {
    chatID := update.Message.Chat.ID
    telegramUserID := update.Message.From.ID

    // Get user's API token
    user, err := h.getUserByTelegramID(telegramUserID)
    if err != nil {
        return h.sendAuthRequired(chatID)
    }

    // Start conversation flow
    session := &models.Session{
        UserID: telegramUserID,
        State:  "invoice_select_client",
        Data:   make(map[string]interface{}),
    }
    h.sessionMgr.Set(session)

    // Fetch clients from API
    clients, err := h.apiClient.ListClients(user.APIToken)
    if err != nil {
        return h.sendError(chatID, "Failed to fetch clients")
    }

    if len(clients) == 0 {
        msg := tgbotapi.NewMessage(chatID,
            "You don't have any clients yet. Let's create one first!\n\n"+
            "Please send the client name:")
        h.bot.Send(msg)
        session.State = "client_create_name"
        h.sessionMgr.Set(session)
        return nil
    }

    // Show client selection
    return h.showClientSelection(chatID, clients, session)
}

func (h *InvoiceHandler) showClientSelection(chatID int64, clients []Client, session *models.Session) error {
    var buttons [][]tgbotapi.InlineKeyboardButton

    for _, client := range clients {
        row := tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData(
                client.Name,
                fmt.Sprintf("invoice_client_%d", client.ID),
            ),
        )
        buttons = append(buttons, row)
    }

    // Add "Create new client" option
    buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("➕ Create new client", "invoice_new_client"),
    ))

    keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

    msg := tgbotapi.NewMessage(chatID, "Select a client for this invoice:")
    msg.ReplyMarkup = keyboard

    _, err := h.bot.Send(msg)
    return err
}

func (h *InvoiceHandler) HandleClientSelected(callbackQuery *tgbotapi.CallbackQuery) error {
    chatID := callbackQuery.Message.Chat.ID
    data := callbackQuery.Data // "invoice_client_123"

    // Parse client ID
    parts := strings.Split(data, "_")
    clientID, _ := strconv.Atoi(parts[2])

    // Get session
    session, _ := h.sessionMgr.Get(callbackQuery.From.ID)
    session.Data["client_id"] = clientID
    session.State = "invoice_amount"
    h.sessionMgr.Set(session)

    // Ask for amount
    msg := tgbotapi.NewMessage(chatID,
        "Great! Now please enter the invoice amount:\n\n"+
        "Example: 1500 or 1500.50")

    h.bot.Send(msg)

    // Answer callback to remove loading state
    callback := tgbotapi.NewCallback(callbackQuery.ID, "Client selected")
    h.bot.Request(callback)

    return nil
}

func (h *InvoiceHandler) HandleAmountInput(message *tgbotapi.Message) error {
    chatID := message.Chat.ID
    text := message.Text

    // Parse amount
    amount, err := strconv.ParseFloat(text, 64)
    if err != nil {
        msg := tgbotapi.NewMessage(chatID,
            "❌ Invalid amount. Please enter a number (e.g., 1500 or 1500.50)")
        h.bot.Send(msg)
        return nil
    }

    // Get session
    session, _ := h.sessionMgr.Get(message.From.ID)
    session.Data["amount"] = amount
    session.State = "invoice_description"
    h.sessionMgr.Set(session)

    // Ask for description
    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("Amount: $%.2f\n\nNow add a description or notes for this invoice:", amount))

    h.bot.Send(msg)
    return nil
}

func (h *InvoiceHandler) HandleDescriptionInput(message *tgbotapi.Message) error {
    chatID := message.Chat.ID
    description := message.Text

    // Get session
    session, _ := h.sessionMgr.Get(message.From.ID)
    session.Data["description"] = description
    session.State = "invoice_due_date"
    h.sessionMgr.Set(session)

    // Ask for due date
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("7 days", "invoice_due_7"),
            tgbotapi.NewInlineKeyboardButtonData("14 days", "invoice_due_14"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("30 days", "invoice_due_30"),
            tgbotapi.NewInlineKeyboardButtonData("Custom", "invoice_due_custom"),
        ),
    )

    msg := tgbotapi.NewMessage(chatID, "When is this invoice due?")
    msg.ReplyMarkup = keyboard

    h.bot.Send(msg)
    return nil
}

func (h *InvoiceHandler) CreateInvoice(callbackQuery *tgbotapi.CallbackQuery, days int) error {
    chatID := callbackQuery.Message.Chat.ID
    telegramUserID := callbackQuery.From.ID

    // Get user and session
    user, _ := h.getUserByTelegramID(telegramUserID)
    session, _ := h.sessionMgr.Get(telegramUserID)

    // Build invoice request
    dueDate := time.Now().AddDate(0, 0, days)
    invoiceReq := &InvoiceCreateRequest{
        ClientID:    session.Data["client_id"].(int),
        Amount:      session.Data["amount"].(float64),
        Currency:    "USD",
        Description: session.Data["description"].(string),
        DueDate:     dueDate.Format("2006-01-02"),
        LineItems: []LineItem{
            {
                ItemName:    session.Data["description"].(string),
                Quantity:    1,
                Rate:        session.Data["amount"].(float64),
                Amount:      session.Data["amount"].(float64),
            },
        },
    }

    // Call API
    invoice, err := h.apiClient.CreateInvoice(user.APIToken, invoiceReq)
    if err != nil {
        return h.sendError(chatID, "Failed to create invoice")
    }

    // Clear session
    h.sessionMgr.Delete(telegramUserID)

    // Send success message with actions
    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("✅ Invoice #%s created successfully!\n\n"+
            "Amount: $%.2f %s\n"+
            "Due Date: %s\n\n"+
            "What would you like to do next?",
            invoice.InvoiceNum,
            invoice.Amount,
            invoice.Currency,
            invoice.DueDate.Format("January 2, 2006")))

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📄 Generate PDF", fmt.Sprintf("invoice_pdf_%d", invoice.ID)),
            tgbotapi.NewInlineKeyboardButtonData("📧 Send Email", fmt.Sprintf("invoice_email_%d", invoice.ID)),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📋 View Details", fmt.Sprintf("invoice_view_%d", invoice.ID)),
            tgbotapi.NewInlineKeyboardButtonData("🏠 Main Menu", "main_menu"),
        ),
    )
    msg.ReplyMarkup = keyboard

    h.bot.Send(msg)

    // Answer callback
    callback := tgbotapi.NewCallback(callbackQuery.ID, "Invoice created!")
    h.bot.Request(callback)

    return nil
}

// Send PDF directly via Telegram
func (h *InvoiceHandler) HandleSendPDF(callbackQuery *tgbotapi.CallbackQuery) error {
    chatID := callbackQuery.Message.Chat.ID
    data := callbackQuery.Data // "invoice_pdf_123"

    // Parse invoice ID
    parts := strings.Split(data, "_")
    invoiceID, _ := strconv.Atoi(parts[2])

    // Get user
    user, _ := h.getUserByTelegramID(callbackQuery.From.ID)

    // Generate PDF via API
    pdfPath, err := h.apiClient.GenerateInvoicePDF(user.APIToken, invoiceID)
    if err != nil {
        return h.sendError(chatID, "Failed to generate PDF")
    }

    // Download PDF from API server
    pdfData, err := h.apiClient.DownloadFile(user.APIToken, pdfPath)
    if err != nil {
        return h.sendError(chatID, "Failed to download PDF")
    }

    // Send as document
    doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
        Name:  fmt.Sprintf("invoice_%d.pdf", invoiceID),
        Bytes: pdfData,
    })
    doc.Caption = "📄 Your invoice PDF is ready!"

    h.bot.Send(doc)

    // Answer callback
    callback := tgbotapi.NewCallback(callbackQuery.ID, "PDF generated!")
    h.bot.Request(callback)

    return nil
}
```

## Time Tracking with Voice Messages

```go
// internal/handlers/track.go
package handlers

type TrackHandler struct {
    bot        *tgbotapi.BotAPI
    apiClient  *services.APIClient
}

// Handle voice messages for time tracking
func (h *TrackHandler) HandleVoice(message *tgbotapi.Message) error {
    chatID := message.Chat.ID

    // Download voice message
    file, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: message.Voice.FileID})
    if err != nil {
        return err
    }

    // Get download URL
    voiceURL := file.Link(h.bot.Token)

    // Send for transcription (use external service like OpenAI Whisper)
    text, err := h.transcribeAudio(voiceURL)
    if err != nil {
        return h.sendError(chatID, "Failed to transcribe voice message")
    }

    // Parse intent from transcription
    // Example: "I worked 3 hours on the website project for Acme Corp"
    entry, err := h.parseTimeEntry(text)
    if err != nil {
        msg := tgbotapi.NewMessage(chatID,
            fmt.Sprintf("I heard: \"%s\"\n\nBut I couldn't understand the time entry. "+
                "Please try again or use the /track command.", text))
        h.bot.Send(msg)
        return nil
    }

    // Show confirmation
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "track_confirm"),
            tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "track_cancel"),
        ),
    )

    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("I heard: \"%s\"\n\n"+
            "⏱️ Time Entry:\n"+
            "Duration: %.1f hours\n"+
            "Project: %s\n"+
            "Client: %s\n\n"+
            "Is this correct?",
            text, entry.Hours, entry.Project, entry.Client))
    msg.ReplyMarkup = keyboard

    h.bot.Send(msg)
    return nil
}

// Quick time tracking with text
func (h *TrackHandler) HandleQuick(message *tgbotapi.Message) error {
    chatID := message.Chat.ID
    text := message.Text

    // Parse formats:
    // "3h website design"
    // "2.5 hours client meeting"
    // "30m email"

    entry, err := h.parseQuickEntry(text)
    if err != nil {
        return h.sendError(chatID, "Invalid format. Try: '3h website design' or '2.5 hours client meeting'")
    }

    // Create time entry via API
    user, _ := h.getUserByTelegramID(message.From.ID)
    created, err := h.apiClient.CreateTimeEntry(user.APIToken, entry)
    if err != nil {
        return h.sendError(chatID, "Failed to create time entry")
    }

    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("✅ Logged %.1f hours\n\n%s", entry.Hours, entry.Description))

    h.bot.Send(msg)
    return nil
}
```

## Notifications

```go
// internal/services/notification.go
package services

type NotificationService struct {
    bot       *tgbotapi.BotAPI
    apiClient *APIClient
    userRepo  *repository.UserRepository
}

// Send notification to user
func (s *NotificationService) SendInvoiceOverdue(userID uint, invoice *Invoice) error {
    // Get Telegram user by UNG user ID
    user, err := s.userRepo.GetByUserID(userID)
    if err != nil {
        return err
    }

    chatID := user.TelegramID

    msg := tgbotapi.NewMessage(chatID,
        fmt.Sprintf("⚠️ Invoice Overdue!\n\n"+
            "Invoice #%s to %s\n"+
            "Amount: $%.2f %s\n"+
            "Due Date: %s\n\n"+
            "This invoice is now %d days overdue.",
            invoice.InvoiceNum,
            invoice.Client.Name,
            invoice.Amount,
            invoice.Currency,
            invoice.DueDate.Format("January 2, 2006"),
            int(time.Since(invoice.DueDate).Hours()/24)))

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📧 Send Reminder", fmt.Sprintf("invoice_remind_%d", invoice.ID)),
            tgbotapi.NewInlineKeyboardButtonData("📄 View PDF", fmt.Sprintf("invoice_pdf_%d", invoice.ID)),
        ),
    )
    msg.ReplyMarkup = keyboard

    _, err = s.bot.Send(msg)
    return err
}

// Start notification listener
func (s *NotificationService) Start() {
    // Listen to API webhooks or poll regularly
    // Send notifications based on events:
    // - Invoice overdue
    // - Payment received
    // - Recurring invoice generated
    // - Time tracking reminder
}
```

## Subscription Management

```go
// internal/services/subscription.go
package services

type SubscriptionService struct {
    revenueCatAPIKey string
}

func (s *SubscriptionService) CheckAccess(telegramUserID int64) (bool, error) {
    // Query RevenueCat for subscription status
    subscriber, err := s.getSubscriberInfo(telegramUserID)
    if err != nil {
        return false, err
    }

    // Check if user has active subscription
    if subscriber.Entitlements["pro"] != nil && subscriber.Entitlements["pro"].IsActive {
        return true, nil
    }

    return false, nil
}

func (s *SubscriptionService) GetSubscribeURL(telegramUserID int64) string {
    // Generate subscription link
    return fmt.Sprintf("https://ung.app/subscribe?ref=telegram_%d", telegramUserID)
}
```

### Subscription Middleware

```go
// internal/middleware/subscription.go
package middleware

func SubscriptionMiddleware(subService *services.SubscriptionService) func(tgbotapi.Update) bool {
    return func(update tgbotapi.Update) bool {
        var userID int64

        if update.Message != nil {
            userID = update.Message.From.ID
        } else if update.CallbackQuery != nil {
            userID = update.CallbackQuery.From.ID
        }

        // Check subscription
        hasAccess, _ := subService.CheckAccess(userID)

        return hasAccess
    }
}

// Send upgrade prompt
func SendUpgradePrompt(bot *tgbotapi.BotAPI, chatID int64, subService *services.SubscriptionService) {
    msg := tgbotapi.NewMessage(chatID,
        "🔒 This is a Pro feature\n\n"+
        "Upgrade to UNG Pro to unlock:\n"+
        "✅ Unlimited invoices\n"+
        "✅ Automated recurring invoices\n"+
        "✅ Email integration\n"+
        "✅ Advanced reports\n"+
        "✅ Priority support\n\n"+
        "Starting at $9/month")

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonURL(
                "Upgrade Now",
                subService.GetSubscribeURL(chatID),
            ),
        ),
    )
    msg.ReplyMarkup = keyboard

    bot.Send(msg)
}
```

## Main Bot Application

```go
// cmd/bot/main.go
package main

import (
    "log"
    "os"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "ung-bot/internal/handlers"
    "ung-bot/internal/services"
    "ung-bot/internal/middleware"
)

func main() {
    // Load config
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    apiBaseURL := os.Getenv("UNG_API_URL")

    // Initialize bot
    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        log.Fatal(err)
    }

    bot.Debug = os.Getenv("DEBUG") == "true"
    log.Printf("Authorized on account %s", bot.Self.UserName)

    // Initialize services
    apiClient := services.NewAPIClient(apiBaseURL)
    sessionMgr := services.NewSessionManager()
    subService := services.NewSubscriptionService(os.Getenv("REVENUECAT_API_KEY"))
    notificationService := services.NewNotificationService(bot, apiClient, userRepo)

    // Initialize handlers
    startHandler := handlers.NewStartHandler(bot, apiClient, userRepo)
    authHandler := handlers.NewAuthHandler(bot, apiClient, sessionMgr, userRepo)
    invoiceHandler := handlers.NewInvoiceHandler(bot, apiClient, sessionMgr)
    trackHandler := handlers.NewTrackHandler(bot, apiClient)

    // Start notification service
    go notificationService.Start()

    // Start listening for updates
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        // Handle different update types
        if update.Message != nil {
            handleMessage(update, bot, handlers, sessionMgr, subService)
        } else if update.CallbackQuery != nil {
            handleCallback(update, bot, handlers, sessionMgr, subService)
        }
    }
}

func handleMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI, handlers map[string]interface{}, sessionMgr *services.SessionManager, subService *services.SubscriptionService) {
    msg := update.Message

    // Commands
    if msg.IsCommand() {
        switch msg.Command() {
        case "start":
            handlers["start"].(*handlers.StartHandler).Handle(update)
        case "invoice":
            if checkSubscription(msg.From.ID, subService, bot, msg.Chat.ID) {
                handlers["invoice"].(*handlers.InvoiceHandler).HandleCreate(update)
            }
        case "track":
            handlers["track"].(*handlers.TrackHandler).HandleQuick(msg)
        // ... other commands
        }
        return
    }

    // Voice messages
    if msg.Voice != nil {
        handlers["track"].(*handlers.TrackHandler).HandleVoice(msg)
        return
    }

    // Handle conversation state
    session, err := sessionMgr.Get(msg.From.ID)
    if err == nil {
        handleConversationState(update, session, handlers)
    }
}

func handleCallback(update tgbotapi.Update, bot *tgbotapi.BotAPI, handlers map[string]interface{}, sessionMgr *services.SessionManager, subService *services.SubscriptionService) {
    callback := update.CallbackQuery
    data := callback.Data

    // Route based on callback data prefix
    if strings.HasPrefix(data, "invoice_") {
        handlers["invoice"].(*handlers.InvoiceHandler).HandleCallback(callback)
    } else if strings.HasPrefix(data, "track_") {
        handlers["track"].(*handlers.TrackHandler).HandleCallback(callback)
    }
    // ... other routes
}

func checkSubscription(userID int64, subService *services.SubscriptionService, bot *tgbotapi.BotAPI, chatID int64) bool {
    hasAccess, _ := subService.CheckAccess(userID)
    if !hasAccess {
        middleware.SendUpgradePrompt(bot, chatID, subService)
        return false
    }
    return true
}
```

## Configuration

```bash
# .env
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
UNG_API_URL=https://api.ung.app
REVENUECAT_API_KEY=your_revenuecat_api_key
REDIS_URL=redis://localhost:6379
DEBUG=false
```

## Deployment

### Docker

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ung-bot cmd/bot/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/ung-bot .

CMD ["./ung-bot"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  bot:
    build: .
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - UNG_API_URL=${UNG_API_URL}
      - REVENUECAT_API_KEY=${REVENUECAT_API_KEY}
      - REDIS_URL=redis://redis:6379
    depends_on:
      - redis
    restart: unless-stopped

  redis:
    image: redis:alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  redis_data:
```

## Monetization Strategy

### Free Tier
- 5 invoices per month
- 3 clients
- Basic time tracking
- Manual PDF generation

### Pro Tier ($9/month)
- Unlimited invoices
- Unlimited clients
- Automated recurring invoices
- Email integration
- Advanced reports
- Voice commands
- Priority support

### Business Tier ($29/month)
- Everything in Pro
- Team collaboration
- Custom branding
- API access
- White-label option
- Dedicated support

## Implementation Timeline

### Phase 1: Core Bot (2 weeks)
- [ ] Bot setup with telegram-bot-api
- [ ] Authentication flow
- [ ] Basic commands (/start, /help)
- [ ] Session management with Redis
- [ ] API client integration

### Phase 2: Invoice Management (2 weeks)
- [ ] Invoice creation flow
- [ ] Client management
- [ ] PDF generation and delivery
- [ ] Email sending
- [ ] Inline keyboards

### Phase 3: Time Tracking (1 week)
- [ ] Quick text commands
- [ ] Voice message support
- [ ] Timer functionality
- [ ] Reports

### Phase 4: Notifications (1 week)
- [ ] Overdue invoice alerts
- [ ] Recurring invoice notifications
- [ ] Daily/weekly summaries
- [ ] Custom reminders

### Phase 5: Subscriptions (1 week)
- [ ] RevenueCat integration
- [ ] Subscription middleware
- [ ] Upgrade prompts
- [ ] Feature gating

### Phase 6: Polish (1 week)
- [ ] Error handling
- [ ] Logging and monitoring
- [ ] Performance optimization
- [ ] Documentation
- [ ] Testing

**Total: 8 weeks**

## Advanced Features

### Natural Language Processing
- Intent recognition for flexible commands
- Support for multiple languages
- Context-aware responses
- Smart suggestions

### Telegram Web App
- Mini app for complex workflows
- Rich UI within Telegram
- Payment integration
- File uploads

### Bot Analytics
- Track command usage
- User engagement metrics
- Conversion tracking
- A/B testing

## Security

1. **Token Storage**: Encrypt API tokens in Redis
2. **Session Timeout**: Expire sessions after 24 hours
3. **Rate Limiting**: Prevent spam and abuse
4. **Input Validation**: Sanitize all user inputs
5. **HTTPS Only**: Secure webhook endpoint
6. **Audit Logging**: Log all sensitive operations

## Conclusion

The UNG Telegram bot provides a powerful, conversational interface to the UNG billing system, making it accessible to users wherever they are. With subscription-based monetization, voice command support, and seamless API integration, it offers a unique value proposition in the freelance billing space.

Key benefits:
- **Accessibility** - Works on any device with Telegram
- **Convenience** - Create invoices via chat in seconds
- **Notifications** - Never miss an overdue invoice
- **Voice Support** - Track time hands-free
- **Subscription Model** - Recurring revenue stream
- **No App Store** - Direct distribution to users
- **Global Reach** - 700M+ potential users

The bot complements the CLI, mobile apps, and web interface, providing users with maximum flexibility in how they manage their billing and invoicing.
