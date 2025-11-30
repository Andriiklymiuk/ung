package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// GigHandler handles gig management commands
type GigHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewGigHandler creates a new gig handler
func NewGigHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *GigHandler {
	return &GigHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// Status workflow: todo → in_progress → sent → done
var statusFlow = []string{"todo", "in_progress", "sent", "done"}

// Helper to get status emoji
func getGigStatusEmoji(status string) string {
	switch status {
	case "todo":
		return "📋"
	case "in_progress":
		return "🚀"
	case "sent":
		return "📦"
	case "done":
		return "✅"
	case "on_hold":
		return "⏸️"
	case "cancelled":
		return "❌"
	default:
		return "⚪"
	}
}

// Helper to get next statuses for moving
func getNextStatuses(currentStatus string) []string {
	var next []string
	currentIdx := -1
	for i, s := range statusFlow {
		if s == currentStatus {
			currentIdx = i
			break
		}
	}
	// Can move forward
	if currentIdx >= 0 && currentIdx < len(statusFlow)-1 {
		next = append(next, statusFlow[currentIdx+1])
	}
	// Can move backward
	if currentIdx > 0 {
		next = append(next, statusFlow[currentIdx-1])
	}
	return next
}

// HandleMenu shows the gigs main menu (board view)
func (h *GigHandler) HandleMenu(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "👋 Please /start first")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)
	gigs, err := h.apiClient.ListGigs(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Group gigs by status
	byStatus := make(map[string][]services.Gig)
	for _, gig := range gigs {
		byStatus[gig.Status] = append(byStatus[gig.Status], gig)
	}

	var text strings.Builder
	text.WriteString("📋 *Gigs Board*\n\n")

	if len(gigs) == 0 {
		text.WriteString("_No gigs yet. Tap + to create one!_")
	} else {
		// Show counts by status
		for _, status := range statusFlow {
			count := len(byStatus[status])
			if count > 0 {
				text.WriteString(fmt.Sprintf("%s %s: %d\n", getGigStatusEmoji(status), strings.Title(status), count))
			}
		}

		text.WriteString("\n")

		// Show in_progress gigs first
		if inProgress := byStatus["in_progress"]; len(inProgress) > 0 {
			text.WriteString("*In Progress:*\n")
			for _, gig := range inProgress {
				text.WriteString(fmt.Sprintf("• %s", gig.Name))
				if gig.Project != "" {
					text.WriteString(fmt.Sprintf(" [%s]", gig.Project))
				}
				if gig.TotalHoursTracked > 0 {
					text.WriteString(fmt.Sprintf(" _%.1fh_", gig.TotalHoursTracked))
				}
				text.WriteString("\n")
			}
			text.WriteString("\n")
		}

		// Show todo (queued)
		if todo := byStatus["todo"]; len(todo) > 0 {
			text.WriteString("*Todo:*\n")
			for i, gig := range todo {
				if i >= 3 {
					text.WriteString(fmt.Sprintf("_+%d more_\n", len(todo)-3))
					break
				}
				text.WriteString(fmt.Sprintf("• %s", gig.Name))
				if gig.Project != "" {
					text.WriteString(fmt.Sprintf(" [%s]", gig.Project))
				}
				text.WriteString("\n")
			}
		}
	}

	// Build keyboard
	var rows [][]tgbotapi.InlineKeyboardButton

	// Gig buttons - show first few gigs
	if len(gigs) > 0 {
		var gigRow []tgbotapi.InlineKeyboardButton
		for i, gig := range gigs {
			if i >= 3 {
				break
			}
			emoji := getGigStatusEmoji(gig.Status)
			gigRow = append(gigRow, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", emoji, truncateGig(gig.Name, 12)),
				fmt.Sprintf("gig_view_%d", gig.ID),
			))
		}
		if len(gigRow) > 0 {
			rows = append(rows, gigRow)
		}
	}

	// Filter by status buttons
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📋 Todo", "gig_filter_todo"),
		tgbotapi.NewInlineKeyboardButtonData("🚀 In Progress", "gig_filter_in_progress"),
		tgbotapi.NewInlineKeyboardButtonData("📦 Sent", "gig_filter_sent"),
	))

	// Action buttons
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ New Gig", "gig_create"),
		tgbotapi.NewInlineKeyboardButtonData("📋 All", "gig_list"),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Menu", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleFilter shows gigs filtered by status
func (h *GigHandler) HandleFilter(callbackQuery *tgbotapi.CallbackQuery, status string) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	h.bot.Request(callback)

	user := h.sessionMgr.GetUser(telegramID)
	gigs, err := h.apiClient.ListGigs(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Filter by status
	var filtered []services.Gig
	for _, gig := range gigs {
		if gig.Status == status {
			filtered = append(filtered, gig)
		}
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("%s *%s* (%d)\n\n", getGigStatusEmoji(status), strings.Title(status), len(filtered)))

	if len(filtered) == 0 {
		text.WriteString("_No gigs in this status_")
	} else {
		for _, gig := range filtered {
			text.WriteString(fmt.Sprintf("• *%s*\n", gig.Name))
			if gig.ClientName != "" {
				text.WriteString(fmt.Sprintf("  └ %s", gig.ClientName))
			}
			if gig.TotalHoursTracked > 0 {
				text.WriteString(fmt.Sprintf(" · %.1fh", gig.TotalHoursTracked))
			}
			text.WriteString("\n")
		}
	}

	// Build keyboard with gig buttons
	var rows [][]tgbotapi.InlineKeyboardButton

	for i, gig := range filtered {
		if i >= 5 {
			break
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📄 %s", truncateGig(gig.Name, 25)),
				fmt.Sprintf("gig_view_%d", gig.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Back", "action_gig"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleList shows all gigs
func (h *GigHandler) HandleList(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	h.bot.Request(callback)

	user := h.sessionMgr.GetUser(telegramID)
	gigs, err := h.apiClient.ListGigs(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("📋 *All Gigs* (%d)\n\n", len(gigs)))

	if len(gigs) == 0 {
		text.WriteString("_No gigs yet_")
	} else {
		for i, gig := range gigs {
			if i >= 10 {
				text.WriteString(fmt.Sprintf("\n_+%d more_", len(gigs)-10))
				break
			}
			emoji := getGigStatusEmoji(gig.Status)
			text.WriteString(fmt.Sprintf("%s *%s*\n", emoji, gig.Name))
		}
	}

	// Build keyboard
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < len(gigs) && i < 6; i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s", getGigStatusEmoji(gigs[i].Status), truncateGig(gigs[i].Name, 14)),
			fmt.Sprintf("gig_view_%d", gigs[i].ID),
		))
		if i+1 < len(gigs) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", getGigStatusEmoji(gigs[i+1].Status), truncateGig(gigs[i+1].Name, 14)),
				fmt.Sprintf("gig_view_%d", gigs[i+1].ID),
			))
		}
		rows = append(rows, row)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Back", "action_gig"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleView shows gig details with move buttons
func (h *GigHandler) HandleView(callbackQuery *tgbotapi.CallbackQuery, gigID uint) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	h.bot.Request(callback)

	user := h.sessionMgr.GetUser(telegramID)
	gig, err := h.apiClient.GetGig(user.APIToken, gigID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Gig not found")
		h.bot.Send(msg)
		return err
	}

	// Get tasks
	tasks, _ := h.apiClient.ListGigTasks(user.APIToken, gigID)

	var text strings.Builder
	text.WriteString(fmt.Sprintf("%s *%s*\n\n", getGigStatusEmoji(gig.Status), gig.Name))

	// Status
	text.WriteString(fmt.Sprintf("📊 Status: *%s*\n", strings.Title(gig.Status)))

	// Client
	if gig.ClientName != "" {
		text.WriteString(fmt.Sprintf("👤 Client: %s\n", gig.ClientName))
	}

	// Project
	if gig.Project != "" {
		text.WriteString(fmt.Sprintf("📁 Project: %s\n", gig.Project))
	}

	// Stats
	if gig.TotalHoursTracked > 0 {
		text.WriteString(fmt.Sprintf("⏱ Hours: %.1f\n", gig.TotalHoursTracked))
	}
	if gig.HourlyRate != nil && *gig.HourlyRate > 0 {
		text.WriteString(fmt.Sprintf("💵 Rate: $%.0f/h\n", *gig.HourlyRate))
	}
	if gig.TotalInvoiced > 0 {
		text.WriteString(fmt.Sprintf("💰 Invoiced: $%.0f\n", gig.TotalInvoiced))
	}

	// Tasks
	if len(tasks) > 0 {
		text.WriteString("\n📝 *Tasks:*\n")
		completed := 0
		for _, task := range tasks {
			checkbox := "○"
			if task.Completed {
				checkbox = "✓"
				completed++
			}
			text.WriteString(fmt.Sprintf("%s %s\n", checkbox, task.Title))
		}
		text.WriteString(fmt.Sprintf("_%d/%d done_\n", completed, len(tasks)))
	}

	// Description
	if gig.Description != "" {
		text.WriteString(fmt.Sprintf("\n_%s_", truncateGig(gig.Description, 100)))
	}

	// Build keyboard with move buttons
	var rows [][]tgbotapi.InlineKeyboardButton

	// Move status buttons
	nextStatuses := getNextStatuses(gig.Status)
	if len(nextStatuses) > 0 {
		var moveRow []tgbotapi.InlineKeyboardButton
		for _, status := range nextStatuses {
			emoji := getGigStatusEmoji(status)
			moveRow = append(moveRow, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("→ %s %s", emoji, strings.Title(status)),
				fmt.Sprintf("gig_move_%d_%s", gigID, status),
			))
		}
		rows = append(rows, moveRow)
	}

	// Task buttons
	if len(tasks) > 0 {
		var taskRow []tgbotapi.InlineKeyboardButton
		for i, task := range tasks {
			if i >= 3 {
				break
			}
			icon := "○"
			if task.Completed {
				icon = "✓"
			}
			taskRow = append(taskRow, tgbotapi.NewInlineKeyboardButtonData(
				icon,
				fmt.Sprintf("gig_task_toggle_%d", task.ID),
			))
		}
		rows = append(rows, taskRow)
	}

	// Action buttons
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ Task", fmt.Sprintf("gig_task_add_%d", gigID)),
		tgbotapi.NewInlineKeyboardButtonData("🗑 Delete", fmt.Sprintf("gig_delete_%d", gigID)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Back", "action_gig"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleMove moves a gig to a new status
func (h *GigHandler) HandleMove(callbackQuery *tgbotapi.CallbackQuery, gigID uint, newStatus string) error {
	telegramID := callbackQuery.From.ID

	user := h.sessionMgr.GetUser(telegramID)

	gig, err := h.apiClient.UpdateGigStatus(user.APIToken, gigID, newStatus)
	if err != nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Failed to move")
		h.bot.Request(callback)
		return err
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, fmt.Sprintf("✅ Moved to %s", strings.Title(newStatus)))
	h.bot.Request(callback)

	// Show updated gig view
	return h.HandleView(callbackQuery, gig.ID)
}

// HandleCreate starts gig creation
func (h *GigHandler) HandleCreate(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	h.bot.Request(callback)

	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateGigCreateName),
		Data:       make(map[string]interface{}),
	})

	msg := tgbotapi.NewMessage(chatID, "📋 *New Gig*\n\nWhat's the gig name?")
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "action_gig"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleNameInput handles gig name input
func (h *GigHandler) HandleNameInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	name := strings.TrimSpace(message.Text)
	if name == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Enter a valid name")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Create the gig directly in todo status
	req := services.GigCreateRequest{
		Name:    name,
		Status:  "todo",
		GigType: "hourly",
	}

	gig, err := h.apiClient.CreateGig(user.APIToken, req)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		h.sessionMgr.ClearSession(telegramID)
		return err
	}

	h.sessionMgr.ClearSession(telegramID)

	// Show success and gig view
	var text strings.Builder
	text.WriteString(fmt.Sprintf("✅ *Created!*\n\n📋 %s\n📋 Todo", gig.Name))

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 View", fmt.Sprintf("gig_view_%d", gig.ID)),
			tgbotapi.NewInlineKeyboardButtonData("➕ Add task", fmt.Sprintf("gig_task_add_%d", gig.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Gigs", "action_gig"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleDelete deletes a gig
func (h *GigHandler) HandleDelete(callbackQuery *tgbotapi.CallbackQuery, gigID uint) error {
	telegramID := callbackQuery.From.ID

	user := h.sessionMgr.GetUser(telegramID)

	err := h.apiClient.DeleteGig(user.APIToken, gigID)
	if err != nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Failed to delete")
		h.bot.Request(callback)
		return err
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "✅ Deleted")
	h.bot.Request(callback)

	// Go back to gig menu
	fakeMsg := &tgbotapi.Message{
		Chat: callbackQuery.Message.Chat,
		From: callbackQuery.From,
	}
	return h.HandleMenu(fakeMsg)
}

// HandleTaskAdd starts adding a task
func (h *GigHandler) HandleTaskAdd(callbackQuery *tgbotapi.CallbackQuery, gigID uint) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	h.bot.Request(callback)

	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateGigTaskAdd),
		Data: map[string]interface{}{
			"gig_id": gigID,
		},
	})

	msg := tgbotapi.NewMessage(chatID, "📝 Enter task title:")
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel", fmt.Sprintf("gig_view_%d", gigID)),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleTaskInput handles task title input
func (h *GigHandler) HandleTaskInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	title := strings.TrimSpace(message.Text)
	if title == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Enter a valid task title")
		h.bot.Send(msg)
		return nil
	}

	session := h.sessionMgr.GetSession(telegramID)
	gigID, ok := session.Data["gig_id"].(uint)
	if !ok {
		// Try float64 (JSON unmarshaling)
		if gigIDFloat, ok := session.Data["gig_id"].(float64); ok {
			gigID = uint(gigIDFloat)
		}
	}

	user := h.sessionMgr.GetUser(telegramID)

	task, err := h.apiClient.CreateGigTask(user.APIToken, gigID, title)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		h.sessionMgr.ClearSession(telegramID)
		return err
	}

	h.sessionMgr.ClearSession(telegramID)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Task added: %s", task.Title))
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 View gig", fmt.Sprintf("gig_view_%d", gigID)),
			tgbotapi.NewInlineKeyboardButtonData("➕ Add another", fmt.Sprintf("gig_task_add_%d", gigID)),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleTaskToggle toggles task completion
func (h *GigHandler) HandleTaskToggle(callbackQuery *tgbotapi.CallbackQuery, taskID uint) error {
	telegramID := callbackQuery.From.ID

	user := h.sessionMgr.GetUser(telegramID)

	task, err := h.apiClient.ToggleGigTask(user.APIToken, taskID)
	if err != nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Failed")
		h.bot.Request(callback)
		return err
	}

	status := "✓ Done"
	if !task.Completed {
		status = "○ Todo"
	}
	callback := tgbotapi.NewCallback(callbackQuery.ID, status)
	h.bot.Request(callback)

	// Refresh the gig view
	return h.HandleView(callbackQuery, task.GigID)
}

func truncateGig(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
