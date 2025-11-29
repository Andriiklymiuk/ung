package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ung-telegram/internal/models"
	"ung-telegram/internal/services"
)

// SearchHandler handles search commands
type SearchHandler struct {
	bot        *tgbotapi.BotAPI
	apiClient  *services.APIClient
	sessionMgr *services.SessionManager
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(bot *tgbotapi.BotAPI, apiClient *services.APIClient, sessionMgr *services.SessionManager) *SearchHandler {
	return &SearchHandler{
		bot:        bot,
		apiClient:  apiClient,
		sessionMgr: sessionMgr,
	}
}

// HandleSearch initiates search flow
func (h *SearchHandler) HandleSearch(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	if !h.sessionMgr.IsAuthenticated(telegramID) {
		msg := tgbotapi.NewMessage(chatID, "Please authenticate first with /start")
		h.bot.Send(msg)
		return nil
	}

	// Check if query is provided with command
	args := strings.Fields(message.Text)
	if len(args) > 1 {
		query := strings.Join(args[1:], " ")
		return h.performSearch(chatID, telegramID, query)
	}

	// Set state for search query input
	h.sessionMgr.SetSession(&models.Session{
		TelegramID: telegramID,
		State:      string(models.StateSearchQuery),
		Data:       make(map[string]interface{}),
	})

	msg := tgbotapi.NewMessage(chatID,
		"--------------------\n"+
			"      *Search*\n"+
			"--------------------\n\n"+
			"Enter your search query:\n\n"+
			"_Search across clients, invoices, contracts, and more._")
	msg.ParseMode = "Markdown"

	h.bot.Send(msg)
	return nil
}

// HandleSearchQuery handles the search query input
func (h *SearchHandler) HandleSearchQuery(message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	query := strings.TrimSpace(message.Text)

	if query == "" {
		msg := tgbotapi.NewMessage(chatID, "Please enter a search query")
		h.bot.Send(msg)
		return nil
	}

	// Clear session
	h.sessionMgr.ClearSession(telegramID)

	return h.performSearch(chatID, telegramID, query)
}

func (h *SearchHandler) performSearch(chatID int64, telegramID int64, query string) error {
	user := h.sessionMgr.GetUser(telegramID)

	results, err := h.apiClient.Search(user.APIToken, query)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Search failed: "+err.Error())
		h.bot.Send(msg)
		return err
	}

	var text strings.Builder
	text.WriteString("--------------------\n")
	text.WriteString("      *Search Results*\n")
	text.WriteString("--------------------\n\n")
	text.WriteString(fmt.Sprintf("Query: _%s_\n\n", results.Query))

	if len(results.Results) == 0 {
		text.WriteString("No results found.\n\n")
		text.WriteString("Try a different search term.")
	} else {
		// Show counts by type
		text.WriteString("*Found:*\n")
		for typ, count := range results.Counts {
			if count > 0 {
				emoji := getTypeEmoji(typ)
				text.WriteString(fmt.Sprintf("%s %s: %d\n", emoji, typ, count))
			}
		}
		text.WriteString("\n")

		// Show results (max 8)
		text.WriteString("*Results:*\n\n")
		shown := 0
		for _, result := range results.Results {
			if shown >= 8 {
				text.WriteString(fmt.Sprintf("\n_...and %d more results_", len(results.Results)-8))
				break
			}

			emoji := getTypeEmoji(result.Type)
			text.WriteString(fmt.Sprintf("%s *%s*\n", emoji, result.Title))
			if result.Subtitle != "" {
				text.WriteString(fmt.Sprintf("   %s\n", result.Subtitle))
			}
			text.WriteString("\n")
			shown++
		}
	}

	text.WriteString("\n--------------------\n")

	msg := tgbotapi.NewMessage(chatID, text.String())
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("New Search", "action_search"),
			tgbotapi.NewInlineKeyboardButtonData("Menu", "main_menu"),
		),
	)
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
	return nil
}

func getTypeEmoji(typ string) string {
	switch typ {
	case "client":
		return "👤"
	case "invoice":
		return "📄"
	case "contract":
		return "📋"
	case "expense":
		return "💸"
	case "company":
		return "🏢"
	case "tracking":
		return "⏱️"
	default:
		return "📌"
	}
}
