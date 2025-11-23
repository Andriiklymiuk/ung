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
	text := `*UNG Bot - Available Commands*

*Invoices:*
/invoice - Create new invoice
/invoices - List all invoices
/unpaid - View unpaid invoices

*Clients:*
/client - Manage clients
/clients - List all clients

*Time Tracking:*
/track - Log time
/today - Today's tracked time
/week - This week's summary

*Reports:*
/report - View reports
/revenue - Revenue summary
/overdue - Overdue invoices

*Other:*
/status - Account status
/settings - Bot settings
/help - Show this help

*Quick Actions:*
Just send me a message like:
• "3h website design" - Log time quickly
• "invoice for Acme Corp" - Create invoice

Need help? Visit https://ung.app/help`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"

	_, err := h.bot.Send(msg)
	return err
}
