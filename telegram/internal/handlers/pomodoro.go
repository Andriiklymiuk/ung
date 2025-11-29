package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/services"
)

// PomodoroHandler handles pomodoro timer commands
type PomodoroHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewPomodoroHandler creates a new pomodoro handler
func NewPomodoroHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *PomodoroHandler {
	return &PomodoroHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleStart starts a pomodoro session
func (h *PomodoroHandler) HandleStart(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	// Show duration selection
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("15 min", "pomodoro_start_15"),
			tgbotapi.NewInlineKeyboardButtonData("25 min", "pomodoro_start_25"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("45 min", "pomodoro_start_45"),
			tgbotapi.NewInlineKeyboardButtonData("60 min", "pomodoro_start_60"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID,
		"--------------------\n"+
			"      *Pomodoro Timer*\n"+
			"--------------------\n\n"+
			"Select session duration:\n\n"+
			"_Focus on one task at a time._")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleStartDuration starts a pomodoro with specific duration
func (h *PomodoroHandler) HandleStartDuration(callbackQuery *tgbotapi.CallbackQuery, duration int) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	user := h.sessionMgr.GetUser(telegramID)

	result, err := h.apiClient.StartPomodoro(user.APIToken, duration, "Focus Session")
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to start pomodoro: "+err.Error())
		h.bot.Send(msg)

		callback := tgbotapi.NewCallback(callbackQuery.ID, "Error")
		h.bot.Request(callback)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Pomodoro Started!*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*Duration:* %d minutes\n\n", duration))
	text.WriteString("Stay focused! You've got this.\n\n")
	text.WriteString("_Timer is running..._")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	// Get session ID from result
	var sessionID float64
	if id, ok := result["id"].(float64); ok {
		sessionID = id
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Check Status", "pomodoro_active"),
			tgbotapi.NewInlineKeyboardButtonData("Stop", fmt.Sprintf("pomodoro_stop_%.0f", sessionID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Pomodoro started!")
	h.bot.Request(callback)

	return nil
}

// HandleActive shows active pomodoro session
func (h *PomodoroHandler) HandleActive(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	result, err := h.apiClient.GetActivePomodoro(user.APIToken)
	if err != nil || result == nil || len(result) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			"--------------------\n"+
				"      *No Active Pomodoro*\n"+
				"--------------------\n\n"+
				"Start a new session with /pomodoro")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Start Pomodoro", "action_pomodoro"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	// Extract values from result
	remaining := 0.0
	progress := 0.0
	if r, ok := result["remaining_seconds"].(float64); ok {
		remaining = r
	}
	if p, ok := result["progress_percent"].(float64); ok {
		progress = p
	}

	minutes := int(remaining) / 60
	seconds := int(remaining) % 60

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Active Pomodoro*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*Remaining:* %d:%02d\n", minutes, seconds))
	text.WriteString(fmt.Sprintf("*Progress:* %.0f%%\n\n", progress))
	text.WriteString(generateProgressBar(progress))
	text.WriteString("\n\n_Keep going!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Refresh", "pomodoro_active"),
			tgbotapi.NewInlineKeyboardButtonData("Stop", "pomodoro_stop_active"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleStats shows pomodoro statistics
func (h *PomodoroHandler) HandleStats(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	stats, err := h.apiClient.GetPomodoroStats(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch pomodoro stats: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Pomodoro Stats*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*Today:* %d pomodoros (%d min)\n", stats.TodayCompleted, stats.TodayMinutes))
	text.WriteString(fmt.Sprintf("*This Week:* %d pomodoros\n", stats.WeekCompleted))
	text.WriteString(fmt.Sprintf("*This Month:* %d pomodoros\n\n", stats.MonthCompleted))
	text.WriteString(fmt.Sprintf("*Current Streak:* %d days\n", stats.CurrentStreak))
	text.WriteString(fmt.Sprintf("*Daily Average (30d):* %.1f\n\n", stats.AvgDaily30d))
	text.WriteString(fmt.Sprintf("*Total Completed:* %d\n\n", stats.TotalCompleted))
	text.WriteString("--------------------\n")
	text.WriteString("_Keep up the great work!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Start Pomodoro", "action_pomodoro"),
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}
