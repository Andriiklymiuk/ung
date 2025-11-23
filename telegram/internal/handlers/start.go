package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/services"
)

// StartHandler handles /start command
type StartHandler struct {
	bot        *tgbotapi.BotAPI
	sessionMgr *services.SessionManager
	webAppURL  string
}

// NewStartHandler creates a new start handler
func NewStartHandler(bot *tgbotapi.BotAPI, sessionMgr *services.SessionManager, webAppURL string) *StartHandler {
	return &StartHandler{
		bot:        bot,
		sessionMgr: sessionMgr,
		webAppURL:  webAppURL,
	}
}

// Handle handles the start command
func (h *StartHandler) Handle(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check if user is authenticated
	if h.sessionMgr.IsAuthenticated(telegramID) {
		user := h.sessionMgr.GetUser(telegramID)
		return h.sendMainMenu(chatID, user.Name)
	}

	// New user - send welcome and auth instructions
	text := fmt.Sprintf(
		"👋 *Welcome to UNG - Your Next Gig, Simplified!*\n\n"+
			"I'm your billing assistant. I can help you:\n"+
			"• 📄 Create and manage invoices\n"+
			"• 👥 Track clients\n"+
			"• ⏱️ Log time and generate invoices\n"+
			"• 📊 View reports\n\n"+
			"To get started, please authenticate with your UNG account.\n\n"+
			"Don't have an account? Create one at %s/register",
		h.webAppURL,
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Create Account", h.webAppURL+"/register"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("I have an account", "auth_login"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

func (h *StartHandler) sendMainMenu(chatID int64, name string) error {
	text := fmt.Sprintf("Welcome back, %s! 👋\n\nWhat would you like to do?", name)

	msg := tgbotapi.NewMessage(chatID, text)

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
