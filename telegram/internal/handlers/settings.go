package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/services"
)

// SettingsHandler handles settings commands
type SettingsHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *SettingsHandler {
	return &SettingsHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleShow shows current settings
func (h *SettingsHandler) HandleShow(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	settings, err := h.apiClient.GetSettings(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch settings: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Settings*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*Hours per Week:* %.0f\n", settings.HoursPerWeek))
	text.WriteString(fmt.Sprintf("*Weeks per Year:* %d\n", settings.WeeksPerYear))
	text.WriteString(fmt.Sprintf("*Default Tax:* %.0f%%\n", settings.DefaultTaxPercent))
	text.WriteString(fmt.Sprintf("*Default Margin:* %.0f%%\n", settings.DefaultMargin))
	text.WriteString(fmt.Sprintf("*Annual Expenses:* $%.2f\n", settings.AnnualExpenses))
	text.WriteString(fmt.Sprintf("*Default Currency:* %s\n\n", settings.DefaultCurrency))
	text.WriteString("--------------------\n")
	text.WriteString("_Modify settings in the web app_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Rate Analysis", "action_rate"),
			tgbotapi.NewInlineKeyboardButtonData("Goals", "goal_status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}
