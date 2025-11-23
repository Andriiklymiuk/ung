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

	// Start listening for updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is running. Press Ctrl+C to stop.")

	for update := range updates {
		// Handle messages
		if update.Message != nil {
			if err := handleMessage(update.Message, bot, startHandler, helpHandler, invoiceHandler, sessionMgr); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}

		// Handle callback queries
		if update.CallbackQuery != nil {
			if err := handleCallback(update.CallbackQuery, bot, invoiceHandler, sessionMgr); err != nil {
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
			return invoiceHandler.HandleList(message)
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
		case models.StateInvoiceAmount:
			return invoiceHandler.HandleAmountInput(message)
		case models.StateInvoiceDescription:
			return invoiceHandler.HandleDescriptionInput(message)
		case models.StateClientCreateName:
			// Handle client creation
			// TODO: Implement client creation handler
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
	sessionMgr *services.SessionManager,
) error {
	data := callbackQuery.Data

	// Route based on callback data prefix
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
