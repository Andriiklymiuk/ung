package handlers

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/services"
)

// TrackingHandler handles time tracking commands
type TrackingHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewTrackingHandler creates a new tracking handler
func NewTrackingHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *TrackingHandler {
	return &TrackingHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleList shows list of tracking sessions
func (h *TrackingHandler) HandleList(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Fetch tracking sessions
	sessions, err := h.apiClient.ListTracking(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to fetch tracking sessions: "+err.Error())
	}

	if len(sessions) == 0 {
		msg := tgbotapi.NewMessage(chatID, "You don't have any tracking sessions yet.\n\nUse /track to start tracking time!")
		_, err := h.bot.Send(msg)
		return err
	}

	// Build tracking list message
	var text strings.Builder
	text.WriteString("⏱️ *Your Time Tracking Sessions*\n\n")

	totalDuration := 0.0
	for i, session := range sessions {
		if i >= 10 {
			text.WriteString(fmt.Sprintf("\n_...and %d more_", len(sessions)-10))
			break
		}

		statusEmoji := "✅"
		if session.Active {
			statusEmoji = "🔴 ACTIVE"
		}

		text.WriteString(fmt.Sprintf("%d. %s\n", i+1, statusEmoji))
		if session.Notes != "" {
			text.WriteString(fmt.Sprintf("   📝 %s\n", session.Notes))
		}
		text.WriteString(fmt.Sprintf("   ⏰ Started: %s\n", formatTime(session.StartTime)))
		if session.EndTime != "" {
			text.WriteString(fmt.Sprintf("   🏁 Ended: %s\n", formatTime(session.EndTime)))
		}
		if session.Duration > 0 {
			text.WriteString(fmt.Sprintf("   ⏳ Duration: %.2f hours\n", session.Duration))
			totalDuration += session.Duration
		}
		text.WriteString("\n")
	}

	if totalDuration > 0 {
		text.WriteString(fmt.Sprintf("*Total Time:* %.2f hours\n\n", totalDuration))
	}
	text.WriteString("_Use /track to start a new session_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	_, err = h.bot.Send(msg)
	return err
}

// HandleStart starts a new tracking session
func (h *TrackingHandler) HandleStart(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Check if there's already an active session
	sessions, err := h.apiClient.ListTracking(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to check active sessions: "+err.Error())
	}

	for _, session := range sessions {
		if session.Active {
			msg := tgbotapi.NewMessage(chatID,
				"⚠️ You already have an active tracking session!\n\n"+
					fmt.Sprintf("Started: %s\n", formatTime(session.StartTime))+
					fmt.Sprintf("Notes: %s\n\n", session.Notes)+
					"Use /stop to stop the current session first.")
			h.bot.Send(msg)
			return nil
		}
	}

	// Start tracking with default project ID (1) and no notes
	// In a real implementation, you might want to ask for project selection
	session, err := h.apiClient.StartTracking(user.APIToken, 1, "")
	if err != nil {
		return h.sendError(chatID, "Failed to start tracking: "+err.Error())
	}

	// Success message
	var text strings.Builder
	text.WriteString("✅ *Time tracking started!*\n\n")
	text.WriteString("⏱️ Timer is now running...\n")
	text.WriteString(fmt.Sprintf("🕐 Started at: %s\n\n", formatTime(session.StartTime)))
	text.WriteString("_Use /stop to stop tracking_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	return nil
}

// HandleStop stops the active tracking session
func (h *TrackingHandler) HandleStop(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Stop tracking
	session, err := h.apiClient.StopTracking(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to stop tracking: "+err.Error())
	}

	// Success message
	var text strings.Builder
	text.WriteString("✅ *Time tracking stopped!*\n\n")
	text.WriteString(fmt.Sprintf("🕐 Started: %s\n", formatTime(session.StartTime)))
	text.WriteString(fmt.Sprintf("🏁 Ended: %s\n", formatTime(session.EndTime)))
	if session.Duration > 0 {
		text.WriteString(fmt.Sprintf("⏳ Duration: %.2f hours\n", session.Duration))
		text.WriteString(fmt.Sprintf("💰 (~$%.2f at $100/hr)\n", session.Duration*100))
	}
	if session.Notes != "" {
		text.WriteString(fmt.Sprintf("\n📝 Notes: %s\n", session.Notes))
	}
	text.WriteString("\n_Your time has been recorded!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	return nil
}

// HandleActive shows the current active tracking session
func (h *TrackingHandler) HandleActive(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		return h.sendAuthRequired(chatID)
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Fetch tracking sessions
	sessions, err := h.apiClient.ListTracking(user.APIToken)
	if err != nil {
		return h.sendError(chatID, "Failed to fetch tracking sessions: "+err.Error())
	}

	// Find active session
	var activeSession *services.TrackingSession
	for i, session := range sessions {
		if session.Active {
			activeSession = &sessions[i]
			break
		}
	}

	if activeSession == nil {
		msg := tgbotapi.NewMessage(chatID,
			"No active tracking session.\n\n"+
				"Use /track to start tracking time!")
		h.bot.Send(msg)
		return nil
	}

	// Calculate elapsed time
	startTime, _ := time.Parse(time.RFC3339, activeSession.StartTime)
	elapsed := time.Since(startTime)
	hours := elapsed.Hours()

	// Show active session
	var text strings.Builder
	text.WriteString("🔴 *Active Time Tracking*\n\n")
	text.WriteString(fmt.Sprintf("🕐 Started: %s\n", formatTime(activeSession.StartTime)))
	text.WriteString(fmt.Sprintf("⏳ Elapsed: %.2f hours\n", hours))
	text.WriteString(fmt.Sprintf("💰 Estimated: $%.2f (at $100/hr)\n", hours*100))
	if activeSession.Notes != "" {
		text.WriteString(fmt.Sprintf("\n📝 Notes: %s\n", activeSession.Notes))
	}
	text.WriteString("\n_Use /stop to stop tracking_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏹️ Stop Tracking", "tracking_stop"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)

	return nil
}

// HandleStopCallback handles stop tracking callback
func (h *TrackingHandler) HandleStopCallback(callbackQuery *tgbotapi.CallbackQuery) error {
	// Convert callback to message for easier handling
	msg := &tgbotapi.Message{
		Chat: callbackQuery.Message.Chat,
		From: callbackQuery.From,
	}

	// Answer callback first
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Stopping tracking...")
	h.bot.Request(callback)

	// Stop tracking
	return h.HandleStop(msg)
}

// Helper methods

func (h *TrackingHandler) sendAuthRequired(chatID int64) error {
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

func (h *TrackingHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, "❌ "+errorMsg)
	_, err := h.bot.Send(msg)
	return err
}

func formatTime(timeStr string) string {
	if timeStr == "" {
		return "N/A"
	}

	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// Try alternative format
		t, err = time.Parse("2006-01-02T15:04:05", timeStr)
		if err != nil {
			return timeStr
		}
	}

	return t.Format("Jan 2, 2006 at 3:04 PM")
}
