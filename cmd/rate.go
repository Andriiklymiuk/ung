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

var rateCmd = &cobra.Command{
	Use:   "rate",
	Short: "Rate calculator and analysis",
	Long: `Calculate and analyze hourly rates.

Commands:
  calc      Calculate hourly rate from income target
  analyze   Analyze your actual rates from tracking data
  compare   Compare target rate vs actual rate`,
}

var rateCalcCmd = &cobra.Command{
	Use:   "calc",
	Short: "Calculate hourly rate from income target",
	RunE:  runRateCalc,
}

var rateAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze your actual rates",
	RunE:  runRateAnalyze,
}

var rateCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare target vs actual rate",
	RunE:  runRateCompare,
}

var (
	rateAnnual      float64
	rateMonthly     float64
	rateHoursWeek   float64
	rateWeeksYear   int
	rateExpenses    float64
	rateTaxPercent  float64
	rateProfitMargin float64
)

func init() {
	rateCalcCmd.Flags().Float64Var(&rateAnnual, "annual", 0, "Target annual income")
	rateCalcCmd.Flags().Float64Var(&rateMonthly, "monthly", 0, "Target monthly income")
	rateCalcCmd.Flags().Float64Var(&rateHoursWeek, "hours", 40, "Billable hours per week")
	rateCalcCmd.Flags().IntVar(&rateWeeksYear, "weeks", 48, "Working weeks per year (account for vacation)")
	rateCalcCmd.Flags().Float64Var(&rateExpenses, "expenses", 0, "Annual business expenses")
	rateCalcCmd.Flags().Float64Var(&rateTaxPercent, "tax", 25, "Estimated tax percentage")
	rateCalcCmd.Flags().Float64Var(&rateProfitMargin, "margin", 20, "Desired profit margin percentage")

	rootCmd.AddCommand(rateCmd)
	rateCmd.AddCommand(rateCalcCmd)
	rateCmd.AddCommand(rateAnalyzeCmd)
	rateCmd.AddCommand(rateCompareCmd)
}

func runRateCalc(cmd *cobra.Command, args []string) error {
	var targetAnnual float64

	if rateAnnual > 0 {
		targetAnnual = rateAnnual
	} else if rateMonthly > 0 {
		targetAnnual = rateMonthly * 12
	} else {
		return fmt.Errorf("specify either --annual or --monthly target income")
	}

	fmt.Println()
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│          RATE CALCULATOR                │\n")
	fmt.Printf("╰─────────────────────────────────────────╯\n")

	// Calculate gross needed (to cover taxes)
	taxMultiplier := 1 / (1 - rateTaxPercent/100)
	grossNeeded := targetAnnual * taxMultiplier

	// Add expenses
	totalNeeded := grossNeeded + rateExpenses

	// Add profit margin
	withMargin := totalNeeded * (1 + rateProfitMargin/100)

	// Calculate billable hours
	annualHours := rateHoursWeek * float64(rateWeeksYear)

	// Calculate hourly rate
	hourlyRate := withMargin / annualHours

	fmt.Println("\n📋 INPUTS")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Target net income:    $%.0f/year ($%.0f/month)\n", targetAnnual, targetAnnual/12)
	fmt.Printf("  Billable hours/week:  %.0f hours\n", rateHoursWeek)
	fmt.Printf("  Working weeks/year:   %d weeks\n", rateWeeksYear)
	fmt.Printf("  Business expenses:    $%.0f/year\n", rateExpenses)
	fmt.Printf("  Tax rate:             %.0f%%\n", rateTaxPercent)
	fmt.Printf("  Profit margin:        %.0f%%\n", rateProfitMargin)

	fmt.Println("\n💰 CALCULATION")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Net income target:    $%.0f\n", targetAnnual)
	fmt.Printf("  + Taxes (%.0f%%):        $%.0f\n", rateTaxPercent, grossNeeded-targetAnnual)
	fmt.Printf("  = Gross needed:       $%.0f\n", grossNeeded)
	fmt.Printf("  + Expenses:           $%.0f\n", rateExpenses)
	fmt.Printf("  + Profit margin:      $%.0f\n", withMargin-totalNeeded)
	fmt.Printf("  = Total revenue:      $%.0f\n", withMargin)
	fmt.Printf("  ÷ Annual hours:       %.0f hours\n", annualHours)

	fmt.Println("\n🎯 RECOMMENDED RATE")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Hourly rate:          $%.0f/hr\n", hourlyRate)
	fmt.Printf("  Daily rate (8hr):     $%.0f/day\n", hourlyRate*8)
	fmt.Printf("  Weekly rate:          $%.0f/week\n", hourlyRate*rateHoursWeek)

	// Show rate tiers
	fmt.Println("\n📊 RATE TIERS")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Budget:       $%.0f/hr (70%% of target)\n", hourlyRate*0.7)
	fmt.Printf("  Standard:     $%.0f/hr (target rate)\n", hourlyRate)
	fmt.Printf("  Premium:      $%.0f/hr (130%% of target)\n", hourlyRate*1.3)
	fmt.Printf("  Rush/Weekend: $%.0f/hr (150%% of target)\n", hourlyRate*1.5)

	fmt.Println()
	return nil
}

func runRateAnalyze(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│          RATE ANALYSIS                  │\n")
	fmt.Printf("╰─────────────────────────────────────────╯\n")

	// Get all tracking sessions with clients
	var sessions []models.TrackingSession
	db.GormDB.Where("billable = ? AND deleted_at IS NULL", true).
		Preload("Client").
		Preload("Contract").
		Find(&sessions)

	if len(sessions) == 0 {
		fmt.Println("\nNo billable tracking sessions found.")
		return nil
	}

	// Analyze by client
	clientStats := make(map[string]*clientRateStats)
	var totalHours, totalRevenue float64

	for _, s := range sessions {
		clientName := "Unknown"
		if s.Client != nil {
			clientName = s.Client.Name
		}

		if clientStats[clientName] == nil {
			clientStats[clientName] = &clientRateStats{}
		}

		hours := 0.0
		if s.Hours != nil {
			hours = *s.Hours
		}

		rate := 0.0
		if s.Contract != nil && s.Contract.HourlyRate != nil {
			rate = *s.Contract.HourlyRate
		}

		clientStats[clientName].hours += hours
		clientStats[clientName].revenue += hours * rate
		totalHours += hours
		totalRevenue += hours * rate
	}

	// Get paid invoices for the last year
	yearAgo := time.Now().AddDate(-1, 0, 0)
	var invoices []models.Invoice
	db.GormDB.Where("status = ? AND updated_at >= ?", models.StatusPaid, yearAgo).Find(&invoices)

	var invoiceRevenue float64
	for _, inv := range invoices {
		invoiceRevenue += inv.Amount
	}

	// Overall stats
	effectiveRate := 0.0
	if totalHours > 0 {
		effectiveRate = totalRevenue / totalHours
	}

	invoiceEffectiveRate := 0.0
	if totalHours > 0 {
		invoiceEffectiveRate = invoiceRevenue / totalHours
	}

	fmt.Println("\n📊 OVERALL STATS (All Time)")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  Total billable hours:    %.1f hours\n", totalHours)
	fmt.Printf("  Contract-based revenue:  $%.2f\n", totalRevenue)
	fmt.Printf("  Invoice revenue (1yr):   $%.2f\n", invoiceRevenue)
	fmt.Printf("  Effective rate (contracts): $%.0f/hr\n", effectiveRate)
	fmt.Printf("  Effective rate (invoices):  $%.0f/hr\n", invoiceEffectiveRate)

	// Per-client breakdown
	fmt.Println("\n👥 BY CLIENT")
	fmt.Println("───────────────────────────────────────")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CLIENT\tHOURS\tREVENUE\tEFF. RATE")

	for name, stats := range clientStats {
		effRate := 0.0
		if stats.hours > 0 {
			effRate = stats.revenue / stats.hours
		}
		fmt.Fprintf(w, "%s\t%.1f\t$%.0f\t$%.0f/hr\n",
			truncateStr(name, 20), stats.hours, stats.revenue, effRate)
	}
	w.Flush()

	// Rate distribution
	fmt.Println("\n📈 RATE RECOMMENDATIONS")
	fmt.Println("───────────────────────────────────────")
	if effectiveRate > 0 {
		fmt.Printf("  Your average effective rate is $%.0f/hr\n", effectiveRate)
		fmt.Printf("  Consider raising rates if clients accept at: $%.0f-$%.0f/hr\n",
			effectiveRate*1.1, effectiveRate*1.25)
	}

	fmt.Println()
	return nil
}

type clientRateStats struct {
	hours   float64
	revenue float64
}

func runRateCompare(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Printf("╭─────────────────────────────────────────╮\n")
	fmt.Printf("│         RATE COMPARISON                 │\n")
	fmt.Printf("╰─────────────────────────────────────────╯\n")

	// Get contracts and their rates
	var contracts []models.Contract
	db.GormDB.Where("active = ?", true).
		Preload("Client").Find(&contracts)

	if len(contracts) == 0 {
		fmt.Println("\nNo active contracts found.")
		return nil
	}

	fmt.Println("\n📋 CONTRACT RATES")
	fmt.Println("───────────────────────────────────────")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CONTRACT\tCLIENT\tTYPE\tRATE")

	var hourlyContracts []models.Contract
	var rateSum float64

	for _, c := range contracts {
		rateStr := "-"
		if c.ContractType == models.ContractTypeHourly && c.HourlyRate != nil {
			rateStr = fmt.Sprintf("$%.0f/hr", *c.HourlyRate)
			hourlyContracts = append(hourlyContracts, c)
			rateSum += *c.HourlyRate
		} else if c.FixedPrice != nil {
			rateStr = fmt.Sprintf("$%.0f (fixed)", *c.FixedPrice)
		}

		clientName := "Unknown"
		if c.Client.Name != "" {
			clientName = c.Client.Name
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			truncateStr(c.Name, 20), truncateStr(clientName, 15), c.ContractType, rateStr)
	}
	w.Flush()

	if len(hourlyContracts) > 0 {
		avgRate := rateSum / float64(len(hourlyContracts))
		minRate := hourlyContracts[0].HourlyRate
		maxRate := hourlyContracts[0].HourlyRate

		for _, c := range hourlyContracts {
			if c.HourlyRate != nil {
				if *c.HourlyRate < *minRate {
					minRate = c.HourlyRate
				}
				if *c.HourlyRate > *maxRate {
					maxRate = c.HourlyRate
				}
			}
		}

		fmt.Println("\n📊 HOURLY RATE SUMMARY")
		fmt.Println("───────────────────────────────────────")
		fmt.Printf("  Contracts:    %d hourly contracts\n", len(hourlyContracts))
		fmt.Printf("  Min rate:     $%.0f/hr\n", *minRate)
		fmt.Printf("  Max rate:     $%.0f/hr\n", *maxRate)
		fmt.Printf("  Average:      $%.0f/hr\n", avgRate)
		fmt.Printf("  Range:        $%.0f spread\n", *maxRate-*minRate)
	}

	// Compare with actual earned
	var sessions []models.TrackingSession
	thirtyDays := time.Now().AddDate(0, 0, -30)
	db.GormDB.Where("billable = ? AND start_time >= ? AND deleted_at IS NULL", true, thirtyDays).
		Find(&sessions)

	var actualHours float64
	for _, s := range sessions {
		if s.Hours != nil {
			actualHours += *s.Hours
		}
	}

	var paidAmount float64
	var paidInvoices []models.Invoice
	db.GormDB.Where("status = ? AND updated_at >= ?", models.StatusPaid, thirtyDays).Find(&paidInvoices)
	for _, inv := range paidInvoices {
		paidAmount += inv.Amount
	}

	if actualHours > 0 {
		actualRate := paidAmount / actualHours

		fmt.Println("\n🎯 ACTUAL VS TARGET (Last 30 Days)")
		fmt.Println("───────────────────────────────────────")
		fmt.Printf("  Billable hours:     %.1f hours\n", actualHours)
		fmt.Printf("  Revenue earned:     $%.2f\n", paidAmount)
		fmt.Printf("  Actual eff. rate:   $%.0f/hr\n", actualRate)

		if len(hourlyContracts) > 0 {
			avgRate := rateSum / float64(len(hourlyContracts))
			diff := actualRate - avgRate
			diffPercent := (diff / avgRate) * 100

			if diff >= 0 {
				fmt.Printf("  vs Avg contract:    +$%.0f/hr (+%.0f%%) ✓\n", diff, diffPercent)
			} else {
				fmt.Printf("  vs Avg contract:    -$%.0f/hr (%.0f%%)\n", -diff, diffPercent)
				fmt.Println("  Consider: reviewing non-billable time or raising rates")
			}
		}
	}

	fmt.Println()
	return nil
}
