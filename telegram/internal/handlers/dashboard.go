package handlers

import (
	"fmt"
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
	text.WriteString("📊 *Revenue Dashboard*\n\n")

	text.WriteString(fmt.Sprintf("💰 *Total Monthly Revenue:* $%.2f\n\n", projection.TotalMonthlyRevenue))

	text.WriteString("📈 *Breakdown:*\n")
	text.WriteString(fmt.Sprintf("  • Hourly: $%.2f\n", projection.HourlyContractsRevenue))
	text.WriteString(fmt.Sprintf("  • Retainer: $%.2f\n", projection.RetainerRevenue))
	text.WriteString(fmt.Sprintf("  • Projects: $%.2f\n",
		projection.TotalMonthlyRevenue-projection.HourlyContractsRevenue-projection.RetainerRevenue))
	text.WriteString("\n")

	text.WriteString(fmt.Sprintf("⏱️ *Projected Hours:* %.1f hours\n", projection.ProjectedHours))
	if projection.AverageHourlyRate > 0 {
		text.WriteString(fmt.Sprintf("💵 *Average Rate:* $%.0f/hr\n", projection.AverageHourlyRate))
	}
	text.WriteString("\n")

	text.WriteString(fmt.Sprintf("📝 *Active Contracts:* %d\n\n", projection.ActiveContracts))

	// Top contracts
	if len(projection.ContractBreakdown) > 0 {
		text.WriteString("*Top Contracts:*\n")
		count := 0
		for _, contract := range projection.ContractBreakdown {
			if count >= 5 {
				text.WriteString(fmt.Sprintf("\n_...and %d more_", len(projection.ContractBreakdown)-5))
				break
			}
			if contract.MonthlyRevenue > 0 {
				text.WriteString(fmt.Sprintf("• %s (%s): $%.2f/mo\n",
					contract.ClientName, contract.ContractType, contract.MonthlyRevenue))
				count++
			}
		}
	}

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	_, err = h.bot.Send(msg)
	return err
}
