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
	Long:  "Generate various financial and business reports",
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

func init() {
	rootCmd.AddCommand(reportCmd)
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
