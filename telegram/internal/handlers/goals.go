package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/services"
)

// GoalsHandler handles income goal commands
type GoalsHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewGoalsHandler creates a new goals handler
func NewGoalsHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *GoalsHandler {
	return &GoalsHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleStatus shows goal progress
func (h *GoalsHandler) HandleStatus(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	status, err := h.apiClient.GetGoalStatus(user.APIToken)
	if err != nil {
		// No goal set
		msg := tgbotapi.NewMessage(chatID,
			"--------------------\n"+
				"      *Income Goals*\n"+
				"--------------------\n\n"+
				"No goal set for this month.\n\n"+
				"Set a goal to track your progress!\n\n"+
				"_Visit the web app to set a goal._")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Dashboard", "action_reports"),
				tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Goal Progress*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*Goal:* $%.2f\n", status.Goal.Amount))
	text.WriteString(fmt.Sprintf("*Current:* $%.2f\n", status.Current))
	text.WriteString(fmt.Sprintf("*Progress:* %.1f%%\n\n", status.Progress))

	// Progress bar
	text.WriteString(generateProgressBar(status.Progress))
	text.WriteString("\n\n")

	if status.Remaining > 0 {
		text.WriteString(fmt.Sprintf("*Remaining:* $%.2f\n", status.Remaining))
	} else {
		text.WriteString("*Goal Achieved!*\n")
	}

	if status.Goal.Description != "" {
		text.WriteString(fmt.Sprintf("\n_%s_", status.Goal.Description))
	}

	text.WriteString("\n\n--------------------\n")
	text.WriteString("_Updated just now_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Refresh", "goal_status"),
			tgbotapi.NewInlineKeyboardButtonData("Dashboard", "action_reports"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleRateAnalysis shows hourly rate analysis
func (h *GoalsHandler) HandleRateAnalysis(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	analysis, err := h.apiClient.GetRateAnalysis(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch rate analysis: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Rate Analysis*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*Average Rate:* $%.2f/hr\n", analysis.AverageRate))
	text.WriteString(fmt.Sprintf("*Effective Rate:* $%.2f/hr\n\n", analysis.EffectiveRate))
	text.WriteString(fmt.Sprintf("*Total Hours:* %.1f hrs\n", analysis.TotalHours))
	text.WriteString(fmt.Sprintf("*Total Revenue:* $%.2f\n\n", analysis.TotalRevenue))

	if analysis.SuggestedRate > 0 {
		text.WriteString(fmt.Sprintf("*Suggested Rate:* $%.2f/hr\n", analysis.SuggestedRate))
	}

	text.WriteString("\n--------------------\n")
	text.WriteString("_Based on your tracking data_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Goals", "goal_status"),
			tgbotapi.NewInlineKeyboardButtonData("Dashboard", "action_reports"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}
