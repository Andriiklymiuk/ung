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
/invoices - List invoices with PDF buttons
/pdf <num> - Generate invoice PDF (e.g., /pdf INV-001)

*Clients & Companies:*
/client - Add new client
/clients - List all clients
/company - Add new company
/companies - List all companies

*Contracts:*
/contract - Create new contract
/contracts - List all contracts

*Time Tracking:*
/track - Start time tracking
/stop - Stop current session
/active - View active session
/log - Log time manually
/tracking - View tracking history

*Expenses:*
/expense - Add new expense
/expenses - List all expenses

*Reports:*
/dashboard - Revenue overview

*Getting Started:*
/start - Main menu & authentication
/help - Show this help

Need help? Visit https://ung.app/help`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = "Markdown"

	_, err := h.bot.Send(msg)
	return err
}
