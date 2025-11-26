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
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"      🚀 *UNG Bot*\n"+
			"  _Your Next Gig, Simplified_\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Welcome! I'm your personal billing assistant.\n\n"+
			"✨ *What I can do for you:*\n\n"+
			"📄 *Invoices* — Create & manage invoices\n"+
			"👥 *Clients* — Track your client database\n"+
			"⏱️ *Time* — Log hours & track work\n"+
			"💰 *Reports* — Revenue dashboards\n"+
			"📋 *Contracts* — Manage agreements\n"+
			"💸 *Expenses* — Track your costs\n\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"🔐 *Get Started*\n\n"+
			"Connect your UNG account to begin.\n"+
			"No account? Sign up free at:\n%s/register",
		h.webAppURL,
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔐 Connect Account", "auth_login"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📝 Create Free Account", h.webAppURL+"/register"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

func (h *StartHandler) sendMainMenu(chatID int64, name string) error {
	text := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"      🏠 *Main Menu*\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Hey %s! 👋\n\n"+
			"What would you like to do today?\n\n"+
			"💡 _Tip: Use /help for all commands_",
		name,
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 New Invoice", "action_invoice"),
			tgbotapi.NewInlineKeyboardButtonData("⏱️ Track Time", "action_track"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 Clients", "action_clients"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Contracts", "action_contracts"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Dashboard", "action_reports"),
			tgbotapi.NewInlineKeyboardButtonData("💸 Expenses", "action_expenses"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📑 All Invoices", "action_invoices_list"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "action_settings"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}
