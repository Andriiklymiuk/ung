package handlers

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
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
		text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
			"      ⏱️ *Time Tracking*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n\n" +
			"📭 No sessions recorded yet!\n\n" +
			"Start tracking your work time."

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("▶️ Start Tracking", "action_track"),
			),
		)
		msg.ReplyMarkup = keyboard

		_, err := h.bot.Send(msg)
		return err
	}

	// Build tracking list message
	var text strings.Builder
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString("      ⏱️ *Time Tracking*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Calculate totals
	totalDuration := 0.0
	activeCount := 0
	for _, session := range sessions {
		if session.Duration > 0 {
			totalDuration += session.Duration
		}
		if session.Active {
			activeCount++
		}
	}

	// Summary
	text.WriteString(fmt.Sprintf("📊 *Summary*\n"))
	text.WriteString(fmt.Sprintf("├ Sessions: %d\n", len(sessions)))
	text.WriteString(fmt.Sprintf("├ Total: *%.1f hours*\n", totalDuration))
	text.WriteString(fmt.Sprintf("└ Est. Value: *$%.2f*\n", totalDuration*100))
	if activeCount > 0 {
		text.WriteString(fmt.Sprintf("\n🔴 *%d active session(s)*\n", activeCount))
	}
	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━\n\n")

	text.WriteString("📋 *Recent Sessions*\n\n")

	for i, session := range sessions {
		if i >= 8 {
			text.WriteString(fmt.Sprintf("\n_+%d more sessions_", len(sessions)-8))
			break
		}

		if session.Active {
			text.WriteString("🔴 *ACTIVE SESSION*\n")
		} else {
			text.WriteString(fmt.Sprintf("✅ *Session %d*\n", i+1))
		}

		if session.Notes != "" {
			text.WriteString(fmt.Sprintf("   📝 %s\n", session.Notes))
		}
		text.WriteString(fmt.Sprintf("   🕐 %s\n", formatTime(session.StartTime)))
		if session.Duration > 0 {
			text.WriteString(fmt.Sprintf("   ⏳ %.1fh · $%.2f\n", session.Duration, session.Duration*100))
		}
		text.WriteString("\n")
	}

	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Start", "action_track"),
			tgbotapi.NewInlineKeyboardButtonData("📝 Log Time", "action_log"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

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
		text := "━━━━━━━━━━━━━━━━━━━━━━\n" +
			"      ⏱️ *Active Session*\n" +
			"━━━━━━━━━━━━━━━━━━━━━━\n\n" +
			"⚪ No active session\n\n" +
			"Start tracking your work time!"

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("▶️ Start Tracking", "action_track"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Menu", "main_menu"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	// Calculate elapsed time
	startTime, _ := time.Parse(time.RFC3339, activeSession.StartTime)
	elapsed := time.Since(startTime)
	hours := elapsed.Hours()
	minutes := int(elapsed.Minutes()) % 60
	secs := int(elapsed.Seconds()) % 60

	// Show active session with visual timer
	var text strings.Builder
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString("   🔴 *TRACKING ACTIVE*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Big timer display
	text.WriteString("⏱️ *Elapsed Time*\n")
	text.WriteString(fmt.Sprintf("┌─────────────────────┐\n"))
	text.WriteString(fmt.Sprintf("│  *%02d:%02d:%02d*  │\n", int(hours), minutes, secs))
	text.WriteString(fmt.Sprintf("└─────────────────────┘\n\n"))

	// Details
	text.WriteString("📋 *Session Details*\n\n")
	text.WriteString(fmt.Sprintf("🕐 Started: %s\n", formatTime(activeSession.StartTime)))
	text.WriteString(fmt.Sprintf("⏳ Hours: *%.2f*\n", hours))
	text.WriteString(fmt.Sprintf("💰 Est. Value: *$%.2f*\n", hours*100))

	if activeSession.Notes != "" {
		text.WriteString(fmt.Sprintf("\n📝 Notes: _%s_\n", activeSession.Notes))
	}

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString("💡 _Tap Stop when done_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏹️ Stop Tracking", "tracking_stop"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "action_active"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu", "main_menu"),
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

// HandleLog starts manual time logging flow
func (h *TrackingHandler) HandleLog(message *tgbotapi.Message) error {
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
		msg := tgbotapi.NewMessage(chatID, "You don't have any contracts yet.\n\nCreate a contract first with /contract")
		h.bot.Send(msg)
		return nil
	}

	// Build inline keyboard with contracts
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, contract := range contracts {
		contractLabel := contract.Name
		if contract.Rate > 0 {
			contractLabel = fmt.Sprintf("%s ($%.0f/hr)", contract.Name, contract.Rate)
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			contractLabel,
			fmt.Sprintf("log_contract_%d", contract.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, "⏱️ *Log Time Manually*\n\nSelect the contract you worked on:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)

	return nil
}

// HandleContractSelected handles contract selection for manual time logging
func (h *TrackingHandler) HandleContractSelected(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID
	data := callbackQuery.Data

	// Extract contract ID
	var contractID uint
	fmt.Sscanf(data, "log_contract_%d", &contractID)

	// Answer callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Contract selected")
	h.bot.Request(callback)

	// Store contract ID and move to hours input
	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateTrackLogHours),
		Data: map[string]interface{}{
			"contract_id": float64(contractID),
		},
	})

	msg := tgbotapi.NewMessage(chatID, "How many hours did you work?\n\n_Example: 2.5 or 8_")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	return nil
}

// HandleHoursInput handles hours input
func (h *TrackingHandler) HandleHoursInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /log")
	}

	// Parse hours
	var hours float64
	_, err := fmt.Sscanf(message.Text, "%f", &hours)
	if err != nil || hours <= 0 || hours > 24 {
		msg := tgbotapi.NewMessage(chatID, "❌ Invalid hours. Please enter a positive number (e.g., 2.5)")
		h.bot.Send(msg)
		return nil
	}

	// Store hours
	session.Data["hours"] = hours
	session.State = string(models.StateTrackLogProject)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "What project/task were you working on?\n\n_Type /skip if you don't want to add a project name_")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	return nil
}

// HandleProjectInput handles project name input
func (h *TrackingHandler) HandleProjectInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /log")
	}

	projectName := ""
	if message.Text != "/skip" {
		projectName = message.Text
	}

	// Store project
	session.Data["project_name"] = projectName
	session.State = string(models.StateTrackLogNotes)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "Any notes about this work?\n\n_Type /skip to finish_")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	return nil
}

// HandleNotesInput handles notes input and creates the time log
func (h *TrackingHandler) HandleNotesInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		return h.sendError(chatID, "Session expired. Please start again with /log")
	}

	user := h.sessionMgr.GetUser(telegramID)

	notes := ""
	if message.Text != "/skip" {
		notes = message.Text
	}

	// Extract data from session
	contractID := uint(session.Data["contract_id"].(float64))
	hours := session.Data["hours"].(float64)
	projectName := ""
	if pn, ok := session.Data["project_name"].(string); ok {
		projectName = pn
	}

	// Create time log
	req := services.TrackingCreateRequest{
		ContractID:  contractID,
		ProjectName: projectName,
		Hours:       hours,
		Notes:       notes,
		Billable:    true,
	}

	loggedSession, err := h.apiClient.CreateTracking(user.APIToken, req)
	if err != nil {
		h.sessionMgr.ClearSession(telegramID)
		return h.sendError(chatID, "Failed to log time: "+err.Error())
	}

	// Clear session
	h.sessionMgr.ClearSession(telegramID)

	// Send success message
	var text strings.Builder
	text.WriteString("✅ *Time logged successfully!*\n\n")
	text.WriteString(fmt.Sprintf("⏱️ Hours: %.2f\n", hours))
	if projectName != "" {
		text.WriteString(fmt.Sprintf("📋 Project: %s\n", projectName))
	}
	if notes != "" {
		text.WriteString(fmt.Sprintf("📝 Notes: %s\n", notes))
	}
	if loggedSession.Duration > 0 {
		billableAmount := hours * 100 // Assuming $100/hr for display
		text.WriteString(fmt.Sprintf("\n💰 Estimated: $%.2f (at $100/hr)", billableAmount))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	return nil
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
