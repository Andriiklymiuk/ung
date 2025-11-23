package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var expenseCmd = &cobra.Command{
	Use:   "expense",
	Short: "Track business expenses",
	Long:  "Add, list, and manage business expenses",
}

var expenseAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new expense",
	RunE:  runExpenseAdd,
}

var expenseListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List expenses",
	RunE:    runExpenseList,
}

var expenseReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Expense report summary",
	RunE:  runExpenseReport,
}

func init() {
	rootCmd.AddCommand(expenseCmd)
	expenseCmd.AddCommand(expenseAddCmd)
	expenseCmd.AddCommand(expenseListCmd)
	expenseCmd.AddCommand(expenseReportCmd)
}

func runExpenseAdd(cmd *cobra.Command, args []string) error {
	fmt.Println("💸 Add Business Expense\n")

	var description, vendor, notes, dateStr, amountStr string
	var category string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Description").
				Placeholder("e.g., Adobe Creative Cloud subscription").
				Value(&description).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("description is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Amount").
				Placeholder("e.g., 52.99").
				Value(&amountStr).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("amount is required")
					}
					_, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return fmt.Errorf("invalid amount")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Category").
				Options(
					huh.NewOption("Software", "software"),
					huh.NewOption("Hardware", "hardware"),
					huh.NewOption("Travel", "travel"),
					huh.NewOption("Meals", "meals"),
					huh.NewOption("Office Supplies", "office_supplies"),
					huh.NewOption("Utilities", "utilities"),
					huh.NewOption("Marketing", "marketing"),
					huh.NewOption("Other", "other"),
				).
				Value(&category),

			huh.NewInput().
				Title("Vendor (optional)").
				Placeholder("e.g., Adobe").
				Value(&vendor),

			huh.NewInput().
				Title("Date (YYYY-MM-DD, leave empty for today)").
				Placeholder(time.Now().Format("2006-01-02")).
				Value(&dateStr),

			huh.NewText().
				Title("Notes (optional)").
				Value(&notes),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("form cancelled: %w", err)
	}

	// Parse amount
	amount, _ := strconv.ParseFloat(amountStr, 64)

	// Parse date
	var expenseDate time.Time
	if dateStr == "" {
		expenseDate = time.Now()
	} else {
		var err error
		expenseDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
	}

	// Insert expense
	query := `
		INSERT INTO expenses (description, amount, currency, category, date, vendor, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.DB.Exec(query,
		description,
		amount,
		"USD",
		category,
		expenseDate,
		vendor,
		notes,
	)
	if err != nil {
		return fmt.Errorf("failed to add expense: %w", err)
	}

	id, _ := result.LastInsertId()

	fmt.Printf("\n✓ Expense added successfully (ID: %d)\n", id)
	fmt.Printf("  Description: %s\n", description)
	fmt.Printf("  Amount: $%.2f\n", amount)
	fmt.Printf("  Category: %s\n", category)
	fmt.Printf("  Date: %s\n", expenseDate.Format("2006-01-02"))

	return nil
}

func runExpenseList(cmd *cobra.Command, args []string) error {
	query := `
		SELECT id, description, amount, currency, category, date, vendor
		FROM expenses
		ORDER BY date DESC
		LIMIT 50
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query expenses: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDATE\tDESCRIPTION\tCATEGORY\tVENDOR\tAMOUNT")

	total := 0.0

	for rows.Next() {
		var id uint
		var description, currency, category, vendor string
		var amount float64
		var date time.Time

		if err := rows.Scan(&id, &description, &amount, &currency, &category, &date, &vendor); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%.2f %s\n",
			id, date.Format("2006-01-02"), description, category, vendor, amount, currency)

		total += amount
	}

	w.Flush()
	fmt.Printf("\nTotal: $%.2f\n", total)

	return nil
}

func runExpenseReport(cmd *cobra.Command, args []string) error {
	fmt.Println("💸 Expense Report\n")

	// Total expenses by category
	categoryQuery := `
		SELECT category, COUNT(*) as count, SUM(amount) as total
		FROM expenses
		GROUP BY category
		ORDER BY total DESC
	`

	rows, err := db.DB.Query(categoryQuery)
	if err != nil {
		return fmt.Errorf("failed to query expenses by category: %w", err)
	}
	defer rows.Close()

	fmt.Println("By Category:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CATEGORY\tCOUNT\tTOTAL")

	grandTotal := 0.0

	for rows.Next() {
		var category string
		var count int
		var total float64

		if err := rows.Scan(&category, &count, &total); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Fprintf(w, "%s\t%d\t$%.2f\n", category, count, total)
		grandTotal += total
	}

	w.Flush()
	fmt.Printf("\nGrand Total: $%.2f\n\n", grandTotal)

	// Monthly expenses
	monthlyQuery := `
		SELECT
			strftime('%Y-%m', date) as month,
			COUNT(*) as count,
			SUM(amount) as total
		FROM expenses
		WHERE date >= date('now', '-6 months')
		GROUP BY strftime('%Y-%m', date)
		ORDER BY month DESC
	`

	rows2, err := db.DB.Query(monthlyQuery)
	if err != nil {
		return fmt.Errorf("failed to query monthly expenses: %w", err)
	}
	defer rows2.Close()

	fmt.Println("By Month (Last 6 months):")
	w2 := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w2, "MONTH\tCOUNT\tTOTAL")

	for rows2.Next() {
		var month string
		var count int
		var total float64

		if err := rows2.Scan(&month, &count, &total); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse month to make it prettier
		t, _ := time.Parse("2006-01", month)
		monthStr := t.Format("Jan 2006")

		fmt.Fprintf(w2, "%s\t%d\t$%.2f\n", monthStr, count, total)
	}

	w2.Flush()

	return nil
}
