package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// HunterHandler handles job hunting commands
type HunterHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewHunterHandler creates a new hunter handler
func NewHunterHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *HunterHandler {
	return &HunterHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleMenu shows the hunter main menu
func (h *HunterHandler) HandleMenu(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "👋 Please /start first")
		h.bot.Send(msg)
		return nil
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("👤 Profile", "hunter_profile"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Applications", "hunter_applications"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Stats", "hunter_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Menu", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID,
		"🎯 *Job Hunter*\n\n"+
			"🇺🇸 🇺🇦 🇳🇱 🇪🇺 sources")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleHunt starts a job hunt - checks profile first, prompts for PDF if needed
func (h *HunterHandler) HandleHunt(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "👋 Please /start first to connect your account")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Check if profile exists
	profile, _ := h.apiClient.GetHunterProfile(user.APIToken)
	if profile == nil || len(profile.Skills) == 0 {
		// No profile - ask for PDF
		h.sessionMgr.SetSession(&models.Session{
			TelegramID: telegramID,
			State:      string(models.StateHunterAwaitingPDF),
			Data:       make(map[string]interface{}),
		})

		msg := tgbotapi.NewMessage(chatID,
			"📄 *Send me your CV* (PDF)\n\n"+
				"I'll extract your skills and find matching jobs!")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✍️ Enter skills manually", "hunter_profile_create"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	// Has profile - start hunting
	return h.doHunt(chatID, user.APIToken)
}

// doHunt performs the actual job hunting
func (h *HunterHandler) doHunt(chatID int64, token string) error {
	// Show brief hunting message
	huntingMsg := tgbotapi.NewMessage(chatID, "🔍 Searching jobs...")
	huntingMsg.ParseMode = "Markdown"
	h.bot.Send(huntingMsg)

	result, err := h.apiClient.HuntJobs(token)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Build results message
	var text strings.Builder

	if result.NewCount > 0 {
		text.WriteString(fmt.Sprintf("✨ *%d new jobs found!*\n\n", result.NewCount))
	} else {
		text.WriteString("📋 *Your matched jobs:*\n\n")
	}

	// Show top jobs inline
	for i, job := range result.Jobs {
		if i >= 5 {
			break
		}
		sourceFlag := getSourceFlag(job.Source)
		matchEmoji := getMatchEmoji(job.MatchScore)

		text.WriteString(fmt.Sprintf("%s %s *%s*\n", matchEmoji, sourceFlag, truncate(job.Title, 35)))
		if job.Company != "" {
			text.WriteString(fmt.Sprintf("└ %s", job.Company))
			if job.MatchScore > 0 {
				text.WriteString(fmt.Sprintf(" · %d%%", int(job.MatchScore)))
			}
			text.WriteString("\n")
		}
		text.WriteString("\n")
	}

	if len(result.Jobs) > 5 {
		text.WriteString(fmt.Sprintf("_+%d more jobs_", len(result.Jobs)-5))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	// Job buttons - compact
	var rows [][]tgbotapi.InlineKeyboardButton
	if len(result.Jobs) > 0 {
		// First row with top 2 jobs
		var firstRow []tgbotapi.InlineKeyboardButton
		for i, job := range result.Jobs {
			if i >= 2 {
				break
			}
			firstRow = append(firstRow, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📄 %s", truncate(job.Title, 18)),
				fmt.Sprintf("hunter_job_%d", job.ID),
			))
		}
		if len(firstRow) > 0 {
			rows = append(rows, firstRow)
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📋 All jobs", "hunter_jobs"),
		tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "hunter_hunt"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleHuntCallback handles hunt callback (when called from button)
func (h *HunterHandler) HandleHuntCallback(callbackQuery *tgbotapi.CallbackQuery) error {
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Starting hunt...")
	h.bot.Request(callback)

	msg := &tgbotapi.Message{
		Chat: callbackQuery.Message.Chat,
		From: callbackQuery.From,
	}
	return h.HandleHunt(msg)
}

// HandleJobs shows matched jobs list
func (h *HunterHandler) HandleJobs(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "👋 Please /start first")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	jobs, err := h.apiClient.GetHunterJobs(user.APIToken, 10)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	if len(jobs) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📋 No jobs yet\n\nTap Hunt to find matching jobs!")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt", "hunter_hunt"),
				tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("📋 *%d jobs*\n\n", len(jobs)))

	for i, job := range jobs {
		if i >= 6 {
			break
		}
		sourceFlag := getSourceFlag(job.Source)
		matchEmoji := getMatchEmoji(job.MatchScore)

		text.WriteString(fmt.Sprintf("%s %s *%s*\n", matchEmoji, sourceFlag, truncate(job.Title, 32)))
		if job.Company != "" {
			text.WriteString(fmt.Sprintf("└ %s", job.Company))
			if job.MatchScore > 0 {
				text.WriteString(fmt.Sprintf(" · %d%%", int(job.MatchScore)))
			}
			text.WriteString("\n")
		}
		text.WriteString("\n")
	}

	if len(jobs) > 6 {
		text.WriteString(fmt.Sprintf("_+%d more_", len(jobs)-6))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	// Job buttons - 2 per row
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(jobs) && i < 4; i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📄 %s", truncate(jobs[i].Title, 16)),
			fmt.Sprintf("hunter_job_%d", jobs[i].ID),
		))
		if i+1 < len(jobs) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📄 %s", truncate(jobs[i+1].Title, 16)),
				fmt.Sprintf("hunter_job_%d", jobs[i+1].ID),
			))
		}
		rows = append(rows, row)
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
		),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleJobDetail shows details of a specific job
func (h *HunterHandler) HandleJobDetail(callbackQuery *tgbotapi.CallbackQuery, jobID uint) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	user := h.sessionMgr.GetUser(telegramID)

	jobs, err := h.apiClient.GetHunterJobs(user.APIToken, 50)
	if err != nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Error loading job")
		h.bot.Request(callback)
		return err
	}

	// Find the job
	var job *services.HunterJob
	for _, j := range jobs {
		if j.ID == jobID {
			job = &j
			break
		}
	}

	if job == nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Job not found")
		h.bot.Request(callback)
		return nil
	}

	sourceFlag := getSourceFlag(job.Source)

	var text strings.Builder
	text.WriteString(fmt.Sprintf("*%s*\n\n", job.Title))

	if job.Company != "" {
		text.WriteString(fmt.Sprintf("🏢 %s\n", job.Company))
	}
	if job.Location != "" {
		text.WriteString(fmt.Sprintf("📍 %s\n", job.Location))
	}

	// Compact info line
	text.WriteString(fmt.Sprintf("%s %s", sourceFlag, job.Source))
	if job.MatchScore > 0 {
		text.WriteString(fmt.Sprintf(" · %d%% match", int(job.MatchScore)))
	}
	if job.Remote {
		text.WriteString(" · 🏠 Remote")
	}
	text.WriteString("\n\n")

	if job.Description != "" {
		desc := truncate(job.Description, 250)
		text.WriteString(fmt.Sprintf("_%s_", desc))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	// Build keyboard
	var rows [][]tgbotapi.InlineKeyboardButton

	// View original job link if available
	if job.SourceURL != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔗 View job", job.SourceURL),
		))
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✨ Generate proposal", fmt.Sprintf("hunter_proposal_%d", jobID)),
			tgbotapi.NewInlineKeyboardButtonData("✅ Mark applied", fmt.Sprintf("hunter_apply_%d", jobID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back", "hunter_jobs"),
		),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	h.bot.Request(callback)
	return nil
}

// HandleProfile shows hunter profile
func (h *HunterHandler) HandleProfile(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "👋 Please /start first")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	profile, err := h.apiClient.GetHunterProfile(user.APIToken)
	if err != nil || profile == nil {
		// No profile yet - prompt to create one
		msg := tgbotapi.NewMessage(chatID, "👤 *No profile yet*\n\nUpload your CV to get started!")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📄 Upload CV", "hunter_hunt"),
				tgbotapi.NewInlineKeyboardButtonData("✍️ Manual", "hunter_profile_create"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	var text strings.Builder
	text.WriteString("👤 *Profile*\n\n")

	if profile.Name != "" {
		text.WriteString(fmt.Sprintf("*%s*\n", profile.Name))
	}
	if profile.Title != "" {
		text.WriteString(fmt.Sprintf("💼 %s\n", profile.Title))
	}
	if len(profile.Skills) > 0 {
		skillsDisplay := profile.Skills
		if len(skillsDisplay) > 6 {
			skillsDisplay = skillsDisplay[:6]
		}
		text.WriteString(fmt.Sprintf("🛠 %s", strings.Join(skillsDisplay, ", ")))
		if len(profile.Skills) > 6 {
			text.WriteString(fmt.Sprintf(" +%d", len(profile.Skills)-6))
		}
		text.WriteString("\n")
	}
	if profile.Rate > 0 {
		text.WriteString(fmt.Sprintf("💰 $%.0f/hr", profile.Rate))
	}
	if profile.Experience > 0 {
		text.WriteString(fmt.Sprintf(" · %dy exp", profile.Experience))
	}
	if profile.Remote {
		text.WriteString(" · 🏠")
	}
	text.WriteString("\n")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Update CV", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit", "hunter_profile_edit"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleProfileCreate starts profile creation flow
func (h *HunterHandler) HandleProfileCreate(callbackQuery *tgbotapi.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateHunterProfileName),
		Data:       make(map[string]interface{}),
	})

	msg := tgbotapi.NewMessage(chatID, "✍️ *Create Profile* (1/4)\n\nWhat's your name?")
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "action_hunter"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	h.bot.Request(callback)
	return nil
}

// HandleNameInput handles name input for profile creation
func (h *HunterHandler) HandleNameInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	name := strings.TrimSpace(message.Text)
	if name == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Enter a valid name")
		h.bot.Send(msg)
		return nil
	}

	// Store name and move to next step
	session := h.sessionMgr.GetSession(telegramID)
	if session == nil {
		session = &models.Session{
			TelegramID: telegramID,
			Data:       make(map[string]interface{}),
		}
	}
	session.Data["name"] = name
	session.State = string(models.StateHunterProfileTitle)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "💼 *(2/4)* Your job title?\n\n_e.g. Senior Go Developer_")
	msg.ParseMode = "Markdown"

	h.bot.Send(msg)
	return nil
}

// HandleTitleInput handles title input for profile creation
func (h *HunterHandler) HandleTitleInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	title := strings.TrimSpace(message.Text)
	if title == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Enter a valid title")
		h.bot.Send(msg)
		return nil
	}

	session := h.sessionMgr.GetSession(telegramID)
	session.Data["title"] = title
	session.State = string(models.StateHunterProfileSkills)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "🛠 *(3/4)* Your skills?\n\n_Comma-separated: Go, Python, AWS_")
	msg.ParseMode = "Markdown"

	h.bot.Send(msg)
	return nil
}

// HandleSkillsInput handles skills input for profile creation
func (h *HunterHandler) HandleSkillsInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	skillsStr := strings.TrimSpace(message.Text)
	if skillsStr == "" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Enter at least one skill")
		h.bot.Send(msg)
		return nil
	}

	// Parse skills
	skills := []string{}
	for _, skill := range strings.Split(skillsStr, ",") {
		skill = strings.TrimSpace(skill)
		if skill != "" {
			skills = append(skills, skill)
		}
	}

	session := h.sessionMgr.GetSession(telegramID)
	session.Data["skills"] = skills
	session.State = string(models.StateHunterProfileRate)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID, "💰 *(4/4)* Hourly rate in USD?\n\n_e.g. 75_")
	msg.ParseMode = "Markdown"

	h.bot.Send(msg)
	return nil
}

// HandleRateInput handles rate input and creates the profile
func (h *HunterHandler) HandleRateInput(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	rateStr := strings.TrimSpace(message.Text)
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil || rate <= 0 {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Enter a valid rate (e.g. 75)")
		h.bot.Send(msg)
		return nil
	}

	session := h.sessionMgr.GetSession(telegramID)
	data := session.Data

	user := h.sessionMgr.GetUser(telegramID)

	// Extract skills from session
	var skills []string
	if s, ok := data["skills"].([]string); ok {
		skills = s
	}

	// Create/update profile
	profile := &services.HunterProfile{
		Name:   data["name"].(string),
		Title:  data["title"].(string),
		Skills: skills,
		Rate:   rate,
		Remote: true,
	}

	updatedProfile, err := h.apiClient.UpdateHunterProfile(user.APIToken, profile)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		h.sessionMgr.ClearSession(telegramID)
		return err
	}

	h.sessionMgr.ClearSession(telegramID)

	var text strings.Builder
	text.WriteString("✅ *Profile created!*\n\n")
	text.WriteString(fmt.Sprintf("💼 %s\n", updatedProfile.Title))
	text.WriteString(fmt.Sprintf("🛠 %s\n", strings.Join(updatedProfile.Skills, ", ")))
	text.WriteString(fmt.Sprintf("💰 $%.0f/hr", updatedProfile.Rate))

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt jobs", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("« Menu", "action_hunter"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleStats shows hunter statistics
func (h *HunterHandler) HandleStats(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "👋 Please /start first")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	stats, err := h.apiClient.GetHunterStats(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Extract counts from StatusCounts
	pending := stats.StatusCounts["pending"]
	accepted := stats.StatusCounts["accepted"]
	rejected := stats.StatusCounts["rejected"]

	var text strings.Builder
	text.WriteString("📊 *Stats*\n\n")
	text.WriteString(fmt.Sprintf("📋 %d jobs · 📝 %d applied\n", stats.TotalJobs, stats.TotalApplications))
	text.WriteString(fmt.Sprintf("⏳ %d pending · ✅ %d accepted · ❌ %d rejected\n", pending, accepted, rejected))
	if stats.AverageMatchScore > 0 {
		text.WriteString(fmt.Sprintf("🎯 %.0f%% avg match\n", stats.AverageMatchScore))
	}
	if len(stats.TopSkills) > 0 {
		skillsDisplay := stats.TopSkills
		if len(skillsDisplay) > 4 {
			skillsDisplay = skillsDisplay[:4]
		}
		text.WriteString(fmt.Sprintf("🔥 %s", strings.Join(skillsDisplay, ", ")))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleApplications shows user's applications
func (h *HunterHandler) HandleApplications(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "👋 Please /start first")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	applications, err := h.apiClient.GetHunterApplications(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	if len(applications) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📝 *No applications yet*\n\nFind jobs and start applying!")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 Jobs", "hunter_jobs"),
				tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt", "hunter_hunt"),
				tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("📝 *%d Applications*\n\n", len(applications)))

	for i, app := range applications {
		if i >= 5 {
			break
		}
		statusEmoji := getStatusEmoji(app.Status)
		text.WriteString(fmt.Sprintf("%s #%d · %s", statusEmoji, app.ID, app.Status))
		if app.AppliedAt != "" && len(app.AppliedAt) >= 10 {
			text.WriteString(fmt.Sprintf(" · %s", app.AppliedAt[:10]))
		}
		text.WriteString("\n")
	}

	if len(applications) > 5 {
		text.WriteString(fmt.Sprintf("\n_+%d more_", len(applications)-5))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleGenerateProposal generates a proposal for a job
func (h *HunterHandler) HandleGenerateProposal(callbackQuery *tgbotapi.CallbackQuery, jobID uint) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Generating...")
	h.bot.Request(callback)

	user := h.sessionMgr.GetUser(telegramID)

	// Show generating message
	genMsg := tgbotapi.NewMessage(chatID, "✨ Generating proposal...")
	h.bot.Send(genMsg)

	app, err := h.apiClient.CreateHunterApplication(user.APIToken, jobID, true)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("✨ *Proposal*\n\n")
	text.WriteString(app.Proposal)

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("« Back", "action_hunter"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleApply marks a job as applied
func (h *HunterHandler) HandleApply(callbackQuery *tgbotapi.CallbackQuery, jobID uint) error {
	chatID := callbackQuery.Message.Chat.ID
	telegramID := callbackQuery.From.ID

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Saving...")
	h.bot.Request(callback)

	user := h.sessionMgr.GetUser(telegramID)

	_, err := h.apiClient.CreateHunterApplication(user.APIToken, jobID, false)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ "+err.Error())
		h.bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(chatID, "✅ *Applied!* Good luck!")
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Applications", "hunter_applications"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("« Menu", "action_hunter"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandlePDFUpload handles PDF CV upload for profile import
func (h *HunterHandler) HandlePDFUpload(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	user := h.sessionMgr.GetUser(telegramID)

	// Get the document
	doc := message.Document
	if doc == nil {
		msg := tgbotapi.NewMessage(chatID, "📄 Please send a PDF file")
		h.bot.Send(msg)
		return nil
	}

	// Check if it's a PDF
	if doc.MimeType != "application/pdf" {
		msg := tgbotapi.NewMessage(chatID, "⚠️ Only PDF files are supported")
		h.bot.Send(msg)
		return nil
	}

	// Show processing message
	processingMsg := tgbotapi.NewMessage(chatID, "⏳ Reading your CV...")
	h.bot.Send(processingMsg)

	// Download the file
	fileConfig := tgbotapi.FileConfig{FileID: doc.FileID}
	file, err := h.bot.GetFile(fileConfig)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Could not download file")
		h.bot.Send(msg)
		return err
	}

	// Get file URL and download
	fileURL := file.Link(h.bot.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Download failed")
		h.bot.Send(msg)
		return err
	}
	defer resp.Body.Close()

	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Could not read file")
		h.bot.Send(msg)
		return err
	}

	// Import profile from PDF
	profile, err := h.apiClient.ImportProfileFromPDF(user.APIToken, pdfData, doc.FileName)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Could not process CV: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Clear the waiting state
	h.sessionMgr.ClearSession(telegramID)

	// Show compact profile summary
	var text strings.Builder
	text.WriteString("✅ *Got it!*\n\n")

	if profile.Title != "" {
		text.WriteString(fmt.Sprintf("💼 %s\n", profile.Title))
	}
	if len(profile.Skills) > 0 {
		skillsPreview := profile.Skills
		if len(skillsPreview) > 5 {
			skillsPreview = skillsPreview[:5]
		}
		text.WriteString(fmt.Sprintf("🛠 %s", strings.Join(skillsPreview, ", ")))
		if len(profile.Skills) > 5 {
			text.WriteString(fmt.Sprintf(" +%d more", len(profile.Skills)-5))
		}
		text.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Automatically start hunting
	return h.doHunt(chatID, user.APIToken)
}

// Helper functions

func getSourceFlag(source string) string {
	switch source {
	case "djinni", "dou":
		return "🇺🇦"
	case "netherlands":
		return "🇳🇱"
	case "eurojobs", "arbeitnow":
		return "🇪🇺"
	case "hackernews", "remoteok", "weworkremotely":
		return "🇺🇸"
	default:
		return "🌐"
	}
}

func getMatchEmoji(score float64) string {
	if score >= 80 {
		return "🟢"
	} else if score >= 60 {
		return "🟡"
	} else if score >= 40 {
		return "🟠"
	}
	return "⚪"
}

func getStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "sent":
		return "📤"
	case "viewed":
		return "👁"
	case "interview":
		return "📞"
	case "accepted":
		return "✅"
	case "rejected":
		return "❌"
	default:
		return "📝"
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
