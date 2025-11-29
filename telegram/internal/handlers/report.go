package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/services"
)

// ReportHandler handles report commands
type ReportHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewReportHandler creates a new report handler
func NewReportHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *ReportHandler {
	return &ReportHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleWeekly shows weekly report
func (h *ReportHandler) HandleWeekly(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	report, err := h.apiClient.GetWeeklyReport(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch weekly report: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Weekly Report*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*%s* to *%s*\n\n", report.WeekStart, report.WeekEnd))
	text.WriteString(fmt.Sprintf("*Total Hours:* %.1f hrs\n", report.TotalHours))
	text.WriteString(fmt.Sprintf("*Total Revenue:* $%.2f\n", report.TotalRevenue))
	text.WriteString(fmt.Sprintf("*Sessions:* %d\n\n", report.Sessions))
	text.WriteString("--------------------\n")
	text.WriteString("_Updated just now_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Monthly", "report_monthly"),
			tgbotapi.NewInlineKeyboardButtonData("Overdue", "report_overdue"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleMonthly shows monthly report
func (h *ReportHandler) HandleMonthly(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	report, err := h.apiClient.GetMonthlyReport(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch monthly report: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Monthly Report*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("*%s %d*\n\n", report.Month, report.Year))
	text.WriteString(fmt.Sprintf("*Total Hours:* %.1f hrs\n", report.TotalHours))
	text.WriteString(fmt.Sprintf("*Revenue:* $%.2f\n", report.TotalRevenue))
	text.WriteString(fmt.Sprintf("*Expenses:* $%.2f\n", report.TotalExpenses))
	text.WriteString(fmt.Sprintf("*Profit:* $%.2f\n\n", report.Profit))
	text.WriteString(fmt.Sprintf("*Invoices Sent:* %d\n", report.InvoicesSent))
	text.WriteString(fmt.Sprintf("*Invoices Paid:* %d\n\n", report.InvoicesPaid))
	text.WriteString("--------------------\n")
	text.WriteString("_Updated just now_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Weekly", "report_weekly"),
			tgbotapi.NewInlineKeyboardButtonData("Unpaid", "report_unpaid"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleOverdue shows overdue invoices report
func (h *ReportHandler) HandleOverdue(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	report, err := h.apiClient.GetOverdueReport(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch overdue report: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Overdue Invoices*\n")
	text.WriteString("--------------------\n\n")

	if report.Count == 0 {
		text.WriteString("No overdue invoices!\n\n")
		text.WriteString("All payments are on track.\n")
	} else {
		text.WriteString(fmt.Sprintf("*Count:* %d invoices\n", report.Count))
		text.WriteString(fmt.Sprintf("*Total Overdue:* $%.2f\n\n", report.TotalOverdue))
		text.WriteString("_Consider following up with clients._\n")
	}

	text.WriteString("\n--------------------\n")
	text.WriteString("_Updated just now_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Unpaid", "report_unpaid"),
			tgbotapi.NewInlineKeyboardButtonData("Weekly", "report_weekly"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

// HandleUnpaid shows unpaid invoices report
func (h *ReportHandler) HandleUnpaid(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	report, err := h.apiClient.GetUnpaidReport(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Failed to fetch unpaid report: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Unpaid Invoices*\n")
	text.WriteString("--------------------\n\n")

	if report.Count == 0 {
		text.WriteString("No unpaid invoices!\n\n")
		text.WriteString("All invoices have been paid.\n")
	} else {
		text.WriteString(fmt.Sprintf("*Count:* %d invoices\n\n", report.Count))
		text.WriteString(fmt.Sprintf("*Total Unpaid:* $%.2f\n", report.TotalUnpaid))
		text.WriteString(fmt.Sprintf("  Pending: $%.2f\n", report.Pending))
		text.WriteString(fmt.Sprintf("  Overdue: $%.2f\n", report.Overdue))
	}

	text.WriteString("\n--------------------\n")
	text.WriteString("_Updated just now_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Overdue", "report_overdue"),
			tgbotapi.NewInlineKeyboardButtonData("Monthly", "report_monthly"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}
