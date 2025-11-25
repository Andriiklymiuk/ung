package handlers

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// CompanyHandler handles company-related operations
type CompanyHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewCompanyHandler creates a new company handler
func NewCompanyHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *CompanyHandler {
	return &CompanyHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleList lists all companies
func (h *CompanyHandler) HandleList(message *tgbotapi.Message) error {
	telegramID := message.From.ID

	// Check if user is authenticated
	user := h.sessionMgr.GetUser(telegramID)
	if user == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please /start first to authenticate.")
		_, err := h.bot.Send(msg)
		return err
	}

	// Fetch companies from API
	companies, err := h.apiClient.ListCompanies(user.APIToken)
	if err != nil {
		log.Printf("Error fetching companies: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to fetch companies. Please try again.")
		h.bot.Send(msg)
		return err
	}

	if len(companies) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "📋 No companies found.\n\nUse /company to create one.")
		_, err := h.bot.Send(msg)
		return err
	}

	// Build response
	var text strings.Builder
	text.WriteString("🏢 *Your Companies*\n\n")

	for i, company := range companies {
		text.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, company.Name))
		if company.Email != "" {
			text.WriteString(fmt.Sprintf("   📧 %s\n", company.Email))
		}
		if company.Phone != "" {
			text.WriteString(fmt.Sprintf("   📞 %s\n", company.Phone))
		}
		if company.TaxID != "" {
			text.WriteString(fmt.Sprintf("   🆔 Tax ID: %s\n", company.TaxID))
		}
		text.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = "Markdown"
	_, err = h.bot.Send(msg)
	return err
}

// HandleCreate starts company creation flow
func (h *CompanyHandler) HandleCreate(message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Check if user is authenticated
	user := h.sessionMgr.GetUser(telegramID)
	if user == nil {
		msg := tgbotapi.NewMessage(chatID, "Please /start first to authenticate.")
		_, err := h.bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(chatID, "🏢 Let's create a new company!\n\nPlease send the company name:")
	h.bot.Send(msg)

	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateCompanyCreateName),
		Data:       make(map[string]interface{}),
	})

	return nil
}

// HandleNameInput handles company name input
func (h *CompanyHandler) HandleNameInput(message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	name := strings.TrimSpace(message.Text)
	if name == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ Company name cannot be empty. Please try again:")
		h.bot.Send(msg)
		return nil
	}

	session := h.sessionMgr.GetSession(telegramID)
	session.Data["name"] = name
	session.State = string(models.StateCompanyCreateEmail)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "📧 Great! Now send the company email:")
	h.bot.Send(msg)

	return nil
}

// HandleEmailInput handles company email input
func (h *CompanyHandler) HandleEmailInput(message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	email := strings.TrimSpace(message.Text)
	if email == "" || !strings.Contains(email, "@") {
		msg := tgbotapi.NewMessage(chatID, "❌ Please enter a valid email address:")
		h.bot.Send(msg)
		return nil
	}

	session := h.sessionMgr.GetSession(telegramID)
	session.Data["email"] = email
	session.State = string(models.StateCompanyCreatePhone)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "📞 Now send the phone number (or /skip):")
	h.bot.Send(msg)

	return nil
}

// HandlePhoneInput handles company phone input
func (h *CompanyHandler) HandlePhoneInput(message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	phone := strings.TrimSpace(message.Text)

	// Check for skip
	if strings.ToLower(phone) == "/skip" {
		phone = ""
	}

	session := h.sessionMgr.GetSession(telegramID)
	session.Data["phone"] = phone
	session.State = string(models.StateCompanyCreateAddress)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "📍 Send the company address (or /skip):")
	h.bot.Send(msg)

	return nil
}

// HandleAddressInput handles company address input
func (h *CompanyHandler) HandleAddressInput(message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	address := strings.TrimSpace(message.Text)

	// Check for skip
	if strings.ToLower(address) == "/skip" {
		address = ""
	}

	session := h.sessionMgr.GetSession(telegramID)
	session.Data["address"] = address
	session.State = string(models.StateCompanyCreateTaxID)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "🆔 Finally, send the Tax ID (or /skip):")
	h.bot.Send(msg)

	return nil
}

// HandleTaxIDInput handles company tax ID input and creates the company
func (h *CompanyHandler) HandleTaxIDInput(message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	taxID := strings.TrimSpace(message.Text)

	// Check for skip
	if strings.ToLower(taxID) == "/skip" {
		taxID = ""
	}

	session := h.sessionMgr.GetSession(telegramID)
	user := h.sessionMgr.GetUser(telegramID)

	if user == nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Session expired. Please /start again.")
		h.bot.Send(msg)
		h.sessionMgr.ClearSession(telegramID)
		return nil
	}

	// Extract data from session
	name, _ := session.Data["name"].(string)
	email, _ := session.Data["email"].(string)
	phone, _ := session.Data["phone"].(string)
	address, _ := session.Data["address"].(string)

	// Create company via API
	company, err := h.apiClient.CreateCompany(user.APIToken, services.CompanyCreateRequest{
		Name:    name,
		Email:   email,
		Phone:   phone,
		Address: address,
		TaxID:   taxID,
	})

	if err != nil {
		log.Printf("Error creating company: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Failed to create company. Please try again.")
		h.bot.Send(msg)
		return err
	}

	// Clear session
	h.sessionMgr.ClearSession(telegramID)

	// Send success message
	var text strings.Builder
	text.WriteString("✅ *Company created successfully!*\n\n")
	text.WriteString(fmt.Sprintf("🏢 *%s*\n", company.Name))
	text.WriteString(fmt.Sprintf("📧 %s\n", company.Email))
	if company.Phone != "" {
		text.WriteString(fmt.Sprintf("📞 %s\n", company.Phone))
	}
	if company.Address != "" {
		text.WriteString(fmt.Sprintf("📍 %s\n", company.Address))
	}
	if company.TaxID != "" {
		text.WriteString(fmt.Sprintf("🆔 Tax ID: %s\n", company.TaxID))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	return nil
}
