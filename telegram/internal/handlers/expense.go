package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// ExpenseHandler handles expense-related commands
type ExpenseHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewExpenseHandler creates a new expense handler
func NewExpenseHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *ExpenseHandler {
	return &ExpenseHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleList shows list of expenses
func (h *ExpenseHandler) HandleList(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Fetch expenses
	expenses, err := h.apiClient.ListExpenses(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to fetch expenses: "+err.Error())
	}

	if len(expenses) == 0 {
		msg := tgbotapi.NewMessage(chatID, "You don't have any expenses yet.\n\nUse /expense to create your first expense!")
		_, err := h.bot.Send(msg)
		return err
	}

	// Build expense list message
	var text strings.Builder
	text.WriteString("💰 *Your Expenses*\n\n")

	totalAmount := 0.0
	for i, expense := range expenses {
		if i >= 15 {
			text.WriteString(fmt.Sprintf("\n_...and %d more_", len(expenses)-15))
			break
		}
		text.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, expense.Description))
		text.WriteString(fmt.Sprintf("   💵 $%.2f\n", expense.Amount))
		text.WriteString(fmt.Sprintf("   📂 %s\n", expense.Category))
		if expense.Vendor != "" {
			text.WriteString(fmt.Sprintf("   🏪 %s\n", expense.Vendor))
		}
		if expense.Date != "" {
			text.WriteString(fmt.Sprintf("   📅 %s\n", expense.Date))
		}
		text.WriteString("\n")
		totalAmount += expense.Amount
	}

	text.WriteString(fmt.Sprintf("*Total:* $%.2f\n\n", totalAmount))
	text.WriteString("_Use /expense to create a new expense_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	_, err = h.bot.Send(msg)
	return err
}

// HandleCreate starts expense creation flow
func (h *ExpenseHandler) HandleCreate(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	// Start expense creation flow
	msg := tgbotapi.NewMessage(chatID, "Let's create a new expense!\n\nPlease send a description for this expense:")
	h.bot.Send(msg)

	// Set state
	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateExpenseDescription),
		Data:       make(map[string]interface{}),
	})

	return nil
}

// HandleDescriptionInput handles expense description input
func (h *ExpenseHandler) HandleDescriptionInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	description := strings.TrimSpace(message.Text)

	if description == "" {
		msg := tgbotapi.NewMessage(chatID, "Expense description cannot be empty. Please send a valid description:")
		h.bot.Send(msg)
		return nil
	}

	// Get session and store description
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /expense")
	}

	session.Data["description"] = description

	// Ask for amount
	msg := tgbotapi.NewMessage(chatID, "Great! Now please enter the expense amount:\n\nExample: 50 or 50.25")
	h.bot.Send(msg)

	// Update state
	session.State = string(models.StateExpenseAmount)
	h.sessionMgr.SetSession(session)

	return nil
}

// HandleAmountInput handles expense amount input
func (h *ExpenseHandler) HandleAmountInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	text := message.Text

	// Parse amount
	amount, err := strconv.ParseFloat(text, 64)
	if err != nil || amount <= 0 {
		msg := tgbotapi.NewMessage(chatID,
			"❌ Invalid amount. Please enter a number (e.g., 50 or 50.25)")
		h.bot.Send(msg)
		return nil
	}

	// Get session and store amount
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /expense")
	}

	session.Data["amount"] = amount

	// Ask for category
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🍔 Meals", "expense_category_meals"),
			tgbotapi.NewInlineKeyboardButtonData("🚗 Travel", "expense_category_travel"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏢 Office", "expense_category_office"),
			tgbotapi.NewInlineKeyboardButtonData("💻 Equipment", "expense_category_equipment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Software", "expense_category_software"),
			tgbotapi.NewInlineKeyboardButtonData("📚 Education", "expense_category_education"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Marketing", "expense_category_marketing"),
			tgbotapi.NewInlineKeyboardButtonData("📦 Other", "expense_category_other"),
		),
	)

	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("Amount: $%.2f\n\nNow select a category:", amount))
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)

	// Update state
	session.State = string(models.StateExpenseCategory)
	h.sessionMgr.SetSession(session)

	return nil
}

// HandleCategorySelected handles expense category selection
func (h *ExpenseHandler) HandleCategorySelected(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID
	data := callbackQuery.Data // "expense_category_meals"

	// Parse category
	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return fmt.Errorf("invalid callback data")
	}

	category := parts[2]

	// Update session
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /expense")
	}

	session.Data["category"] = category
	session.State = string(models.StateExpenseVendor)
	h.sessionMgr.SetSession(session)

	// Ask for vendor (optional)
	msg := tgbotapi.NewMessage(chatID,
		"Perfect! Now please send the vendor name (or /skip if you want to skip this step):")
	h.bot.Send(msg)

	// Answer callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Category selected")
	h.bot.Request(callback)

	return nil
}

// HandleVendorInput handles expense vendor input and creates the expense
func (h *ExpenseHandler) HandleVendorInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Get session
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /expense")
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Check if skipping
	vendor := strings.TrimSpace(message.Text)
	if message.IsCommand() && message.Command() == "skip" {
		vendor = ""
	}

	// Extract data from session
	description, _ := session.Data["description"].(string)
	amount, _ := session.Data["amount"].(float64)
	category, _ := session.Data["category"].(string)

	// Create expense via API
	expense, err := h.apiClient.CreateExpense(user.APIToken, services.ExpenseCreateRequest{
		Description: description,
		Amount:      amount,
		Category:    category,
		Vendor:      vendor,
	})

	if err != nil {
		return h.sendError(chatID, "Failed to create expense: "+err.Error())
	}

	// Success message
	var text strings.Builder
	text.WriteString("✅ *Expense created successfully!*\n\n")
	text.WriteString(fmt.Sprintf("*Description:* %s\n", expense.Description))
	text.WriteString(fmt.Sprintf("*Amount:* $%.2f\n", expense.Amount))
	text.WriteString(fmt.Sprintf("*Category:* %s\n", expense.Category))
	if expense.Vendor != "" {
		text.WriteString(fmt.Sprintf("*Vendor:* %s\n", expense.Vendor))
	}
	if expense.Date != "" {
		text.WriteString(fmt.Sprintf("*Date:* %s\n", expense.Date))
	}
	text.WriteString("\n_Your expense has been recorded!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Clear session
	h.sessionMgr.ClearSession(telegramID)

	return nil
}

// Helper methods

func (h *ExpenseHandler) sendAuthRequired(chatID int64) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔐 Login", "auth_login"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "You need to authenticate first to use this feature.")
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

func (h *ExpenseHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
	_, err := h.bot.Send(msg)
	return err
}
