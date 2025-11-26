package handlers

import (
	"fmt"
	"math"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/services"
)

// DashboardHandler handles dashboard commands
type DashboardHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *DashboardHandler {
	return &DashboardHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// Handle shows the revenue dashboard
func (h *DashboardHandler) Handle(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check authentication
	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "🔒 Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	user := h.sessionMgr.GetUser(telegramID)

	// Fetch dashboard data
	projection, err := h.apiClient.GetDashboard(user.APIToken)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Failed to fetch dashboard: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	// Build dashboard message
	var text strings.Builder
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString("      📊 *Dashboard*\n")
	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Main revenue highlight
	text.WriteString("💰 *Monthly Revenue*\n")
	text.WriteString(fmt.Sprintf("┌─────────────────────┐\n"))
	text.WriteString(fmt.Sprintf("│    *$%.2f*    │\n", projection.TotalMonthlyRevenue))
	text.WriteString(fmt.Sprintf("└─────────────────────┘\n\n"))

	// Revenue breakdown with visual bars
	text.WriteString("📈 *Revenue Breakdown*\n\n")

	hourlyPct := 0.0
	retainerPct := 0.0
	projectPct := 0.0

	if projection.TotalMonthlyRevenue > 0 {
		hourlyPct = (projection.HourlyContractsRevenue / projection.TotalMonthlyRevenue) * 100
		retainerPct = (projection.RetainerRevenue / projection.TotalMonthlyRevenue) * 100
		projectPct = 100 - hourlyPct - retainerPct
	}

	projectRevenue := projection.TotalMonthlyRevenue - projection.HourlyContractsRevenue - projection.RetainerRevenue

	text.WriteString(fmt.Sprintf("⏰ *Hourly*: $%.2f\n", projection.HourlyContractsRevenue))
	text.WriteString(fmt.Sprintf("   %s %.0f%%\n\n", generateProgressBar(hourlyPct), hourlyPct))

	text.WriteString(fmt.Sprintf("🔄 *Retainer*: $%.2f\n", projection.RetainerRevenue))
	text.WriteString(fmt.Sprintf("   %s %.0f%%\n\n", generateProgressBar(retainerPct), retainerPct))

	text.WriteString(fmt.Sprintf("📁 *Projects*: $%.2f\n", projectRevenue))
	text.WriteString(fmt.Sprintf("   %s %.0f%%\n", generateProgressBar(projectPct), projectPct))

	text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Quick stats
	text.WriteString("📋 *Quick Stats*\n\n")
	text.WriteString(fmt.Sprintf("⏱️ Projected Hours: *%.0f hrs*\n", math.Ceil(projection.ProjectedHours)))
	if projection.AverageHourlyRate > 0 {
		text.WriteString(fmt.Sprintf("💵 Avg Rate: *$%.0f/hr*\n", projection.AverageHourlyRate))
	}
	text.WriteString(fmt.Sprintf("📝 Active Contracts: *%d*\n", projection.ActiveContracts))

	// Top contracts
	if len(projection.ContractBreakdown) > 0 {
		text.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━\n\n")
		text.WriteString("🏆 *Top Contracts*\n\n")
		count := 0
		for _, contract := range projection.ContractBreakdown {
			if count >= 5 {
				remaining := len(projection.ContractBreakdown) - 5
				if remaining > 0 {
					text.WriteString(fmt.Sprintf("\n_+%d more contracts_", remaining))
				}
				break
			}
			if contract.MonthlyRevenue > 0 {
				typeEmoji := "📄"
				switch contract.ContractType {
				case "hourly":
					typeEmoji = "⏰"
				case "retainer":
					typeEmoji = "🔄"
				case "project":
					typeEmoji = "📁"
				}
				text.WriteString(fmt.Sprintf("%s *%s*\n", typeEmoji, contract.ClientName))
				text.WriteString(fmt.Sprintf("   $%.2f/mo · %s\n\n", contract.MonthlyRevenue, contract.ContractType))
				count++
			}
		}
	}

	text.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	text.WriteString("_Updated just now_")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "action_reports"),
			tgbotapi.NewInlineKeyboardButtonData("📄 Invoices", "action_invoices_list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏱️ Track Time", "action_track"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err = h.bot.Send(msg)
	return err
}

// generateProgressBar creates a visual progress bar
func generateProgressBar(percentage float64) string {
	filled := int(percentage / 10)
	if filled > 10 {
		filled = 10
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < 10; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
