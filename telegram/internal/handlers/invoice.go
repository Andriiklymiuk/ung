package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// InvoiceHandler handles invoice-related commands
type InvoiceHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *InvoiceHandler {
	return &InvoiceHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleCreate starts invoice creation flow
func (h *InvoiceHandler) HandleCreate(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Fetch clients
	clients, err := h.apiClient.ListClients(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to fetch clients: "+err.Error())
	}

	if len(clients) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			"You don't have any clients yet.\n\n"+
				"Let's create one first! Please send the client name:")
		h.bot.Send(msg)

		// Set state
		h.sessionMgr.SetSession(&models.Session{
			TelegramID: telegramID,
			State:      string(models.StateClientCreateName),
			Data:       make(map[string]interface{}),
		})
		return nil
	}

	// Show client selection
	return h.showClientSelection(chatID, telegramID, clients)
}

func (h *InvoiceHandler) showClientSelection(chatID int64, telegramID int64, clients []services.Client) error {
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, client := range clients {
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				client.Name,
				fmt.Sprintf("invoice_client_%d", client.ID),
			),
		}
		buttons = append(buttons, row)
	}

	// Add "Create new client" option
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("➕ Create new client", "invoice_new_client"),
	})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "Select a client for this invoice:")
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)

	// Set initial state
	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateInvoiceSelectClient),
		Data:       make(map[string]interface{}),
	})

	return err
}

// HandleClientSelected handles client selection callback
func (h *InvoiceHandler) HandleClientSelected(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID
	data := callbackQuery.Data // "invoice_client_123"

	// Parse client ID
	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return fmt.Errorf("invalid callback data")
	}

	clientID, err := strconv.Atoi(parts[2])
	if err != nil {
		return err
	}

	// Update session
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		session = &models.Session{
			TelegramID: telegramID,
			Data:       make(map[string]interface{}),
		}
	}

	session.Data["client_id"] = clientID
	session.State = string(models.StateInvoiceAmount)
	h.sessionMgr.SetSession(session)

	// Ask for amount
	msg := tgbotapi.NewMessage(chatID,
		"Great! Now please enter the invoice amount:\n\n"+
			"Example: 1500 or 1500.50")

	h.bot.Send(msg)

	// Answer callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Client selected")
	h.bot.Request(callback)

	return nil
}

// HandleAmountInput handles amount input
func (h *InvoiceHandler) HandleAmountInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	text := message.Text

	// Parse amount
	amount, err := strconv.ParseFloat(text, 64)
	if err != nil || amount <= 0 {
		msg := tgbotapi.NewMessage(chatID,
			"❌ Invalid amount. Please enter a number (e.g., 1500 or 1500.50)")
		h.bot.Send(msg)
		return nil
	}

	// Update session
	session := h.sessionMgr.GetSession(telegramID)
	session.Data["amount"] = amount
	session.State = string(models.StateInvoiceDescription)
	h.sessionMgr.SetSession(session)

	// Ask for description
	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("Amount: $%.2f\n\nNow add a description or notes for this invoice:", amount))

	h.bot.Send(msg)
	return nil
}

// HandleDescriptionInput handles description input
func (h *InvoiceHandler) HandleDescriptionInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	description := message.Text

	// Update session
	session := h.sessionMgr.GetSession(telegramID)
	session.Data["description"] = description
	session.State = string(models.StateInvoiceDueDate)
	h.sessionMgr.SetSession(session)

	// Ask for due date with quick options
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

// CreateInvoice creates the invoice
func (h *InvoiceHandler) CreateInvoice(callbackQuery *tgbotapi.CallbackQuery, days int) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	// Get user and session
	user := h.sessionMgr.GetUser(telegramID)
	session := h.sessionMgr.GetSession(telegramID)

	// Build invoice data
	dueDate := time.Now().AddDate(0, 0, days)

	invoiceData := map[string]interface{}{
		"client_id":   session.Data["client_id"],
		"amount":      session.Data["amount"],
		"currency":    "USD",
		"description": session.Data["description"],
		"due_date":    dueDate.Format("2006-01-02"),
	}

	// Call API
	invoice, err := h.apiClient.CreateInvoice(user.APIToken, invoiceData)
	if err != nil {
		return h.sendError(chatID, "Failed to create invoice: "+err.Error())
	}

	// Clear session
	h.sessionMgr.ClearSession(telegramID)

	// Send success message
	text := fmt.Sprintf(
		"✅ *Invoice created successfully!*\n\n"+
			"Invoice #%s\n"+
			"Amount: $%.2f %s\n"+
			"Due Date: %s\n\n"+
			"What would you like to do next?",
		invoice.InvoiceNum,
		invoice.Amount,
		invoice.Currency,
		dueDate.Format("January 2, 2006"),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 View PDF", fmt.Sprintf("invoice_pdf_%d", invoice.ID)),
			tgbotapi.NewInlineKeyboardButtonData("📧 Send Email", fmt.Sprintf("invoice_email_%d", invoice.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Create Another", "action_invoice"),
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

// HandleList lists all invoices
func (h *InvoiceHandler) HandleList(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	invoices, err := h.apiClient.ListInvoices(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to fetch invoices: "+err.Error())
	}

	if len(invoices) == 0 {
		msg := tgbotapi.NewMessage(chatID, "You don't have any invoices yet.\n\nCreate one with /invoice")
		h.bot.Send(msg)
		return nil
	}

	text := "*Your Invoices:*\n\n"
	for i, inv := range invoices {
		if i >= 10 {
			text += fmt.Sprintf("\n_...and %d more_", len(invoices)-10)
			break
		}
		statusEmoji := "⏳"
		switch inv.Status {
		case "paid":
			statusEmoji = "✅"
		case "overdue":
			statusEmoji = "⚠️"
		}
		text += fmt.Sprintf("%s *%s* - $%.2f - %s\n", statusEmoji, inv.InvoiceNum, inv.Amount, inv.Status)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	_, err = h.bot.Send(msg)
	return err
}

func (h *InvoiceHandler) sendAuthRequired(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "🔒 Please authenticate first with /start")
	_, err := h.bot.Send(msg)
	return err
}

func (h *InvoiceHandler) sendError(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, "❌ "+message)
	_, err := h.bot.Send(msg)
	return err
}
