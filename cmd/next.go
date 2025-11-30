package cmd

import (
	"fmt"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/internal/repository"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "What's your next move?",
	Long:  "Shows your next actions: work to do, invoices to send, gigs to complete, and goal progress",
	RunE:  runNext,
}

var nextShowAll bool

func init() {
	nextCmd.Flags().BoolVarP(&nextShowAll, "all", "a", false, "Show all sections including empty ones")
	rootCmd.AddCommand(nextCmd)
}

func runNext(cmd *cobra.Command, args []string) error {
	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3366FF")).
		MarginBottom(1)

	cardTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#888888")).
		MarginBottom(1)

	valueStyle := lipgloss.NewStyle().
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22CC55"))

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF9933"))

	// Print header
	hour := time.Now().Hour()
	greeting := "Good evening"
	if hour < 12 {
		greeting = "Good morning"
	} else if hour < 17 {
		greeting = "Good afternoon"
	}

	fmt.Println()
	fmt.Println(headerStyle.Render(greeting + " - What's your next move?"))
	fmt.Println()

	// Initialize repositories
	trackingRepo := repository.NewTrackingSessionRepository()
	invoiceRepo := repository.NewInvoiceRepository()
	contractRepo := repository.NewContractRepository()

	// ========== WORK ==========
	fmt.Println(cardTitleStyle.Render("WORK"))

	// Check for active tracking session
	activeSession, _ := trackingRepo.GetActiveSession()
	if activeSession != nil {
		duration := time.Since(activeSession.StartTime)
		hours := int(duration.Hours())
		mins := int(duration.Minutes()) % 60

		fmt.Printf("  %s Tracking: %s\n", successStyle.Render("●"), activeSession.ProjectName)
		fmt.Printf("  %s %dh %02dm\n", mutedStyle.Render("  Duration:"), hours, mins)
		fmt.Println()
	} else {
		// Show most recent contract to work on
		contracts, _ := contractRepo.ListActive()
		if len(contracts) > 0 {
			fmt.Printf("  Ready to work?\n")
			fmt.Printf("  %s Start tracking with: ung track start\n", mutedStyle.Render("→"))
			fmt.Println()
		} else {
			fmt.Printf("  %s\n", mutedStyle.Render("No active contracts"))
			fmt.Println()
		}
	}

	// ========== BILL ==========
	fmt.Println(cardTitleStyle.Render("BILL"))

	// Check pending invoices
	pendingInvoices, _ := invoiceRepo.GetByStatus(models.StatusPending)
	sentInvoices, _ := invoiceRepo.GetByStatus(models.StatusSent)
	overdueInvoices, _ := invoiceRepo.GetByStatus(models.StatusOverdue)

	allPending := append(pendingInvoices, sentInvoices...)
	allPending = append(allPending, overdueInvoices...)

	var pendingAmount float64
	for _, inv := range allPending {
		pendingAmount += inv.Amount
	}

	var overdueAmount float64
	for _, inv := range overdueInvoices {
		overdueAmount += inv.Amount
	}

	if pendingAmount > 0 {
		fmt.Printf("  Pending: %s\n", valueStyle.Render(fmt.Sprintf("$%.2f", pendingAmount)))
		if len(overdueInvoices) > 0 {
			fmt.Printf("  %s %d overdue ($%.2f)\n",
				warningStyle.Render("!"),
				len(overdueInvoices),
				overdueAmount)
		}
	}

	// Check unbilled hours this month
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	sessions, _ := trackingRepo.List()
	var unbilledHours float64
	for _, s := range sessions {
		if s.StartTime.After(startOfMonth) && s.EndTime != nil && s.Billable {
			if s.Hours != nil {
				unbilledHours += *s.Hours
			}
		}
	}

	if unbilledHours > 0 {
		fmt.Printf("  Unbilled: %s this month\n",
			valueStyle.Render(fmt.Sprintf("%.1fh", unbilledHours)))
		fmt.Printf("  %s Create invoice: ung invoice new\n", mutedStyle.Render("→"))
	}

	if pendingAmount == 0 && unbilledHours == 0 {
		fmt.Printf("  %s\n", mutedStyle.Render("All caught up!"))
	}
	fmt.Println()

	// ========== GOAL ==========
	fmt.Println(cardTitleStyle.Render("GOAL"))

	// Get current month's paid invoices
	paidInvoices, _ := invoiceRepo.GetByStatus(models.StatusPaid)
	var monthlyRevenue float64
	for _, inv := range paidInvoices {
		if inv.IssuedDate.After(startOfMonth) {
			monthlyRevenue += inv.Amount
		}
	}

	// Try to get goal from database using GORM directly
	var goal IncomeGoal
	year := time.Now().Year()
	month := int(time.Now().Month())

	err := db.GormDB.Where("period = ? AND year = ? AND month = ?", "monthly", year, month).First(&goal).Error
	hasGoal := err == nil && goal.Amount > 0

	if hasGoal {
		progress := monthlyRevenue / goal.Amount

		// Progress bar
		barWidth := 20
		filled := int(progress * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}

		bar := ""
		for i := 0; i < barWidth; i++ {
			if i < filled {
				bar += "█"
			} else {
				bar += "░"
			}
		}

		// Days remaining
		daysInMonth := time.Date(time.Now().Year(), time.Now().Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
		daysRemaining := daysInMonth - time.Now().Day()

		progressColor := successStyle
		if progress < 0.5 && float64(time.Now().Day())/float64(daysInMonth) > 0.5 {
			progressColor = warningStyle
		}

		fmt.Printf("  %s\n", time.Now().Format("January 2006"))
		fmt.Printf("  [%s] %s\n", bar, progressColor.Render(fmt.Sprintf("%.0f%%", progress*100)))
		fmt.Printf("  $%.0f / $%.0f  •  %d days left\n", monthlyRevenue, goal.Amount, daysRemaining)
	} else {
		fmt.Printf("  This month: %s\n", valueStyle.Render(fmt.Sprintf("$%.2f", monthlyRevenue)))
		fmt.Printf("  %s Set a goal: ung goal set <amount>\n", mutedStyle.Render("→"))
	}
	fmt.Println()

	// ========== GIGS ==========
	fmt.Println(cardTitleStyle.Render("GIGS"))

	var gigs []models.Gig
	db.GormDB.Where("status IN ?", []string{"todo", "in_progress", "sent"}).
		Order("CASE status WHEN 'in_progress' THEN 1 WHEN 'sent' THEN 2 WHEN 'todo' THEN 3 END, updated_at DESC").
		Limit(5).
		Find(&gigs)

	if len(gigs) > 0 {
		statusColors := map[models.GigStatus]lipgloss.Style{
			models.GigStatusTodo:       mutedStyle,
			models.GigStatusInProgress: valueStyle,
			models.GigStatusSent:       warningStyle,
		}

		statusIcons := map[models.GigStatus]string{
			models.GigStatusTodo:       "○",
			models.GigStatusInProgress: "●",
			models.GigStatusSent:       "◉",
		}

		for _, gig := range gigs {
			style := statusColors[gig.Status]
			icon := statusIcons[gig.Status]
			fmt.Printf("  %s %s %s\n", style.Render(icon), gig.Name, mutedStyle.Render(fmt.Sprintf("(%s)", gig.Status)))
		}

		// Count by status
		var todoCount, inProgressCount, sentCount int64
		db.GormDB.Model(&models.Gig{}).Where("status = ?", "todo").Count(&todoCount)
		db.GormDB.Model(&models.Gig{}).Where("status = ?", "in_progress").Count(&inProgressCount)
		db.GormDB.Model(&models.Gig{}).Where("status = ?", "sent").Count(&sentCount)

		fmt.Println()
		fmt.Printf("  %s todo: %d  |  in_progress: %d  |  sent: %d\n",
			mutedStyle.Render(""),
			todoCount, inProgressCount, sentCount)
		fmt.Printf("  %s ung gig list    - See all gigs\n", mutedStyle.Render("→"))
	} else {
		fmt.Printf("  %s\n", mutedStyle.Render("No active gigs"))
		fmt.Printf("  %s ung gig add <name>  - Create a gig\n", mutedStyle.Render("→"))
	}
	fmt.Println()

	// ========== QUICK ACTIONS ==========
	fmt.Println(cardTitleStyle.Render("QUICK ACTIONS"))

	// Contextual actions based on state
	if activeSession == nil {
		fmt.Printf("  %s ung track start      Start tracking time\n", successStyle.Render("→"))
	} else {
		fmt.Printf("  %s ung track stop       Stop current session\n", warningStyle.Render("→"))
	}

	if unbilledHours > 0 {
		fmt.Printf("  %s ung invoice new      Create invoice (%.1fh unbilled)\n", warningStyle.Render("→"), unbilledHours)
	} else {
		fmt.Printf("  %s ung invoice new      Create an invoice\n", mutedStyle.Render("→"))
	}

	if len(gigs) == 0 {
		fmt.Printf("  %s ung gig add <name>   Add a gig to track\n", mutedStyle.Render("→"))
	} else {
		fmt.Printf("  %s ung gig list         View your gigs\n", mutedStyle.Render("→"))
	}

	if !hasGoal {
		fmt.Printf("  %s ung goal set <amt>   Set income goal\n", mutedStyle.Render("→"))
	} else {
		fmt.Printf("  %s ung goal status      Check goal progress\n", mutedStyle.Render("→"))
	}

	fmt.Println()

	return nil
}
