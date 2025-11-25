package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/internal/repository"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show revenue dashboard and projections",
	Long:  "Display monthly revenue projections, contract breakdown, and key metrics",
	RunE:  runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard(cmd *cobra.Command, args []string) error {
	// Initialize repositories
	contractRepo := repository.NewContractRepository()
	trackingRepo := repository.NewTrackingSessionRepository()
	invoiceRepo := repository.NewInvoiceRepository()
	clientRepo := repository.NewClientRepository()

	// Fetch active contracts
	contracts, err := contractRepo.ListActive()
	if err != nil {
		return fmt.Errorf("failed to fetch contracts: %w", err)
	}

	fmt.Println("📊 UNG Revenue Dashboard")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println()

	// Calculate projections
	var totalMonthly float64
	var hourlyRevenue float64
	var retainerRevenue float64
	var projectedHours float64

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// Contract breakdown
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CONTRACT\tCLIENT\tTYPE\tMONTHLY REVENUE\tDETAILS")
	fmt.Fprintln(w, "--------\t------\t----\t---------------\t-------")

	for _, contract := range contracts {
		var monthly float64
		var details string

		switch contract.ContractType {
		case models.ContractTypeHourly:
			if contract.HourlyRate != nil {
				// Get last 30 days hours
				sessions, err := trackingRepo.GetByContractID(contract.ID)
				if err != nil {
					continue
				}

				var hours float64
				for _, s := range sessions {
					if s.StartTime.After(thirtyDaysAgo) && s.Billable && s.Hours != nil {
						hours += *s.Hours
					}
				}

				monthly = hours * (*contract.HourlyRate)
				details = fmt.Sprintf("%.1fh @ $%.0f/hr", hours, *contract.HourlyRate)
				hourlyRevenue += monthly
				projectedHours += hours
			}

		case models.ContractTypeRetainer:
			if contract.FixedPrice != nil {
				monthly = *contract.FixedPrice
				details = "Monthly retainer"
				retainerRevenue += monthly
			}

		case models.ContractTypeFixedPrice:
			if contract.FixedPrice != nil {
				// Prorate over contract duration
				if contract.EndDate != nil {
					duration := contract.EndDate.Sub(contract.StartDate)
					months := duration.Hours() / 24 / 30
					if months > 0 {
						monthly = *contract.FixedPrice / months
					} else {
						monthly = *contract.FixedPrice / 3 // Default 3 months
					}
				} else {
					monthly = *contract.FixedPrice / 3
				}
				details = fmt.Sprintf("$%.0f total", *contract.FixedPrice)
			}
		}

		totalMonthly += monthly
		fmt.Fprintf(w, "%s\t%s\t%s\t$%.2f\t%s\n",
			contract.Name, contract.Client.Name, contract.ContractType, monthly, details)
	}

	fmt.Fprintln(w, "--------\t------\t----\t---------------\t-------")
	fmt.Fprintf(w, "TOTAL\t\t\t$%.2f\t%d contracts\n", totalMonthly, len(contracts))
	w.Flush()

	fmt.Println()
	fmt.Println("💰 Revenue Breakdown:")
	fmt.Printf("  Hourly Contracts:  $%.2f\n", hourlyRevenue)
	fmt.Printf("  Retainer Contracts: $%.2f\n", retainerRevenue)
	fmt.Printf("  Fixed/Project:     $%.2f\n", totalMonthly-hourlyRevenue-retainerRevenue)
	fmt.Printf("  Projected Hours:   %.1f hours\n", projectedHours)

	if projectedHours > 0 {
		avgRate := hourlyRevenue / projectedHours
		fmt.Printf("  Average Rate:      $%.0f/hr\n", avgRate)
	}

	// Quick stats
	fmt.Println()
	fmt.Println("📈 Quick Stats:")

	// Count total clients
	allClients, err := clientRepo.List()
	totalClients := len(allClients)
	fmt.Printf("  Total Clients:     %d\n", totalClients)

	// Count pending invoices
	pendingInvoices, err := invoiceRepo.GetByStatus(models.StatusPending)
	fmt.Printf("  Pending Invoices:  %d\n", len(pendingInvoices))

	// Calculate unpaid amount
	var unpaidAmount float64
	unpaidStatuses := []models.InvoiceStatus{models.StatusPending, models.StatusOverdue}
	for _, status := range unpaidStatuses {
		invoices, err := invoiceRepo.GetByStatus(status)
		if err == nil {
			for _, inv := range invoices {
				unpaidAmount += inv.Amount
			}
		}
	}
	fmt.Printf("  Unpaid Amount:     $%.2f\n", unpaidAmount)

	return nil
}
