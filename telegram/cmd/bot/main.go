package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/config"
	"ung-telegram/internal/handlers"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Println("Starting UNG Telegram Bot...")

	// Initialize bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatal("Failed to initialize bot:", err)
	}

	bot.Debug = cfg.Debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize services
	apiClient := services.NewAPIClient(cfg.APIURL)
	sessionMgr := services.NewSessionManager()

	// Initialize handlers
	startHandler := handlers.NewStartHandler(bot, sessionMgr, cfg.WebAppURL)
	helpHandler := handlers.NewHelpHandler(bot)
	invoiceHandler := handlers.NewInvoiceHandler(bot, apiClient, sessionMgr)
	clientHandler := handlers.NewClientHandler(bot, apiClient, sessionMgr)
	companyHandler := handlers.NewCompanyHandler(bot, apiClient, sessionMgr)
	contractHandler := handlers.NewContractHandler(bot, apiClient, sessionMgr)
	expenseHandler := handlers.NewExpenseHandler(bot, apiClient, sessionMgr)
	trackingHandler := handlers.NewTrackingHandler(bot, apiClient, sessionMgr)

	// Start listening for updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is running. Press Ctrl+C to stop.")

	for update := range updates {
		// Handle messages
		if update.Message != nil {
			if err := handleMessage(update.Message, bot, startHandler, helpHandler, invoiceHandler, clientHandler, companyHandler, contractHandler, expenseHandler, trackingHandler, sessionMgr); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}

		// Handle callback queries
		if update.CallbackQuery != nil {
			if err := handleCallback(update.CallbackQuery, bot, invoiceHandler, clientHandler, contractHandler, expenseHandler, trackingHandler, sessionMgr); err != nil {
				log.Printf("Error handling callback: %v", err)
			}
		}
	}
}

func handleMessage(
	message *tgbotapi.Message,
	bot *tgbotapi.BotAPI,
	startHandler *handlers.StartHandler,
	helpHandler *handlers.HelpHandler,
	invoiceHandler *handlers.InvoiceHandler,
	clientHandler *handlers.ClientHandler,
	companyHandler *handlers.CompanyHandler,
	contractHandler *handlers.ContractHandler,
	expenseHandler *handlers.ExpenseHandler,
	trackingHandler *handlers.TrackingHandler,
	sessionMgr *services.SessionManager,
) error {
	// Handle commands
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			return startHandler.Handle(message)
		case "help":
			return helpHandler.Handle(message)
		case "invoice":
			return invoiceHandler.HandleCreate(message)
		case "invoices":
			return invoiceHandler.HandleListWithPDF(message)
		case "pdf":
			return invoiceHandler.HandlePDF(message)
		case "client":
			return clientHandler.HandleCreate(message)
		case "clients":
			return clientHandler.HandleList(message)
		case "company":
			return companyHandler.HandleCreate(message)
		case "companies":
			return companyHandler.HandleList(message)
		case "contract":
			return contractHandler.HandleCreate(message)
		case "contracts":
			return contractHandler.HandleList(message)
		case "expense":
			return expenseHandler.HandleCreate(message)
		case "expenses":
			return expenseHandler.HandleList(message)
		case "track":
			return trackingHandler.HandleStart(message)
		case "tracking":
			return trackingHandler.HandleList(message)
		case "stop":
			return trackingHandler.HandleStop(message)
		case "active":
			return trackingHandler.HandleActive(message)
		case "log":
			return trackingHandler.HandleLog(message)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Try /help")
			bot.Send(msg)
		}
		return nil
	}

	// Handle conversation states
	telegramID := message.From.ID
	session := sessionMgr.GetSession(telegramID)

	if session != nil {
		switch models.SessionState(session.State) {
		// Invoice states
		case models.StateInvoiceAmount:
			return invoiceHandler.HandleAmountInput(message)
		case models.StateInvoiceDescription:
			return invoiceHandler.HandleDescriptionInput(message)
		// Client states
		case models.StateClientCreateName:
			return clientHandler.HandleNameInput(message)
		case models.StateClientCreateEmail:
			return clientHandler.HandleEmailInput(message)
		case models.StateClientCreateAddress:
			return clientHandler.HandleAddressInput(message)
		case models.StateClientCreateTaxID:
			return clientHandler.HandleTaxIDInput(message)
		// Company states
		case models.StateCompanyCreateName:
			return companyHandler.HandleNameInput(message)
		case models.StateCompanyCreateEmail:
			return companyHandler.HandleEmailInput(message)
		case models.StateCompanyCreatePhone:
			return companyHandler.HandlePhoneInput(message)
		case models.StateCompanyCreateAddress:
			return companyHandler.HandleAddressInput(message)
		case models.StateCompanyCreateTaxID:
			return companyHandler.HandleTaxIDInput(message)
		// Contract states
		case models.StateContractName:
			return contractHandler.HandleNameInput(message)
		case models.StateContractRate:
			return contractHandler.HandleRateInput(message)
		// Expense states
		case models.StateExpenseDescription:
			return expenseHandler.HandleDescriptionInput(message)
		case models.StateExpenseAmount:
			return expenseHandler.HandleAmountInput(message)
		case models.StateExpenseVendor:
			return expenseHandler.HandleVendorInput(message)
		// Tracking log states
		case models.StateTrackLogHours:
			return trackingHandler.HandleHoursInput(message)
		case models.StateTrackLogProject:
			return trackingHandler.HandleProjectInput(message)
		case models.StateTrackLogNotes:
			return trackingHandler.HandleNotesInput(message)
		}
	}

	// Default: show help
	msg := tgbotapi.NewMessage(message.Chat.ID, "I didn't understand that. Try /help for available commands.")
	_, err := bot.Send(msg)
	return err
}

func handleCallback(
	callbackQuery *tgbotapi.CallbackQuery,
	bot *tgbotapi.BotAPI,
	invoiceHandler *handlers.InvoiceHandler,
	clientHandler *handlers.ClientHandler,
	contractHandler *handlers.ContractHandler,
	expenseHandler *handlers.ExpenseHandler,
	trackingHandler *handlers.TrackingHandler,
	sessionMgr *services.SessionManager,
) error {
	data := callbackQuery.Data

	// Route based on callback data prefix
	// Invoice callbacks
	if strings.HasPrefix(data, "invoice_client_") {
		return invoiceHandler.HandleClientSelected(callbackQuery)
	}

	if strings.HasPrefix(data, "invoice_due_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			var days int
			switch parts[2] {
			case "7":
				days = 7
			case "14":
				days = 14
			case "30":
				days = 30
			default:
				days = 30
			}
			return invoiceHandler.CreateInvoice(callbackQuery, days)
		}
	}

	if data == "action_invoice" {
		// Convert callback to message for easier handling
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		return invoiceHandler.HandleCreate(msg)
	}

	if data == "invoice_new_client" {
		// Start client creation flow
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		return clientHandler.HandleCreate(msg)
	}

	// Contract callbacks
	if strings.HasPrefix(data, "contract_client_") {
		return contractHandler.HandleClientSelected(callbackQuery)
	}

	if strings.HasPrefix(data, "contract_type_") {
		return contractHandler.HandleTypeSelected(callbackQuery)
	}

	// Expense callbacks
	if strings.HasPrefix(data, "expense_category_") {
		return expenseHandler.HandleCategorySelected(callbackQuery)
	}

	// Tracking callbacks
	if data == "tracking_stop" {
		return trackingHandler.HandleStopCallback(callbackQuery)
	}

	if strings.HasPrefix(data, "log_contract_") {
		return trackingHandler.HandleContractSelected(callbackQuery)
	}

	// Invoice PDF callbacks
	if strings.HasPrefix(data, "invoice_pdf_") {
		return invoiceHandler.HandlePDFCallback(callbackQuery)
	}

	// Auth callback
	if data == "auth_login" {
		chatID := callbackQuery.Message.Chat.ID
		text := "To authenticate, please visit:\n\n" +
			"https://ung.app/telegram/auth\n\n" +
			"And follow the instructions to link your account."

		msg := tgbotapi.NewMessage(chatID, text)
		bot.Send(msg)

		callback := tgbotapi.NewCallback(callbackQuery.ID, "Please visit the link to authenticate")
		bot.Request(callback)
		return nil
	}

	// Answer callback with default response
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	_, err := bot.Request(callback)
	return err
}
