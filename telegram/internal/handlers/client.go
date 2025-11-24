package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// ClientHandler handles client-related commands
type ClientHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewClientHandler creates a new client handler
func NewClientHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *ClientHandler {
	return &ClientHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleList shows list of clients
func (h *ClientHandler) HandleList(message *tgbotapi.Message) error {
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
		msg := tgbotapi.NewMessage(chatID, "You don't have any clients yet.\n\nUse /client to create your first client!")
		_, err := h.bot.Send(msg)
		return err
	}

	// Build client list message
	var text strings.Builder
	text.WriteString("👥 *Your Clients*\n\n")

	for i, client := range clients {
		text.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, client.Name))
		if client.Email != "" {
			text.WriteString(fmt.Sprintf("   📧 %s\n", client.Email))
		}
		if client.TaxID != "" {
			text.WriteString(fmt.Sprintf("   🏢 Tax ID: %s\n", client.TaxID))
		}
		text.WriteString("\n")
	}

	text.WriteString("_Use /client to create a new client_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	_, err = h.bot.Send(msg)
	return err
}

// HandleCreate starts client creation flow
func (h *ClientHandler) HandleCreate(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	// Start client creation flow
	msg := tgbotapi.NewMessage(chatID, "Let's create a new client!\n\nPlease send the client's name:")
	h.bot.Send(msg)

	// Set state
	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateClientCreateName),
		Data:       make(map[string]interface{}),
	})

	return nil
}

// HandleNameInput handles client name input
func (h *ClientHandler) HandleNameInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	name := strings.TrimSpace(message.Text)

	if name == "" {
		msg := tgbotapi.NewMessage(chatID, "Client name cannot be empty. Please send a valid name:")
		h.bot.Send(msg)
		return nil
	}

	// Get session and store name
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /client")
	}

	session.Data["name"] = name

	// Ask for email
	msg := tgbotapi.NewMessage(chatID, "Great! Now please send the client's email address:")
	h.bot.Send(msg)

	// Update state
	session.State = string(models.StateClientCreateEmail)
	h.sessionMgr.SetSession(session)

	return nil
}

// HandleEmailInput handles client email input
func (h *ClientHandler) HandleEmailInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	email := strings.TrimSpace(message.Text)

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		msg := tgbotapi.NewMessage(chatID, "Please send a valid email address:")
		h.bot.Send(msg)
		return nil
	}

	// Get session and store email
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /client")
	}

	session.Data["email"] = email

	// Ask for address (optional)
	msg := tgbotapi.NewMessage(chatID,
		"Perfect! Now please send the client's address:\n\n"+
			"_Send /skip if you want to skip this step_")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Update state
	session.State = string(models.StateClientCreateAddress)
	h.sessionMgr.SetSession(session)

	return nil
}

// HandleAddressInput handles client address input
func (h *ClientHandler) HandleAddressInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Get session
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /client")
	}

	// Check if skipping
	address := strings.TrimSpace(message.Text)
	if message.IsCommand() && message.Command() == "skip" {
		address = ""
	}

	if address != "" {
		session.Data["address"] = address
	}

	// Ask for tax ID (optional)
	msg := tgbotapi.NewMessage(chatID,
		"Almost done! Please send the client's tax ID or company registration number:\n\n"+
			"_Send /skip if you want to skip this step_")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Update state
	session.State = string(models.StateClientCreateTaxID)
	h.sessionMgr.SetSession(session)

	return nil
}

// HandleTaxIDInput handles client tax ID input and creates the client
func (h *ClientHandler) HandleTaxIDInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Get session
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /client")
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Check if skipping
	taxID := strings.TrimSpace(message.Text)
	if message.IsCommand() && message.Command() == "skip" {
		taxID = ""
	}

	if taxID != "" {
		session.Data["tax_id"] = taxID
	}

	// Create client via API
	name, _ := session.Data["name"].(string)
	email, _ := session.Data["email"].(string)
	address, _ := session.Data["address"].(string)
	taxIDStr, _ := session.Data["tax_id"].(string)

	client, err := h.apiClient.CreateClient(user.APIToken, services.ClientCreateRequest{
		Name:    name,
		Email:   email,
		Address: address,
		TaxID:   taxIDStr,
	})

	if err != nil {
		return h.sendError(chatID, "Failed to create client: "+err.Error())
	}

	// Success message
	var text strings.Builder
	text.WriteString("✅ *Client created successfully!*\n\n")
	text.WriteString(fmt.Sprintf("*Name:* %s\n", client.Name))
	text.WriteString(fmt.Sprintf("*Email:* %s\n", client.Email))
	if client.Address != "" {
		text.WriteString(fmt.Sprintf("*Address:* %s\n", client.Address))
	}
	if client.TaxID != "" {
		text.WriteString(fmt.Sprintf("*Tax ID:* %s\n", client.TaxID))
	}
	text.WriteString("\n_You can now create invoices for this client!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Clear session
	h.sessionMgr.ClearSession(telegramID)

	return nil
}

// Helper methods

func (h *ClientHandler) sendAuthRequired(chatID int64) error {
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

func (h *ClientHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
	_, err := h.bot.Send(msg)
	return err
}
