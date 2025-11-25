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

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate financial reports",
	Long: `Generate various financial and business reports.

Commands:
  weekly     Weekly summary of hours, earnings, and activity
  monthly    Monthly summary
  revenue    Revenue breakdown
  clients    Client summary
  overdue    Overdue invoices
  unpaid     All unpaid invoices`,
}

var reportWeeklyCmd = &cobra.Command{
	Use:   "weekly",
	Short: "Generate weekly summary report",
	RunE:  runReportWeekly,
}

var reportMonthlyCmd = &cobra.Command{
	Use:   "monthly",
	Short: "Generate monthly summary report",
	RunE:  runReportMonthly,
}

var reportRevenueCmd = &cobra.Command{
	Use:   "revenue",
	Short: "Show revenue summary",
	RunE:  runReportRevenue,
}

var reportClientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Show client summary",
	RunE:  runReportClients,
}

var reportOverdueCmd = &cobra.Command{
	Use:   "overdue",
	Short: "Show overdue invoices",
	RunE:  runReportOverdue,
}

var reportUnpaidCmd = &cobra.Command{
	Use:   "unpaid",
	Short: "Show all unpaid invoices",
	RunE:  runReportUnpaid,
}

var (
	reportLast  bool
	reportMonth int
	reportYear  int
)

func init() {
	reportWeeklyCmd.Flags().BoolVar(&reportLast, "last", false, "Show last week's report")
	reportMonthlyCmd.Flags().IntVar(&reportMonth, "month", 0, "Specific month (1-12)")
	reportMonthlyCmd.Flags().IntVar(&reportYear, "year", 0, "Specific year")

	rootCmd.AddCommand(reportCmd)
	reportCmd.AddCommand(reportWeeklyCmd)
	reportCmd.AddCommand(reportMonthlyCmd)
	reportCmd.AddCommand(reportRevenueCmd)
	reportCmd.AddCommand(reportClientsCmd)
	reportCmd.AddCommand(reportOverdueCmd)
	reportCmd.AddCommand(reportUnpaidCmd)
}

func runReportRevenue(cmd *cobra.Command, args []string) error {
	fmt.Println("💰 Revenue Summary\n")

	// Total revenue (paid invoices)
	var totalPaid sql.NullFloat64
	err := db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM invoices
		WHERE status = ?
	`, models.StatusPaid).Scan(&totalPaid)
	if err != nil {
		return fmt.Errorf("failed to get paid total: %w", err)
	}

	// Total pending
	var totalPending sql.NullFloat64
	err = db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM invoices
		WHERE status IN (?, ?)
	`, models.StatusPending, models.StatusSent).Scan(&totalPending)
	if err != nil {
		return fmt.Errorf("failed to get pending total: %w", err)
	}

	// Total overdue
	var totalOverdue sql.NullFloat64
	err = db.DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM invoices
		WHERE status = ? OR (status IN (?, ?) AND due_date < ?)
	`, models.StatusOverdue, models.StatusPending, models.StatusSent, time.Now()).Scan(&totalOverdue)
	if err != nil {
		return fmt.Errorf("failed to get overdue total: %w", err)
	}

	// Revenue by month (last 6 months)
	monthlyQuery := `
		SELECT
			strftime('%Y-%m', issued_date) as month,
			COUNT(*) as count,
			SUM(amount) as total
		FROM invoices
		WHERE status = ?
		  AND issued_date >= date('now', '-6 months')
		GROUP BY strftime('%Y-%m', issued_date)
		ORDER BY month DESC
	`

	rows, err := db.DB.Query(monthlyQuery, models.StatusPaid)
	if err != nil {
		return fmt.Errorf("failed to get monthly revenue: %w", err)
	}
	defer rows.Close()

	fmt.Printf("Overall:\n")
	fmt.Printf("  Paid:    $%.2f\n", totalPaid.Float64)
	fmt.Printf("  Pending: $%.2f\n", totalPending.Float64)
	fmt.Printf("  Overdue: $%.2f\n\n", totalOverdue.Float64)

	fmt.Println("Monthly Revenue (Paid Invoices):")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MONTH\tCOUNT\tTOTAL")

	for rows.Next() {
		var month string
		var count int
		var total float64
		if err := rows.Scan(&month, &count, &total); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse month to make it prettier
		t, _ := time.Parse("2006-01", month)
		monthStr := t.Format("Jan 2006")

		fmt.Fprintf(w, "%s\t%d\t$%.2f\n", monthStr, count, total)
	}

	w.Flush()
	return nil
}

func runReportClients(cmd *cobra.Command, args []string) error {
	fmt.Println("👥 Client Summary\n")

	query := `
		SELECT
			c.id,
			c.name,
			COUNT(DISTINCT i.id) as invoice_count,
			COALESCE(SUM(CASE WHEN i.status = ? THEN i.amount ELSE 0 END), 0) as paid,
			COALESCE(SUM(CASE WHEN i.status IN (?, ?) THEN i.amount ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN i.status = ? OR (i.status IN (?, ?) AND i.due_date < ?) THEN i.amount ELSE 0 END), 0) as overdue
		FROM clients c
		LEFT JOIN invoice_recipients ir ON c.id = ir.client_id
		LEFT JOIN invoices i ON ir.invoice_id = i.id
		GROUP BY c.id, c.name
		ORDER BY paid DESC
	`

	rows, err := db.DB.Query(query,
		models.StatusPaid,
		models.StatusPending, models.StatusSent,
		models.StatusOverdue, models.StatusPending, models.StatusSent, time.Now())
	if err != nil {
		return fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCLIENT\tINVOICES\tPAID\tPENDING\tOVERDUE")

	for rows.Next() {
		var id uint
		var name string
		var invoiceCount int
		var paid, pending, overdue float64

		if err := rows.Scan(&id, &name, &invoiceCount, &paid, &pending, &overdue); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Fprintf(w, "%d\t%s\t%d\t$%.2f\t$%.2f\t$%.2f\n",
			id, name, invoiceCount, paid, pending, overdue)
	}

	w.Flush()
	return nil
}

func runReportOverdue(cmd *cobra.Command, args []string) error {
	fmt.Println("⚠️  Overdue Invoices\n")

	// First, auto-mark overdue invoices
	_, err := db.DB.Exec(`
		UPDATE invoices
		SET status = ?
		WHERE status IN (?, ?)
		  AND due_date < ?
	`, models.StatusOverdue, models.StatusPending, models.StatusSent, time.Now())
	if err != nil {
		fmt.Printf("Warning: Could not auto-mark overdue invoices: %v\n", err)
	}

	query := `
		SELECT
			i.id,
			i.invoice_num,
			c.name as client_name,
			i.amount,
			i.currency,
			i.due_date,
			CAST((julianday('now') - julianday(i.due_date)) AS INTEGER) as days_overdue
		FROM invoices i
		LEFT JOIN invoice_recipients ir ON i.id = ir.invoice_id
		LEFT JOIN clients c ON ir.client_id = c.id
		WHERE i.status = ?
		ORDER BY i.due_date ASC
	`

	rows, err := db.DB.Query(query, models.StatusOverdue)
	if err != nil {
		return fmt.Errorf("failed to query overdue invoices: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tINVOICE#\tCLIENT\tAMOUNT\tDUE DATE\tDAYS OVERDUE")

	totalOverdue := 0.0
	count := 0

	for rows.Next() {
		var id uint
		var invoiceNum, clientName, currency string
		var amount float64
		var dueDate time.Time
		var daysOverdue int

		if err := rows.Scan(&id, &invoiceNum, &clientName, &amount, &currency, &dueDate, &daysOverdue); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%.2f %s\t%s\t%d days\n",
			id, invoiceNum, clientName, amount, currency, dueDate.Format("2006-01-02"), daysOverdue)

		totalOverdue += amount
		count++
	}

	w.Flush()

	if count == 0 {
		fmt.Println("No overdue invoices. Great job! 🎉")
	} else {
		fmt.Printf("\nTotal: %d invoices, $%.2f overdue\n", count, totalOverdue)
	}

	return nil
}

func runReportUnpaid(cmd *cobra.Command, args []string) error {
	fmt.Println("📄 Unpaid Invoices\n")

	query := `
		SELECT
			i.id,
			i.invoice_num,
			c.name as client_name,
			i.amount,
			i.currency,
			i.status,
			i.issued_date,
			i.due_date
		FROM invoices i
		LEFT JOIN invoice_recipients ir ON i.id = ir.invoice_id
		LEFT JOIN clients c ON ir.client_id = c.id
		WHERE i.status != ?
		ORDER BY i.due_date ASC
	`

	rows, err := db.DB.Query(query, models.StatusPaid)
	if err != nil {
		return fmt.Errorf("failed to query unpaid invoices: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tINVOICE#\tCLIENT\tAMOUNT\tSTATUS\tISSUED\tDUE")

	totalUnpaid := 0.0
	count := 0

	for rows.Next() {
		var id uint
		var invoiceNum, clientName, currency, status string
		var amount float64
		var issuedDate, dueDate time.Time

		if err := rows.Scan(&id, &invoiceNum, &clientName, &amount, &currency, &status, &issuedDate, &dueDate); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%.2f %s\t%s\t%s\t%s\n",
			id, invoiceNum, clientName, amount, currency, status,
			issuedDate.Format("2006-01-02"), dueDate.Format("2006-01-02"))

		totalUnpaid += amount
		count++
	}

	w.Flush()

	if count == 0 {
		fmt.Println("All invoices are paid! 🎉")
	} else {
		fmt.Printf("\nTotal: %d unpaid invoices, $%.2f\n", count, totalUnpaid)
	}

	return nil
}

func runReportWeekly(cmd *cobra.Command, args []string) error {
	now := time.Now()

	// Calculate week boundaries (Monday-Sunday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	startOfWeek := now.AddDate(0, 0, -weekday+1)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())

	if reportLast {
		startOfWeek = startOfWeek.AddDate(0, 0, -7)
	}

	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	// Header
	fmt.Println()
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│            WEEKLY REPORT                │\n")
	fmt.Printf("│      %s - %s        │\n",
		startOfWeek.Format("Jan 02"),
		endOfWeek.AddDate(0, 0, -1).Format("Jan 02, 2006"))
	fmt.Printf("╰─────────────────────────────────────────╯\n")

	// Get time tracking data
	var sessions []models.TrackingSession
	db.DB.Where("start_time >= ? AND start_time < ? AND deleted_at IS NULL", startOfWeek, endOfWeek).
		Preload("Client").Find(&sessions)

	totalHours := 0.0
	billableHours := 0.0
	clientHours := make(map[string]float64)
	projectHours := make(map[string]float64)

	for _, s := range sessions {
		hours := 0.0
		if s.Hours != nil {
			hours = *s.Hours
		} else if s.Duration != nil {
			hours = float64(*s.Duration) / 3600
		}
		totalHours += hours

		if s.Billable {
			billableHours += hours
		}

		clientName := "No Client"
		if s.Client != nil {
			clientName = s.Client.Name
		}
		clientHours[clientName] += hours

		if s.ProjectName != "" {
			projectHours[s.ProjectName] += hours
		}
	}

	// Get invoices sent this week
	var invoices []models.Invoice
	db.DB.Where("issued_date >= ? AND issued_date < ?", startOfWeek, endOfWeek).Find(&invoices)

	var invoicedAmount float64
	for _, inv := range invoices {
		invoicedAmount += inv.Amount
	}

	// Get payments received this week
	var paidInvoices []models.Invoice
	db.DB.Where("status = ? AND updated_at >= ? AND updated_at < ?", "paid", startOfWeek, endOfWeek).Find(&paidInvoices)

	var paidAmount float64
	for _, inv := range paidInvoices {
		paidAmount += inv.Amount
	}

	// Get expenses
	var expenses []models.Expense
	db.DB.Where("date >= ? AND date < ?", startOfWeek, endOfWeek).Find(&expenses)

	var expenseAmount float64
	for _, e := range expenses {
		expenseAmount += e.Amount
	}

	// Print summary
	fmt.Println("\n📊 SUMMARY")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Hours worked:      %.1f hours\n", totalHours)
	fmt.Printf("  Billable hours:    %.1f hours (%.0f%%)\n", billableHours, safePercent(billableHours, totalHours))
	fmt.Printf("  Sessions:          %d\n", len(sessions))
	fmt.Println()
	fmt.Printf("  Invoices sent:     %d\n", len(invoices))
	fmt.Printf("  Amount invoiced:   $%.2f\n", invoicedAmount)
	fmt.Printf("  Payments received: $%.2f\n", paidAmount)
	fmt.Printf("  Expenses:          $%.2f\n", expenseAmount)

	if billableHours > 0 && invoicedAmount > 0 {
		effectiveRate := invoicedAmount / billableHours
		fmt.Printf("  Effective rate:    $%.0f/hr\n", effectiveRate)
	}

	// Client breakdown
	if len(clientHours) > 0 {
		fmt.Println("\n👥 BY CLIENT")
		fmt.Println("───────────────────────────────────────")
		for client, hours := range clientHours {
			bar := progressBar(hours, totalHours, 20)
			fmt.Printf("  %-15s %s %.1fh\n", truncateStr(client, 15), bar, hours)
		}
	}

	// Project breakdown
	if len(projectHours) > 0 {
		fmt.Println("\n📁 BY PROJECT")
		fmt.Println("───────────────────────────────────────")
		for project, hours := range projectHours {
			bar := progressBar(hours, totalHours, 20)
			fmt.Printf("  %-15s %s %.1fh\n", truncateStr(project, 15), bar, hours)
		}
	}

	// Daily breakdown
	fmt.Println("\n📅 DAILY BREAKDOWN")
	fmt.Println("───────────────────────────────────────")
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	dailyHours := make([]float64, 7)

	for _, s := range sessions {
		dayIdx := int(s.StartTime.Weekday())
		if dayIdx == 0 {
			dayIdx = 6
		} else {
			dayIdx--
		}

		hours := 0.0
		if s.Hours != nil {
			hours = *s.Hours
		}
		dailyHours[dayIdx] += hours
	}

	maxDaily := 0.0
	for _, h := range dailyHours {
		if h > maxDaily {
			maxDaily = h
		}
	}

	for i, day := range days {
		bar := progressBar(dailyHours[i], maxDaily, 15)
		fmt.Printf("  %s  %s %.1fh\n", day, bar, dailyHours[i])
	}

	fmt.Println()
	return nil
}

func runReportMonthly(cmd *cobra.Command, args []string) error {
	now := time.Now()

	year := now.Year()
	month := int(now.Month())

	if reportYear > 0 {
		year = reportYear
	}
	if reportMonth > 0 && reportMonth <= 12 {
		month = reportMonth
	}

	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	fmt.Println()
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│           MONTHLY REPORT                │\n")
	fmt.Printf("│              %s %d                   │\n", startOfMonth.Format("January")[:3], year)
	fmt.Printf("╰─────────────────────────────────────────╯\n")

	// Get time tracking data
	var sessions []models.TrackingSession
	db.DB.Where("start_time >= ? AND start_time < ? AND deleted_at IS NULL", startOfMonth, endOfMonth).
		Preload("Client").Find(&sessions)

	totalHours := 0.0
	billableHours := 0.0
	for _, s := range sessions {
		if s.Hours != nil {
			totalHours += *s.Hours
			if s.Billable {
				billableHours += *s.Hours
			}
		}
	}

	// Get invoices
	var invoices []models.Invoice
	db.DB.Where("issued_date >= ? AND issued_date < ?", startOfMonth, endOfMonth).Find(&invoices)

	var revenue float64
	var pending float64
	for _, inv := range invoices {
		revenue += inv.Amount
		if inv.Status != models.StatusPaid {
			pending += inv.Amount
		}
	}

	// Get expenses
	var expenses []models.Expense
	db.DB.Where("date >= ? AND date < ?", startOfMonth, endOfMonth).Find(&expenses)

	var expenseTotal float64
	expenseByCategory := make(map[string]float64)
	for _, e := range expenses {
		expenseTotal += e.Amount
		expenseByCategory[string(e.Category)] += e.Amount
	}

	profit := revenue - expenseTotal

	fmt.Println("\n💰 FINANCIAL SUMMARY")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Revenue:           $%.2f\n", revenue)
	fmt.Printf("  Expenses:          $%.2f\n", expenseTotal)
	fmt.Printf("  ─────────────────────────\n")
	fmt.Printf("  Profit:            $%.2f\n", profit)
	fmt.Printf("  Pending payment:   $%.2f\n", pending)

	fmt.Println("\n⏱️  TIME SUMMARY")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Total hours:       %.1f\n", totalHours)
	fmt.Printf("  Billable hours:    %.1f (%.0f%%)\n", billableHours, safePercent(billableHours, totalHours))
	fmt.Printf("  Invoices sent:     %d\n", len(invoices))

	if billableHours > 0 {
		fmt.Printf("  Avg hourly rate:   $%.0f/hr\n", revenue/billableHours)
	}

	// Expense breakdown
	if len(expenseByCategory) > 0 {
		fmt.Println("\n💸 EXPENSES BY CATEGORY")
		fmt.Println("───────────────────────────────────────")
		for cat, amount := range expenseByCategory {
			bar := progressBar(amount, expenseTotal, 15)
			fmt.Printf("  %-12s %s $%.0f\n", truncateStr(cat, 12), bar, amount)
		}
	}

	fmt.Println()
	return nil
}

// Helper functions
func progressBar(value, max float64, width int) string {
	if max == 0 {
		return "│" + repeatStr("░", width) + "│"
	}
	filled := int((value / max) * float64(width))
	if filled > width {
		filled = width
	}
	return "│" + repeatStr("█", filled) + repeatStr("░", width-filled) + "│"
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s + repeatStr(" ", maxLen-len(s))
	}
	return s[:maxLen-2] + ".."
}

func safePercent(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (part / total) * 100
}
