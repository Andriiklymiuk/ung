package cmd

import (
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
	fmt.Println("💰 Revenue Summary")
	fmt.Println()

	// Total revenue (paid invoices)
	var paidInvoices []models.Invoice
	db.GormDB.Where("status = ?", models.StatusPaid).Find(&paidInvoices)
	var totalPaid float64
	for _, inv := range paidInvoices {
		totalPaid += inv.Amount
	}

	// Total pending
	var pendingInvoices []models.Invoice
	db.GormDB.Where("status IN ?", []models.InvoiceStatus{models.StatusPending, models.StatusSent}).Find(&pendingInvoices)
	var totalPending float64
	for _, inv := range pendingInvoices {
		totalPending += inv.Amount
	}

	// Total overdue
	var overdueInvoices []models.Invoice
	db.GormDB.Where("status = ? OR (status IN ? AND due_date < ?)",
		models.StatusOverdue,
		[]models.InvoiceStatus{models.StatusPending, models.StatusSent},
		time.Now()).Find(&overdueInvoices)
	var totalOverdue float64
	for _, inv := range overdueInvoices {
		totalOverdue += inv.Amount
	}

	// Revenue by month (last 6 months)
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)
	var monthlyInvoices []models.Invoice
	db.GormDB.Where("status = ? AND issued_date >= ?", models.StatusPaid, sixMonthsAgo).
		Order("issued_date DESC").Find(&monthlyInvoices)

	// Group by month
	monthlyData := make(map[string]struct {
		count int
		total float64
	})
	for _, inv := range monthlyInvoices {
		monthKey := inv.IssuedDate.Format("2006-01")
		data := monthlyData[monthKey]
		data.count++
		data.total += inv.Amount
		monthlyData[monthKey] = data
	}

	fmt.Printf("Overall:\n")
	fmt.Printf("  Paid:    $%.2f\n", totalPaid)
	fmt.Printf("  Pending: $%.2f\n", totalPending)
	fmt.Printf("  Overdue: $%.2f\n\n", totalOverdue)

	fmt.Println("Monthly Revenue (Paid Invoices):")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MONTH\tCOUNT\tTOTAL")

	// Sort months in descending order
	for monthKey, data := range monthlyData {
		t, _ := time.Parse("2006-01", monthKey)
		monthStr := t.Format("Jan 2006")
		fmt.Fprintf(w, "%s\t%d\t$%.2f\n", monthStr, data.count, data.total)
	}

	w.Flush()
	return nil
}

func runReportClients(cmd *cobra.Command, args []string) error {
	fmt.Println("👥 Client Summary")
	fmt.Println()

	var clients []models.Client
	db.GormDB.Find(&clients)

	// Get all invoices
	var invoices []models.Invoice
	db.GormDB.Find(&invoices)

	// Get all invoice recipients
	var recipients []models.InvoiceRecipient
	db.GormDB.Find(&recipients)

	// Build invoice to client mapping
	invoiceClients := make(map[uint][]uint) // invoiceID -> []clientID
	for _, r := range recipients {
		invoiceClients[r.InvoiceID] = append(invoiceClients[r.InvoiceID], r.ClientID)
	}

	// Build client stats
	type clientStats struct {
		invoiceCount int
		paid         float64
		pending      float64
		overdue      float64
	}
	stats := make(map[uint]*clientStats)

	for _, client := range clients {
		stats[client.ID] = &clientStats{}
	}

	now := time.Now()
	for _, inv := range invoices {
		clientIDs := invoiceClients[inv.ID]
		for _, clientID := range clientIDs {
			if s, ok := stats[clientID]; ok {
				s.invoiceCount++
				switch inv.Status {
				case models.StatusPaid:
					s.paid += inv.Amount
				case models.StatusPending, models.StatusSent:
					if inv.DueDate.Before(now) {
						s.overdue += inv.Amount
					} else {
						s.pending += inv.Amount
					}
				case models.StatusOverdue:
					s.overdue += inv.Amount
				}
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCLIENT\tINVOICES\tPAID\tPENDING\tOVERDUE")

	for _, client := range clients {
		s := stats[client.ID]
		fmt.Fprintf(w, "%d\t%s\t%d\t$%.2f\t$%.2f\t$%.2f\n",
			client.ID, client.Name, s.invoiceCount, s.paid, s.pending, s.overdue)
	}

	w.Flush()
	return nil
}

func runReportOverdue(cmd *cobra.Command, args []string) error {
	fmt.Println("⚠️  Overdue Invoices")
	fmt.Println()

	now := time.Now()

	// First, auto-mark overdue invoices
	db.GormDB.Model(&models.Invoice{}).
		Where("status IN ? AND due_date < ?",
			[]models.InvoiceStatus{models.StatusPending, models.StatusSent}, now).
		Update("status", models.StatusOverdue)

	// Get overdue invoices
	var invoices []models.Invoice
	db.GormDB.Where("status = ?", models.StatusOverdue).Order("due_date ASC").Find(&invoices)

	// Get invoice recipients to find client names
	var recipients []models.InvoiceRecipient
	db.GormDB.Find(&recipients)

	var clients []models.Client
	db.GormDB.Find(&clients)

	clientMap := make(map[uint]string)
	for _, c := range clients {
		clientMap[c.ID] = c.Name
	}

	invoiceClient := make(map[uint]string)
	for _, r := range recipients {
		if name, ok := clientMap[r.ClientID]; ok {
			invoiceClient[r.InvoiceID] = name
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tINVOICE#\tCLIENT\tAMOUNT\tDUE DATE\tDAYS OVERDUE")

	totalOverdue := 0.0
	count := 0

	for _, inv := range invoices {
		clientName := invoiceClient[inv.ID]
		if clientName == "" {
			clientName = "Unknown"
		}
		daysOverdue := int(now.Sub(inv.DueDate).Hours() / 24)

		fmt.Fprintf(w, "%d\t%s\t%s\t%.2f %s\t%s\t%d days\n",
			inv.ID, inv.InvoiceNum, clientName, inv.Amount, inv.Currency,
			inv.DueDate.Format("2006-01-02"), daysOverdue)

		totalOverdue += inv.Amount
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
	fmt.Println("📄 Unpaid Invoices")
	fmt.Println()

	// Get unpaid invoices
	var invoices []models.Invoice
	db.GormDB.Where("status != ?", models.StatusPaid).Order("due_date ASC").Find(&invoices)

	// Get invoice recipients to find client names
	var recipients []models.InvoiceRecipient
	db.GormDB.Find(&recipients)

	var clients []models.Client
	db.GormDB.Find(&clients)

	clientMap := make(map[uint]string)
	for _, c := range clients {
		clientMap[c.ID] = c.Name
	}

	invoiceClient := make(map[uint]string)
	for _, r := range recipients {
		if name, ok := clientMap[r.ClientID]; ok {
			invoiceClient[r.InvoiceID] = name
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tINVOICE#\tCLIENT\tAMOUNT\tSTATUS\tISSUED\tDUE")

	totalUnpaid := 0.0
	count := 0

	for _, inv := range invoices {
		clientName := invoiceClient[inv.ID]
		if clientName == "" {
			clientName = "Unknown"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%.2f %s\t%s\t%s\t%s\n",
			inv.ID, inv.InvoiceNum, clientName, inv.Amount, inv.Currency, inv.Status,
			inv.IssuedDate.Format("2006-01-02"), inv.DueDate.Format("2006-01-02"))

		totalUnpaid += inv.Amount
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
	db.GormDB.Where("start_time >= ? AND start_time < ? AND deleted_at IS NULL", startOfWeek, endOfWeek).
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
	db.GormDB.Where("issued_date >= ? AND issued_date < ?", startOfWeek, endOfWeek).Find(&invoices)

	var invoicedAmount float64
	for _, inv := range invoices {
		invoicedAmount += inv.Amount
	}

	// Get payments received this week
	var paidInvoices []models.Invoice
	db.GormDB.Where("status = ? AND updated_at >= ? AND updated_at < ?", "paid", startOfWeek, endOfWeek).Find(&paidInvoices)

	var paidAmount float64
	for _, inv := range paidInvoices {
		paidAmount += inv.Amount
	}

	// Get expenses
	var expenses []models.Expense
	db.GormDB.Where("date >= ? AND date < ?", startOfWeek, endOfWeek).Find(&expenses)

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
	db.GormDB.Where("start_time >= ? AND start_time < ? AND deleted_at IS NULL", startOfMonth, endOfMonth).
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
	db.GormDB.Where("issued_date >= ? AND issued_date < ?", startOfMonth, endOfMonth).Find(&invoices)

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
	db.GormDB.Where("date >= ? AND date < ?", startOfMonth, endOfMonth).Find(&expenses)

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
