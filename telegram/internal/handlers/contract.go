package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// ContractHandler handles contract-related commands
type ContractHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewContractHandler creates a new contract handler
func NewContractHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *ContractHandler {
	return &ContractHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleList shows list of contracts
func (h *ContractHandler) HandleList(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Fetch contracts
	contracts, err := h.apiClient.ListContracts(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to fetch contracts: "+err.Error())
	}

	if len(contracts) == 0 {
		msg := tgbotapi.NewMessage(chatID, "You don't have any contracts yet.\n\nUse /contract to create your first contract!")
		_, err := h.bot.Send(msg)
		return err
	}

	// Build contract list message
	var text strings.Builder
	text.WriteString("📋 *Your Contracts*\n\n")

	for i, contract := range contracts {
		text.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, contract.Name))
		text.WriteString(fmt.Sprintf("   Type: %s\n", contract.Type))
		text.WriteString(fmt.Sprintf("   Rate: $%.2f\n", contract.Rate))
		text.WriteString(fmt.Sprintf("   Status: %s\n", contract.Status))
		text.WriteString("\n")
	}

	text.WriteString("_Use /contract to create a new contract_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	_, err = h.bot.Send(msg)
	return err
}

// HandleCreate starts contract creation flow
func (h *ContractHandler) HandleCreate(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Fetch clients for selection
	clients, err := h.apiClient.ListClients(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to fetch clients: "+err.Error())
	}

	if len(clients) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			"You don't have any clients yet.\n\n"+
				"Please create a client first using /client")
		h.bot.Send(msg)
		return nil
	}

	// Show client selection
	return h.showClientSelection(chatID, telegramID, clients)
}

func (h *ContractHandler) showClientSelection(chatID int64, telegramID int64, clients []services.Client) error {
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, client := range clients {
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				client.Name,
				fmt.Sprintf("contract_client_%d", client.ID),
			),
		}
		buttons = append(buttons, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "Let's create a new contract!\n\nSelect a client for this contract:")
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)

	// Set initial state
	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateContractSelectClient),
		Data:       make(map[string]interface{}),
	})

	return err
}

// HandleClientSelected handles client selection callback
func (h *ContractHandler) HandleClientSelected(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID
	data := callbackQuery.Data // "contract_client_123"

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
	session.State = string(models.StateContractName)
	h.sessionMgr.SetSession(session)

	// Ask for contract name
	msg := tgbotapi.NewMessage(chatID,
		"Great! Now please enter the contract name:\n\n"+
			"Example: Monthly Retainer, Project Development, etc.")

	h.bot.Send(msg)

	// Answer callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Client selected")
	h.bot.Request(callback)

	return nil
}

// HandleNameInput handles contract name input
func (h *ContractHandler) HandleNameInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	name := strings.TrimSpace(message.Text)

	if name == "" {
		msg := tgbotapi.NewMessage(chatID, "Contract name cannot be empty. Please send a valid name:")
		h.bot.Send(msg)
		return nil
	}

	// Get session and store name
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /contract")
	}

	session.Data["name"] = name

	// Ask for contract type
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Hourly", "contract_type_hourly"),
			tgbotapi.NewInlineKeyboardButtonData("Fixed", "contract_type_fixed"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Monthly", "contract_type_monthly"),
			tgbotapi.NewInlineKeyboardButtonData("Project", "contract_type_project"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "What type of contract is this?")
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)

	// Update state
	session.State = string(models.StateContractType)
	h.sessionMgr.SetSession(session)

	return nil
}

// HandleTypeSelected handles contract type selection
func (h *ContractHandler) HandleTypeSelected(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID
	data := callbackQuery.Data // "contract_type_hourly"

	// Parse type
	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return fmt.Errorf("invalid callback data")
	}

	contractType := parts[2]

	// Update session
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /contract")
	}

	session.Data["type"] = contractType
	session.State = string(models.StateContractRate)
	h.sessionMgr.SetSession(session)

	// Ask for rate
	var ratePrompt string
	switch contractType {
	case "hourly":
		ratePrompt = "What is the hourly rate?\n\nExample: 150 or 150.50"
	case "monthly":
		ratePrompt = "What is the monthly rate?\n\nExample: 5000 or 5000.00"
	case "fixed":
		ratePrompt = "What is the fixed rate?\n\nExample: 10000 or 10000.00"
	case "project":
		ratePrompt = "What is the project rate?\n\nExample: 25000 or 25000.00"
	default:
		ratePrompt = "What is the rate?\n\nExample: 1000 or 1000.00"
	}

	msg := tgbotapi.NewMessage(chatID, ratePrompt)
	h.bot.Send(msg)

	// Answer callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Type selected")
	h.bot.Request(callback)

	return nil
}

// HandleRateInput handles contract rate input and creates the contract
func (h *ContractHandler) HandleRateInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	inputText := message.Text

	// Parse rate
	rate, err := strconv.ParseFloat(inputText, 64)
	if err != nil || rate <= 0 {
		msg := tgbotapi.NewMessage(chatID,
			"❌ Invalid rate. Please enter a number (e.g., 1000 or 1000.50)")
		h.bot.Send(msg)
		return nil
	}

	// Get session
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /contract")
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Extract data from session
	clientID, _ := session.Data["client_id"].(int)
	name, _ := session.Data["name"].(string)
	contractType, _ := session.Data["type"].(string)

	// Create contract via API
	contract, err := h.apiClient.CreateContract(user.APIToken, services.ContractCreateRequest{
		ClientID: uint(clientID),
		Name:     name,
		Type:     contractType,
		Rate:     rate,
	})

	if err != nil {
		return h.sendError(chatID, "Failed to create contract: "+err.Error())
	}

	// Success message
	var text strings.Builder
	text.WriteString("✅ *Contract created successfully!*\n\n")
	text.WriteString(fmt.Sprintf("*Name:* %s\n", contract.Name))
	text.WriteString(fmt.Sprintf("*Type:* %s\n", contract.Type))
	text.WriteString(fmt.Sprintf("*Rate:* $%.2f\n", contract.Rate))
	text.WriteString(fmt.Sprintf("*Status:* %s\n", contract.Status))
	text.WriteString("\n_You can now track time against this contract!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Clear session
	h.sessionMgr.ClearSession(telegramID)

	return nil
}

// Helper methods

func (h *ContractHandler) sendAuthRequired(chatID int64) error {
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

func (h *ContractHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
	_, err := h.bot.Send(msg)
	return err
}
