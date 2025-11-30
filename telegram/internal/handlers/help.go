package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HelpHandler handles /help command
type HelpHandler struct {
	bot *tgbotapi.BotAPI
}

// NewHelpHandler creates a new help handler
func NewHelpHandler(bot *tgbotapi.BotAPI) *HelpHandler {
	return &HelpHandler{bot: bot}
}

// Handle handles the help command
func (h *HelpHandler) Handle(message *tgbotapi.Message) error {
	text := `━━━━━━━━━━━━━━━━━━━━━━
      📚 *Help & Commands*
━━━━━━━━━━━━━━━━━━━━━━

*💰 INVOICING*
├ /invoice — Create new invoice
├ /invoices — List all invoices
└ /pdf ‹num› — Get PDF (e.g. /pdf INV-001)

*👥 CLIENTS & COMPANIES*
├ /client — Add new client
├ /clients — View all clients
├ /company — Add company
└ /companies — List companies

*📋 CONTRACTS*
├ /contract — Create contract
└ /contracts — List contracts

*⏱️ TIME TRACKING*
├ /track — Start timer ▶️
├ /stop — Stop timer ⏹️
├ /active — View current session
├ /log — Log time manually
└ /tracking — View history

*💸 EXPENSES*
├ /expense — Add expense
└ /expenses — List expenses

*📊 REPORTS*
└ /dashboard — Revenue overview

*🎯 JOB HUNTER*
├ /hunter — Hunter menu
├ /hunt — Search for jobs
├ /jobs — View matched jobs
├ /profile — Your profile
└ /applications — Your applications

*🔧 GENERAL*
├ /start — Main menu
└ /help — This help message

━━━━━━━━━━━━━━━━━━━━━━
💡 *Quick Tips:*
• Start tracking with /track
• Create invoices from tracked time
• View dashboard for revenue stats
━━━━━━━━━━━━━━━━━━━━━━
🌐 Need more help? Visit ung.app/help`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Main Menu", "main_menu"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Dashboard", "action_reports"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}
