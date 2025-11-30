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
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt Jobs", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("📋 My Jobs", "hunter_jobs"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Profile", "hunter_profile"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Stats", "hunter_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Applications", "hunter_applications"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back to Menu", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID,
		"--------------------\n"+
			"      *Job Hunter*\n"+
			"--------------------\n\n"+
			"🎯 Find your next opportunity!\n\n"+
			"_Sources: HackerNews, RemoteOK, WeWorkRemotely,_\n"+
			"_Djinni 🇺🇦, DOU 🇺🇦, Netherlands 🇳🇱, EuroJobs 🇪🇺_")
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
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
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
			"🎯 *Job Hunter*\n\n"+
				"To find jobs matching your skills, I need your CV.\n\n"+
				"📄 *Send me your CV as a PDF file*\n\n"+
				"_I'll extract your skills and start hunting!_")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✍️ Enter manually instead", "hunter_profile_create"),
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
	// Show "hunting" message
	huntingMsg := tgbotapi.NewMessage(chatID,
		"🔍 *Hunting for jobs...*\n\n"+
			"_Searching across multiple platforms:_\n"+
			"• HackerNews Jobs 🇺🇸\n"+
			"• RemoteOK 🇺🇸\n"+
			"• WeWorkRemotely 🇺🇸\n"+
			"• Djinni 🇺🇦\n"+
			"• DOU 🇺🇦\n"+
			"• Netherlands 🇳🇱\n"+
			"• EuroJobs 🇪🇺\n\n"+
			"_This may take a minute..._")
	huntingMsg.ParseMode = "Markdown"
	h.bot.Send(huntingMsg)

	result, err := h.apiClient.HuntJobs(token)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to hunt jobs: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Show results with top jobs inline
	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Hunt Complete!*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("🆕 *New Jobs:* %d\n", result.NewCount))
	text.WriteString(fmt.Sprintf("📊 *Total Jobs:* %d\n\n", result.TotalCount))

	// Show top 3 jobs if available
	if len(result.Jobs) > 0 {
		text.WriteString("*Top Matches:*\n")
		for i, job := range result.Jobs {
			if i >= 3 {
				break
			}
			sourceFlag := getSourceFlag(job.Source)
			matchEmoji := getMatchEmoji(job.MatchScore)
			text.WriteString(fmt.Sprintf("%s %s *%s*\n", matchEmoji, sourceFlag, truncate(job.Title, 30)))
			if job.Company != "" {
				text.WriteString(fmt.Sprintf("   🏢 %s | %.0f%% match\n", job.Company, job.MatchScore))
			}
		}
		text.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	// Create job buttons for top 3
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, job := range result.Jobs {
		if i >= 3 {
			break
		}
		shortTitle := truncate(job.Title, 25)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("📄 %s", shortTitle), fmt.Sprintf("hunter_job_%d", job.ID)),
		))
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 All Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt Again", "hunter_hunt"),
		),
	)

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
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	jobs, err := h.apiClient.GetHunterJobs(user.APIToken, 10)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch jobs: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	if len(jobs) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			"--------------------\n"+
				"      *No Jobs Found*\n"+
				"--------------------\n\n"+
				"You haven't found any jobs yet.\n\n"+
				"_Use /hunt to search for new opportunities!_")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt Jobs", "hunter_hunt"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Top Matched Jobs*\n")
	text.WriteString("--------------------\n\n")

	for i, job := range jobs {
		if i >= 5 {
			break
		}
		sourceFlag := getSourceFlag(job.Source)
		matchEmoji := getMatchEmoji(job.MatchScore)
		text.WriteString(fmt.Sprintf("%s *%s*\n", matchEmoji, job.Title))
		if job.Company != "" {
			text.WriteString(fmt.Sprintf("   🏢 %s\n", job.Company))
		}
		text.WriteString(fmt.Sprintf("   %s %s | Match: %.0f%%\n\n", sourceFlag, job.Source, job.MatchScore))
	}

	text.WriteString(fmt.Sprintf("_Showing top %d of %d jobs_", min(5, len(jobs)), len(jobs)))

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	// Create job buttons for top 3
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, job := range jobs {
		if i >= 3 {
			break
		}
		shortTitle := truncate(job.Title, 25)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("📄 %s", shortTitle), fmt.Sprintf("hunter_job_%d", job.ID)),
		))
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt More", "hunter_hunt"),
			tgbotapi.NewInlineKeyboardButtonData("« Menu", "action_hunter"),
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
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch job: "+err.Error())
		h.bot.Send(msg)

		callback := tgbotapi.NewCallback(callbackQuery.ID, "Error")
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
		msg := tgbotapi.NewMessage(chatID, "Job not found")
		h.bot.Send(msg)

		callback := tgbotapi.NewCallback(callbackQuery.ID, "Not found")
		h.bot.Request(callback)
		return nil
	}

	sourceFlag := getSourceFlag(job.Source)

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString(fmt.Sprintf("*%s*\n", job.Title))
	text.WriteString("--------------------\n\n")

	if job.Company != "" {
		text.WriteString(fmt.Sprintf("🏢 *Company:* %s\n", job.Company))
	}
	if job.Location != "" {
		text.WriteString(fmt.Sprintf("📍 *Location:* %s\n", job.Location))
	}
	text.WriteString(fmt.Sprintf("%s *Source:* %s\n", sourceFlag, job.Source))
	text.WriteString(fmt.Sprintf("🎯 *Match:* %.0f%%\n", job.MatchScore))

	if job.Remote {
		text.WriteString("🌐 *Remote:* Yes\n")
	}

	text.WriteString("\n")

	if job.Description != "" {
		desc := truncate(job.Description, 300)
		text.WriteString(fmt.Sprintf("_%s_\n\n", desc))
	}

	if job.SourceURL != "" {
		text.WriteString(fmt.Sprintf("[🔗 View Original Job](%s)\n", job.SourceURL))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Generate Proposal", fmt.Sprintf("hunter_proposal_%d", jobID)),
			tgbotapi.NewInlineKeyboardButtonData("✅ Apply", fmt.Sprintf("hunter_apply_%d", jobID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back to Jobs", "hunter_jobs"),
		),
	)
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
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	profile, err := h.apiClient.GetHunterProfile(user.APIToken)
	if err != nil || profile == nil {
		// No profile yet - prompt to create one
		msg := tgbotapi.NewMessage(chatID,
			"--------------------\n"+
				"      *No Profile Found*\n"+
				"--------------------\n\n"+
				"You haven't set up your hunter profile yet.\n\n"+
				"_Set up your profile to get matched jobs!_")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📝 Create Profile", "hunter_profile_create"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Hunter Profile*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("👤 *Name:* %s\n", profile.Name))
	text.WriteString(fmt.Sprintf("💼 *Title:* %s\n", profile.Title))

	if len(profile.Skills) > 0 {
		text.WriteString(fmt.Sprintf("🛠 *Skills:* %s\n", strings.Join(profile.Skills, ", ")))
	}

	if profile.Rate > 0 {
		text.WriteString(fmt.Sprintf("💰 *Rate:* $%.0f/hr\n", profile.Rate))
	}

	if profile.Experience > 0 {
		text.WriteString(fmt.Sprintf("📅 *Experience:* %d years\n", profile.Experience))
	}

	if profile.Location != "" {
		text.WriteString(fmt.Sprintf("📍 *Location:* %s\n", profile.Location))
	}

	if profile.Remote {
		text.WriteString("🌐 *Remote:* Yes\n")
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit Profile", "hunter_profile_edit"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
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

	msg := tgbotapi.NewMessage(chatID,
		"--------------------\n"+
			"      *Create Profile*\n"+
			"--------------------\n\n"+
			"Let's set up your hunter profile!\n\n"+
			"*Step 1/4*: What's your name?")
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
		msg := tgbotapi.NewMessage(chatID, "Please enter a valid name.")
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

	msg := tgbotapi.NewMessage(chatID,
		"*Step 2/4*: What's your professional title?\n\n"+
			"_(e.g., Senior Go Developer, Full Stack Engineer)_")
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
		msg := tgbotapi.NewMessage(chatID, "Please enter a valid title.")
		h.bot.Send(msg)
		return nil
	}

	session := h.sessionMgr.GetSession(telegramID)
	session.Data["title"] = title
	session.State = string(models.StateHunterProfileSkills)
	h.sessionMgr.SetSession(session)

	msg := tgbotapi.NewMessage(chatID,
		"*Step 3/4*: What are your main skills?\n\n"+
			"_(Enter comma-separated skills, e.g., Go, Python, PostgreSQL, AWS)_")
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
		msg := tgbotapi.NewMessage(chatID, "Please enter at least one skill.")
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

	msg := tgbotapi.NewMessage(chatID,
		"*Step 4/4*: What's your hourly rate? (USD)\n\n"+
			"_(Enter a number, e.g., 75)_")
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
		msg := tgbotapi.NewMessage(chatID, "Please enter a valid rate (e.g., 75).")
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
		msg := tgbotapi.NewMessage(chatID, "Failed to create profile: "+err.Error())
		h.bot.Send(msg)

		h.sessionMgr.ClearSession(telegramID)
		return err
	}

	h.sessionMgr.ClearSession(telegramID)

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Profile Created!*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("👤 *Name:* %s\n", updatedProfile.Name))
	text.WriteString(fmt.Sprintf("💼 *Title:* %s\n", updatedProfile.Title))
	text.WriteString(fmt.Sprintf("🛠 *Skills:* %s\n", strings.Join(updatedProfile.Skills, ", ")))
	text.WriteString(fmt.Sprintf("💰 *Rate:* $%.0f/hr\n\n", updatedProfile.Rate))
	text.WriteString("_You're ready to hunt for jobs!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Start Hunting", "hunter_hunt"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
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
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	stats, err := h.apiClient.GetHunterStats(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch stats: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Extract counts from StatusCounts
	pending := stats.StatusCounts["pending"]
	accepted := stats.StatusCounts["accepted"]
	rejected := stats.StatusCounts["rejected"]

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Hunter Stats*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("📋 *Total Jobs:* %d\n", stats.TotalJobs))
	text.WriteString(fmt.Sprintf("📝 *Applications:* %d\n", stats.TotalApplications))
	text.WriteString(fmt.Sprintf("⏳ *Pending:* %d\n", pending))
	text.WriteString(fmt.Sprintf("✅ *Accepted:* %d\n", accepted))
	text.WriteString(fmt.Sprintf("❌ *Rejected:* %d\n", rejected))
	text.WriteString(fmt.Sprintf("🎯 *Avg Match:* %.0f%%\n", stats.AverageMatchScore))
	if len(stats.TopSkills) > 0 {
		text.WriteString(fmt.Sprintf("🛠 *Top Skills:* %s\n", strings.Join(stats.TopSkills, ", ")))
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 View Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt", "hunter_hunt"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
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
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	applications, err := h.apiClient.GetHunterApplications(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch applications: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	if len(applications) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			"--------------------\n"+
				"      *No Applications*\n"+
				"--------------------\n\n"+
				"You haven't applied to any jobs yet.\n\n"+
				"_Find jobs with /jobs and apply!_")
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 View Jobs", "hunter_jobs"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
			),
		)
		msg.ReplyMarkup = keyboard

		h.bot.Send(msg)
		return nil
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *My Applications*\n")
	text.WriteString("--------------------\n\n")

	for i, app := range applications {
		if i >= 5 {
			break
		}
		statusEmoji := getStatusEmoji(app.Status)
		text.WriteString(fmt.Sprintf("%s *Application #%d*\n", statusEmoji, app.ID))
		text.WriteString(fmt.Sprintf("   Job ID: %d | Status: %s\n", app.JobID, app.Status))
		if app.AppliedAt != "" {
			text.WriteString(fmt.Sprintf("   Applied: %s\n", app.AppliedAt[:10]))
		}
		text.WriteString("\n")
	}

	text.WriteString(fmt.Sprintf("_Showing %d of %d applications_", min(5, len(applications)), len(applications)))

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 View Jobs", "hunter_jobs"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Hunt", "hunter_hunt"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
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

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Generating proposal...")
	h.bot.Request(callback)

	user := h.sessionMgr.GetUser(telegramID)

	// Show generating message
	genMsg := tgbotapi.NewMessage(chatID, "✨ *Generating personalized proposal...*\n\n_This may take a moment._")
	genMsg.ParseMode = "Markdown"
	h.bot.Send(genMsg)

	app, err := h.apiClient.CreateHunterApplication(user.APIToken, jobID, true)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to generate proposal: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Generated Proposal*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(app.Proposal)
	text.WriteString("\n\n_Copy and customize this proposal!_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back to Jobs", "hunter_jobs"),
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

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Creating application...")
	h.bot.Request(callback)

	user := h.sessionMgr.GetUser(telegramID)

	_, err := h.apiClient.CreateHunterApplication(user.APIToken, jobID, false)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to create application: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	msg := tgbotapi.NewMessage(chatID,
		"--------------------\n"+
			"      *Application Created!*\n"+
			"--------------------\n\n"+
			"✅ Your application has been tracked.\n\n"+
			"_Good luck with your application!_")
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 My Applications", "hunter_applications"),
			tgbotapi.NewInlineKeyboardButtonData("📋 More Jobs", "hunter_jobs"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Hunter Menu", "action_hunter"),
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
		msg := tgbotapi.NewMessage(chatID, "Please send a PDF file.")
		h.bot.Send(msg)
		return nil
	}

	// Check if it's a PDF
	if doc.MimeType != "application/pdf" {
		msg := tgbotapi.NewMessage(chatID, "Please send a *PDF* file only.")
		msg.ParseMode = "Markdown"
		h.bot.Send(msg)
		return nil
	}

	// Show processing message
	processingMsg := tgbotapi.NewMessage(chatID,
		"📄 *Processing your CV...*\n\n"+
			"_Extracting skills and experience..._")
	processingMsg.ParseMode = "Markdown"
	h.bot.Send(processingMsg)

	// Download the file
	fileConfig := tgbotapi.FileConfig{FileID: doc.FileID}
	file, err := h.bot.GetFile(fileConfig)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to download file: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Get file URL and download
	fileURL := file.Link(h.bot.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to download file: "+err.Error())
		h.bot.Send(msg)
		return err
	}
	defer resp.Body.Close()

	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to read file: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Import profile from PDF
	profile, err := h.apiClient.ImportProfileFromPDF(user.APIToken, pdfData, doc.FileName)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to process CV: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Clear the waiting state
	h.sessionMgr.ClearSession(telegramID)

	// Show extracted profile
	var text strings.Builder
	text.WriteString("✅ *Profile Created from CV!*\n\n")
	if profile.Name != "" {
		text.WriteString(fmt.Sprintf("👤 *Name:* %s\n", profile.Name))
	}
	if profile.Title != "" {
		text.WriteString(fmt.Sprintf("💼 *Title:* %s\n", profile.Title))
	}
	if len(profile.Skills) > 0 {
		text.WriteString(fmt.Sprintf("🛠 *Skills:* %s\n", strings.Join(profile.Skills, ", ")))
	}
	if profile.Experience > 0 {
		text.WriteString(fmt.Sprintf("📅 *Experience:* %d years\n", profile.Experience))
	}
	text.WriteString("\n_Starting job hunt..._")

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
