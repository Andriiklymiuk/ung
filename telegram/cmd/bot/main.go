package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/config"
	"ung-telegram/internal/handlers"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Println("Starting UNG Telegram Bot...")

	// Initialize bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatal("Failed to initialize bot:", err)
	}

	bot.Debug = cfg.Debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize services
	apiClient := services.NewAPIClient(cfg.APIURL)
	sessionMgr := services.NewSessionManager()

	// Initialize handlers
	startHandler := handlers.NewStartHandler(bot, sessionMgr, cfg.WebAppURL)
	helpHandler := handlers.NewHelpHandler(bot)
	invoiceHandler := handlers.NewInvoiceHandler(bot, apiClient, sessionMgr)
	clientHandler := handlers.NewClientHandler(bot, apiClient, sessionMgr)
	companyHandler := handlers.NewCompanyHandler(bot, apiClient, sessionMgr)
	contractHandler := handlers.NewContractHandler(bot, apiClient, sessionMgr)
	expenseHandler := handlers.NewExpenseHandler(bot, apiClient, sessionMgr)
	trackingHandler := handlers.NewTrackingHandler(bot, apiClient, sessionMgr)
	dashboardHandler := handlers.NewDashboardHandler(bot, apiClient, sessionMgr)
	reportHandler := handlers.NewReportHandler(bot, apiClient, sessionMgr)
	pomodoroHandler := handlers.NewPomodoroHandler(bot, apiClient, sessionMgr)
	goalsHandler := handlers.NewGoalsHandler(bot, apiClient, sessionMgr)
	settingsHandler := handlers.NewSettingsHandler(bot, apiClient, sessionMgr)
	searchHandler := handlers.NewSearchHandler(bot, apiClient, sessionMgr)
	hunterHandler := handlers.NewHunterHandler(bot, apiClient, sessionMgr)
	gigHandler := handlers.NewGigHandler(bot, apiClient, sessionMgr)

	// Start listening for updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is running. Press Ctrl+C to stop.")

	for update := range updates {
		// Handle messages
		if update.Message != nil {
			if err := handleMessage(update.Message, bot, startHandler, helpHandler, invoiceHandler, clientHandler, companyHandler, contractHandler, expenseHandler, trackingHandler, dashboardHandler, reportHandler, pomodoroHandler, goalsHandler, settingsHandler, searchHandler, hunterHandler, gigHandler, sessionMgr); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}

		// Handle callback queries
		if update.CallbackQuery != nil {
			if err := handleCallback(update.CallbackQuery, bot, invoiceHandler, clientHandler, contractHandler, expenseHandler, trackingHandler, dashboardHandler, reportHandler, pomodoroHandler, goalsHandler, settingsHandler, searchHandler, hunterHandler, gigHandler, sessionMgr); err != nil {
				log.Printf("Error handling callback: %v", err)
			}
		}
	}
}

func handleMessage(
	message *tgbotapi.Message,
	bot *tgbotapi.BotAPI,
	startHandler *handlers.StartHandler,
	helpHandler *handlers.HelpHandler,
	invoiceHandler *handlers.InvoiceHandler,
	clientHandler *handlers.ClientHandler,
	companyHandler *handlers.CompanyHandler,
	contractHandler *handlers.ContractHandler,
	expenseHandler *handlers.ExpenseHandler,
	trackingHandler *handlers.TrackingHandler,
	dashboardHandler *handlers.DashboardHandler,
	reportHandler *handlers.ReportHandler,
	pomodoroHandler *handlers.PomodoroHandler,
	goalsHandler *handlers.GoalsHandler,
	settingsHandler *handlers.SettingsHandler,
	searchHandler *handlers.SearchHandler,
	hunterHandler *handlers.HunterHandler,
	gigHandler *handlers.GigHandler,
	sessionMgr *services.SessionManager,
) error {
	// Handle commands
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			return startHandler.Handle(message)
		case "help":
			return helpHandler.Handle(message)
		case "invoice":
			return invoiceHandler.HandleCreate(message)
		case "invoices":
			return invoiceHandler.HandleListWithPDF(message)
		case "pdf":
			return invoiceHandler.HandlePDF(message)
		case "client":
			return clientHandler.HandleCreate(message)
		case "clients":
			return clientHandler.HandleList(message)
		case "company":
			return companyHandler.HandleCreate(message)
		case "companies":
			return companyHandler.HandleList(message)
		case "contract":
			return contractHandler.HandleCreate(message)
		case "contracts":
			return contractHandler.HandleList(message)
		case "expense":
			return expenseHandler.HandleCreate(message)
		case "expenses":
			return expenseHandler.HandleList(message)
		case "track":
			return trackingHandler.HandleStart(message)
		case "tracking":
			return trackingHandler.HandleList(message)
		case "stop":
			return trackingHandler.HandleStop(message)
		case "active":
			return trackingHandler.HandleActive(message)
		case "log":
			return trackingHandler.HandleLog(message)
		case "dashboard":
			return dashboardHandler.Handle(message)
		// New commands
		case "report", "weekly":
			return reportHandler.HandleWeekly(message)
		case "monthly":
			return reportHandler.HandleMonthly(message)
		case "overdue":
			return reportHandler.HandleOverdue(message)
		case "unpaid":
			return reportHandler.HandleUnpaid(message)
		case "pomodoro":
			return pomodoroHandler.HandleStart(message)
		case "pomo":
			return pomodoroHandler.HandleActive(message)
		case "pomostats":
			return pomodoroHandler.HandleStats(message)
		case "goal", "goals":
			return goalsHandler.HandleStatus(message)
		case "rate":
			return goalsHandler.HandleRateAnalysis(message)
		case "settings":
			return settingsHandler.HandleShow(message)
		case "search":
			return searchHandler.HandleSearch(message)
		// Hunter commands
		case "hunter":
			return hunterHandler.HandleMenu(message)
		case "hunt":
			return hunterHandler.HandleHunt(message)
		case "jobs":
			return hunterHandler.HandleJobs(message)
		case "profile":
			return hunterHandler.HandleProfile(message)
		case "applications":
			return hunterHandler.HandleApplications(message)
		// Gig commands
		case "gig", "gigs":
			return gigHandler.HandleMenu(message)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Try /help")
			bot.Send(msg)
		}
		return nil
	}

	// Handle conversation states
	telegramID := message.From.ID
	session := sessionMgr.GetSession(telegramID)

	if session != nil {
		switch models.SessionState(session.State) {
		// Invoice states
		case models.StateInvoiceAmount:
			return invoiceHandler.HandleAmountInput(message)
		case models.StateInvoiceDescription:
			return invoiceHandler.HandleDescriptionInput(message)
		// Client states
		case models.StateClientCreateName:
			return clientHandler.HandleNameInput(message)
		case models.StateClientCreateEmail:
			return clientHandler.HandleEmailInput(message)
		case models.StateClientCreateAddress:
			return clientHandler.HandleAddressInput(message)
		case models.StateClientCreateTaxID:
			return clientHandler.HandleTaxIDInput(message)
		// Company states
		case models.StateCompanyCreateName:
			return companyHandler.HandleNameInput(message)
		case models.StateCompanyCreateEmail:
			return companyHandler.HandleEmailInput(message)
		case models.StateCompanyCreatePhone:
			return companyHandler.HandlePhoneInput(message)
		case models.StateCompanyCreateAddress:
			return companyHandler.HandleAddressInput(message)
		case models.StateCompanyCreateTaxID:
			return companyHandler.HandleTaxIDInput(message)
		// Contract states
		case models.StateContractName:
			return contractHandler.HandleNameInput(message)
		case models.StateContractRate:
			return contractHandler.HandleRateInput(message)
		// Expense states
		case models.StateExpenseDescription:
			return expenseHandler.HandleDescriptionInput(message)
		case models.StateExpenseAmount:
			return expenseHandler.HandleAmountInput(message)
		case models.StateExpenseVendor:
			return expenseHandler.HandleVendorInput(message)
		// Tracking log states
		case models.StateTrackLogHours:
			return trackingHandler.HandleHoursInput(message)
		case models.StateTrackLogProject:
			return trackingHandler.HandleProjectInput(message)
		case models.StateTrackLogNotes:
			return trackingHandler.HandleNotesInput(message)
		// Search states
		case models.StateSearchQuery:
			return searchHandler.HandleSearchQuery(message)
		// Hunter states
		case models.StateHunterProfileName:
			return hunterHandler.HandleNameInput(message)
		case models.StateHunterProfileTitle:
			return hunterHandler.HandleTitleInput(message)
		case models.StateHunterProfileSkills:
			return hunterHandler.HandleSkillsInput(message)
		case models.StateHunterProfileRate:
			return hunterHandler.HandleRateInput(message)
		case models.StateHunterAwaitingPDF:
			// Check if a document was sent
			if message.Document != nil {
				return hunterHandler.HandlePDFUpload(message)
			}
			// If not a document, remind user
			msg := tgbotapi.NewMessage(message.Chat.ID, "Please send your CV as a *PDF file*, or tap the button below to enter manually.")
			msg.ParseMode = "Markdown"
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("✍️ Enter manually", "hunter_profile_create"),
				),
			)
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			return nil
		// Gig states
		case models.StateGigCreateName:
			return gigHandler.HandleNameInput(message)
		case models.StateGigTaskAdd:
			return gigHandler.HandleTaskInput(message)
		}
	}

	// Default: show help
	msg := tgbotapi.NewMessage(message.Chat.ID, "I didn't understand that. Try /help for available commands.")
	_, err := bot.Send(msg)
	return err
}

func handleCallback(
	callbackQuery *tgbotapi.CallbackQuery,
	bot *tgbotapi.BotAPI,
	invoiceHandler *handlers.InvoiceHandler,
	clientHandler *handlers.ClientHandler,
	contractHandler *handlers.ContractHandler,
	expenseHandler *handlers.ExpenseHandler,
	trackingHandler *handlers.TrackingHandler,
	dashboardHandler *handlers.DashboardHandler,
	reportHandler *handlers.ReportHandler,
	pomodoroHandler *handlers.PomodoroHandler,
	goalsHandler *handlers.GoalsHandler,
	settingsHandler *handlers.SettingsHandler,
	searchHandler *handlers.SearchHandler,
	hunterHandler *handlers.HunterHandler,
	gigHandler *handlers.GigHandler,
	sessionMgr *services.SessionManager,
) error {
	data := callbackQuery.Data

	// Route based on callback data prefix
	// Invoice callbacks
	if strings.HasPrefix(data, "invoice_client_") {
		return invoiceHandler.HandleClientSelected(callbackQuery)
	}

	if strings.HasPrefix(data, "invoice_due_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			var days int
			switch parts[2] {
			case "7":
				days = 7
			case "14":
				days = 14
			case "30":
				days = 30
			default:
				days = 30
			}
			return invoiceHandler.CreateInvoice(callbackQuery, days)
		}
	}

	if data == "action_invoice" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		return invoiceHandler.HandleCreate(msg)
	}

	if data == "invoice_new_client" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		return clientHandler.HandleCreate(msg)
	}

	// Contract callbacks
	if strings.HasPrefix(data, "contract_client_") {
		return contractHandler.HandleClientSelected(callbackQuery)
	}

	if strings.HasPrefix(data, "contract_type_") {
		return contractHandler.HandleTypeSelected(callbackQuery)
	}

	// Expense callbacks
	if strings.HasPrefix(data, "expense_category_") {
		return expenseHandler.HandleCategorySelected(callbackQuery)
	}

	// Tracking callbacks
	if data == "tracking_stop" {
		return trackingHandler.HandleStopCallback(callbackQuery)
	}

	if strings.HasPrefix(data, "log_contract_") {
		return trackingHandler.HandleContractSelected(callbackQuery)
	}

	// Invoice PDF callbacks
	if strings.HasPrefix(data, "invoice_pdf_") {
		return invoiceHandler.HandlePDFCallback(callbackQuery)
	}

	// Report callbacks
	if data == "report_weekly" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading weekly report...")
		bot.Request(callback)
		return reportHandler.HandleWeekly(msg)
	}

	if data == "report_monthly" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading monthly report...")
		bot.Request(callback)
		return reportHandler.HandleMonthly(msg)
	}

	if data == "report_overdue" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading overdue report...")
		bot.Request(callback)
		return reportHandler.HandleOverdue(msg)
	}

	if data == "report_unpaid" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading unpaid report...")
		bot.Request(callback)
		return reportHandler.HandleUnpaid(msg)
	}

	// Pomodoro callbacks
	if strings.HasPrefix(data, "pomodoro_start_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			duration, _ := strconv.Atoi(parts[2])
			return pomodoroHandler.HandleStartDuration(callbackQuery, duration)
		}
	}

	if data == "pomodoro_active" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Checking pomodoro...")
		bot.Request(callback)
		return pomodoroHandler.HandleActive(msg)
	}

	if data == "action_pomodoro" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Starting pomodoro...")
		bot.Request(callback)
		return pomodoroHandler.HandleStart(msg)
	}

	// Goals callbacks
	if data == "goal_status" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading goals...")
		bot.Request(callback)
		return goalsHandler.HandleStatus(msg)
	}

	if data == "action_rate" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading rate analysis...")
		bot.Request(callback)
		return goalsHandler.HandleRateAnalysis(msg)
	}

	// Search callbacks
	if data == "action_search" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Search...")
		bot.Request(callback)
		return searchHandler.HandleSearch(msg)
	}

	// Auth callback
	if data == "auth_login" {
		chatID := callbackQuery.Message.Chat.ID
		text := "To authenticate, please visit:\n\n" +
			"https://ung.app/telegram/auth\n\n" +
			"And follow the instructions to link your account."

		msg := tgbotapi.NewMessage(chatID, text)
		bot.Send(msg)

		callback := tgbotapi.NewCallback(callbackQuery.ID, "Please visit the link to authenticate")
		bot.Request(callback)
		return nil
	}

	// Quick action callbacks from main menu
	if data == "action_clients" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading clients...")
		bot.Request(callback)
		return clientHandler.HandleList(msg)
	}

	if data == "action_track" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Starting tracking...")
		bot.Request(callback)
		return trackingHandler.HandleStart(msg)
	}

	if data == "action_reports" {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading dashboard...")
		bot.Request(callback)
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		return dashboardHandler.Handle(msg)
	}

	if data == "action_settings" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading settings...")
		bot.Request(callback)
		return settingsHandler.HandleShow(msg)
	}

	if data == "main_menu" {
		chatID := callbackQuery.Message.Chat.ID
		telegramID := callbackQuery.From.ID

		user := sessionMgr.GetUser(telegramID)
		name := "there"
		if user != nil {
			name = user.Name
		}

		text := fmt.Sprintf(
			"--------------------\n"+
				"      *Main Menu*\n"+
				"--------------------\n\n"+
				"Hey %s!\n\n"+
				"What would you like to do today?\n\n"+
				"_Tip: Use /help for all commands_",
			name,
		)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 Gigs", "action_gig"),
				tgbotapi.NewInlineKeyboardButtonData("⏱ Track", "action_track"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📄 Invoice", "action_invoice"),
				tgbotapi.NewInlineKeyboardButtonData("👤 Clients", "action_clients"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📝 Contracts", "action_contracts"),
				tgbotapi.NewInlineKeyboardButtonData("💰 Expenses", "action_expenses"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 Dashboard", "action_reports"),
				tgbotapi.NewInlineKeyboardButtonData("📈 Reports", "report_weekly"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🍅 Pomodoro", "action_pomodoro"),
				tgbotapi.NewInlineKeyboardButtonData("🎯 Hunter", "action_hunter"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎯 Goals", "goal_status"),
				tgbotapi.NewInlineKeyboardButtonData("🔍 Search", "action_search"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "action_settings"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

		callback := tgbotapi.NewCallback(callbackQuery.ID, "")
		bot.Request(callback)
		return nil
	}

	// New action callbacks
	if data == "action_contracts" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading contracts...")
		bot.Request(callback)
		return contractHandler.HandleList(msg)
	}

	if data == "action_expenses" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading expenses...")
		bot.Request(callback)
		return expenseHandler.HandleList(msg)
	}

	if data == "action_invoices_list" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading invoices...")
		bot.Request(callback)
		return invoiceHandler.HandleListWithPDF(msg)
	}

	if data == "action_log" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading...")
		bot.Request(callback)
		return trackingHandler.HandleLog(msg)
	}

	if data == "action_active" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Checking session...")
		bot.Request(callback)
		return trackingHandler.HandleActive(msg)
	}

	// Hunter callbacks
	if data == "action_hunter" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "")
		bot.Request(callback)
		return hunterHandler.HandleMenu(msg)
	}

	if data == "hunter_hunt" {
		return hunterHandler.HandleHuntCallback(callbackQuery)
	}

	if data == "hunter_jobs" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading jobs...")
		bot.Request(callback)
		return hunterHandler.HandleJobs(msg)
	}

	if data == "hunter_profile" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading profile...")
		bot.Request(callback)
		return hunterHandler.HandleProfile(msg)
	}

	if data == "hunter_profile_create" {
		return hunterHandler.HandleProfileCreate(callbackQuery)
	}

	if data == "hunter_profile_edit" {
		return hunterHandler.HandleProfileCreate(callbackQuery) // Same flow for edit
	}

	if data == "hunter_stats" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading stats...")
		bot.Request(callback)
		return hunterHandler.HandleStats(msg)
	}

	if data == "hunter_applications" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Loading applications...")
		bot.Request(callback)
		return hunterHandler.HandleApplications(msg)
	}

	if strings.HasPrefix(data, "hunter_job_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			jobID, _ := strconv.Atoi(parts[2])
			return hunterHandler.HandleJobDetail(callbackQuery, uint(jobID))
		}
	}

	if strings.HasPrefix(data, "hunter_proposal_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			jobID, _ := strconv.Atoi(parts[2])
			return hunterHandler.HandleGenerateProposal(callbackQuery, uint(jobID))
		}
	}

	if strings.HasPrefix(data, "hunter_apply_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			jobID, _ := strconv.Atoi(parts[2])
			return hunterHandler.HandleApply(callbackQuery, uint(jobID))
		}
	}

	// Gig callbacks
	if data == "action_gig" {
		msg := &tgbotapi.Message{
			Chat: callbackQuery.Message.Chat,
			From: callbackQuery.From,
		}
		callback := tgbotapi.NewCallback(callbackQuery.ID, "")
		bot.Request(callback)
		return gigHandler.HandleMenu(msg)
	}

	if data == "gig_create" {
		return gigHandler.HandleCreate(callbackQuery)
	}

	if data == "gig_list" {
		return gigHandler.HandleList(callbackQuery)
	}

	if strings.HasPrefix(data, "gig_filter_") {
		status := strings.TrimPrefix(data, "gig_filter_")
		return gigHandler.HandleFilter(callbackQuery, status)
	}

	if strings.HasPrefix(data, "gig_view_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			gigID, _ := strconv.Atoi(parts[2])
			return gigHandler.HandleView(callbackQuery, uint(gigID))
		}
	}

	if strings.HasPrefix(data, "gig_move_") {
		// gig_move_123_active
		parts := strings.Split(data, "_")
		if len(parts) == 4 {
			gigID, _ := strconv.Atoi(parts[2])
			status := parts[3]
			return gigHandler.HandleMove(callbackQuery, uint(gigID), status)
		}
	}

	if strings.HasPrefix(data, "gig_delete_") {
		parts := strings.Split(data, "_")
		if len(parts) == 3 {
			gigID, _ := strconv.Atoi(parts[2])
			return gigHandler.HandleDelete(callbackQuery, uint(gigID))
		}
	}

	if strings.HasPrefix(data, "gig_task_add_") {
		parts := strings.Split(data, "_")
		if len(parts) == 4 {
			gigID, _ := strconv.Atoi(parts[3])
			return gigHandler.HandleTaskAdd(callbackQuery, uint(gigID))
		}
	}

	if strings.HasPrefix(data, "gig_task_toggle_") {
		parts := strings.Split(data, "_")
		if len(parts) == 4 {
			taskID, _ := strconv.Atoi(parts[3])
			return gigHandler.HandleTaskToggle(callbackQuery, uint(taskID))
		}
	}

	// Answer callback with default response
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	_, err := bot.Request(callback)
	return err
}
