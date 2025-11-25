package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/spf13/cobra"
)

var goalCmd = &cobra.Command{
	Use:   "goal",
	Short: "Track income goals",
	Long: `Set and track income goals for various time periods.

Commands:
  set       Set an income goal
  ls        List all goals
  status    Show progress toward goals
  rm        Remove a goal`,
}

var goalSetCmd = &cobra.Command{
	Use:   "set <amount>",
	Short: "Set an income goal",
	Args:  cobra.ExactArgs(1),
	RunE:  runGoalSet,
}

var goalListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all income goals",
	RunE:  runGoalList,
}

var goalStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show progress toward goals",
	RunE:  runGoalStatus,
}

var goalRemoveCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove a goal",
	Args:  cobra.ExactArgs(1),
	RunE:  runGoalRemove,
}

var (
	goalPeriod      string
	goalYear        int
	goalMonth       int
	goalDescription string
)

func init() {
	goalSetCmd.Flags().StringVarP(&goalPeriod, "period", "p", "monthly", "Goal period: monthly, quarterly, yearly")
	goalSetCmd.Flags().IntVar(&goalYear, "year", 0, "Target year (defaults to current)")
	goalSetCmd.Flags().IntVar(&goalMonth, "month", 0, "Target month for monthly goals (1-12)")
	goalSetCmd.Flags().StringVarP(&goalDescription, "description", "d", "", "Goal description")

	goalStatusCmd.Flags().StringVarP(&goalPeriod, "period", "p", "", "Filter by period: monthly, quarterly, yearly")

	rootCmd.AddCommand(goalCmd)
	goalCmd.AddCommand(goalSetCmd)
	goalCmd.AddCommand(goalListCmd)
	goalCmd.AddCommand(goalStatusCmd)
	goalCmd.AddCommand(goalRemoveCmd)
}

type IncomeGoal struct {
	ID          uint      `gorm:"primaryKey"`
	Amount      float64   `gorm:"not null"`
	Period      string    `gorm:"not null"` // monthly, quarterly, yearly
	Year        int       `gorm:"not null"`
	Month       int       // for monthly goals
	Quarter     int       // for quarterly goals
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func init() {
	// Migrate the schema
	db.OnConnect(func() {
		db.DB.AutoMigrate(&IncomeGoal{})
	})
}

func runGoalSet(cmd *cobra.Command, args []string) error {
	var amount float64
	_, err := fmt.Sscanf(args[0], "%f", &amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %s", args[0])
	}

	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	now := time.Now()
	year := goalYear
	if year == 0 {
		year = now.Year()
	}

	month := goalMonth
	if month == 0 && goalPeriod == "monthly" {
		month = int(now.Month())
	}

	quarter := 0
	if goalPeriod == "quarterly" {
		quarter = (int(now.Month())-1)/3 + 1
	}

	// Check for existing goal
	var existing IncomeGoal
	query := db.DB.Where("period = ? AND year = ?", goalPeriod, year)
	if goalPeriod == "monthly" {
		query = query.Where("month = ?", month)
	} else if goalPeriod == "quarterly" {
		query = query.Where("quarter = ?", quarter)
	}

	if query.First(&existing).Error == nil {
		// Update existing
		existing.Amount = amount
		existing.Description = goalDescription
		if err := db.DB.Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update goal: %w", err)
		}
		fmt.Printf("Updated %s goal for %s: $%.2f\n", goalPeriod, formatGoalPeriod(year, month, quarter, goalPeriod), amount)
	} else {
		// Create new
		goal := IncomeGoal{
			Amount:      amount,
			Period:      goalPeriod,
			Year:        year,
			Month:       month,
			Quarter:     quarter,
			Description: goalDescription,
		}

		if err := db.DB.Create(&goal).Error; err != nil {
			return fmt.Errorf("failed to create goal: %w", err)
		}
		fmt.Printf("Created %s goal for %s: $%.2f\n", goalPeriod, formatGoalPeriod(year, month, quarter, goalPeriod), amount)
	}

	return nil
}

func runGoalList(cmd *cobra.Command, args []string) error {
	var goals []IncomeGoal
	if err := db.DB.Order("year DESC, period, month DESC, quarter DESC").Find(&goals).Error; err != nil {
		return fmt.Errorf("failed to list goals: %w", err)
	}

	if len(goals) == 0 {
		fmt.Println("No income goals set.")
		fmt.Println("\nUse 'ung goal set <amount>' to create one.")
		return nil
	}

	fmt.Println("Income Goals")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPERIOD\tTARGET\tAMOUNT\tDESCRIPTION")

	for _, g := range goals {
		period := formatGoalPeriod(g.Year, g.Month, g.Quarter, g.Period)
		fmt.Fprintf(w, "%d\t%s\t%s\t$%.2f\t%s\n",
			g.ID, g.Period, period, g.Amount, g.Description)
	}
	w.Flush()

	return nil
}

func runGoalStatus(cmd *cobra.Command, args []string) error {
	var goals []IncomeGoal
	query := db.DB.Order("year DESC, period, month DESC, quarter DESC")
	if goalPeriod != "" {
		query = query.Where("period = ?", goalPeriod)
	}
	if err := query.Find(&goals).Error; err != nil {
		return fmt.Errorf("failed to list goals: %w", err)
	}

	if len(goals) == 0 {
		fmt.Println("No income goals set.")
		return nil
	}

	fmt.Println()
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│           INCOME GOAL STATUS            │\n")
	fmt.Printf("╰─────────────────────────────────────────╯\n")

	for _, g := range goals {
		startDate, endDate := getGoalDateRange(g)
		actual := getIncomeForPeriod(startDate, endDate)
		progress := safePercent(actual, g.Amount)

		fmt.Println()
		fmt.Printf("  %s - %s\n", g.Period, formatGoalPeriod(g.Year, g.Month, g.Quarter, g.Period))
		if g.Description != "" {
			fmt.Printf("  %s\n", g.Description)
		}
		fmt.Println("  ─────────────────────────────────────")

		// Progress bar
		barWidth := 30
		filled := int(progress / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		bar := "│" + repeatStr("█", filled) + repeatStr("░", barWidth-filled) + "│"

		fmt.Printf("  %s %.1f%%\n", bar, progress)
		fmt.Printf("  $%.2f / $%.2f\n", actual, g.Amount)

		remaining := g.Amount - actual
		if remaining > 0 {
			daysLeft := getDaysRemaining(endDate)
			if daysLeft > 0 {
				dailyNeeded := remaining / float64(daysLeft)
				fmt.Printf("  $%.0f remaining (%d days left, ~$%.0f/day needed)\n", remaining, daysLeft, dailyNeeded)
			} else {
				fmt.Printf("  $%.0f remaining (period ended)\n", remaining)
			}
		} else {
			fmt.Printf("  Goal achieved! +$%.0f over target\n", -remaining)
		}
	}

	fmt.Println()
	return nil
}

func runGoalRemove(cmd *cobra.Command, args []string) error {
	var id uint
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid ID: %s", args[0])
	}

	result := db.DB.Delete(&IncomeGoal{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete goal: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("goal not found: %d", id)
	}

	fmt.Printf("Removed goal %d\n", id)
	return nil
}

func formatGoalPeriod(year, month, quarter int, period string) string {
	switch period {
	case "monthly":
		return time.Month(month).String()[:3] + fmt.Sprintf(" %d", year)
	case "quarterly":
		return fmt.Sprintf("Q%d %d", quarter, year)
	case "yearly":
		return fmt.Sprintf("%d", year)
	}
	return ""
}

func getGoalDateRange(g IncomeGoal) (time.Time, time.Time) {
	loc := time.Now().Location()

	switch g.Period {
	case "monthly":
		start := time.Date(g.Year, time.Month(g.Month), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 1, 0)
		return start, end

	case "quarterly":
		startMonth := (g.Quarter-1)*3 + 1
		start := time.Date(g.Year, time.Month(startMonth), 1, 0, 0, 0, 0, loc)
		end := start.AddDate(0, 3, 0)
		return start, end

	case "yearly":
		start := time.Date(g.Year, 1, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(1, 0, 0)
		return start, end
	}

	return time.Time{}, time.Time{}
}

func getIncomeForPeriod(start, end time.Time) float64 {
	// Sum of paid invoices in the period
	var total sql.NullFloat64
	db.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0)
		FROM invoices
		WHERE status = ?
		  AND updated_at >= ?
		  AND updated_at < ?
	`, models.StatusPaid, start, end).Scan(&total)

	return total.Float64
}

func getDaysRemaining(end time.Time) int {
	now := time.Now()
	if now.After(end) {
		return 0
	}
	return int(end.Sub(now).Hours() / 24)
}
