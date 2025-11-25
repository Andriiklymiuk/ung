package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Backup and sync your data",
	Long: `Backup and synchronize your UNG data.

Commands:
  backup     Create a backup of all data
  restore    Restore data from a backup
  ls         List available backups

Examples:
  ung sync backup                Create backup
  ung sync backup --output ~/    Save backup to custom location
  ung sync restore               Restore from latest backup
  ung sync ls                    List backups`,
	RunE: runSyncInteractive,
}

var syncBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of all data",
	RunE:  runSyncBackup,
}

var syncRestoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore data from a backup",
	RunE:  runSyncRestore,
}

var syncListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List available backups",
	RunE:    runSyncList,
}

var (
	syncOutputPath string
	syncForce      bool
)

func init() {
	syncBackupCmd.Flags().StringVarP(&syncOutputPath, "output", "o", "", "Output directory for backup")
	syncRestoreCmd.Flags().BoolVarP(&syncForce, "force", "f", false, "Force restore without confirmation")

	syncCmd.AddCommand(syncBackupCmd)
	syncCmd.AddCommand(syncRestoreCmd)
	syncCmd.AddCommand(syncListCmd)

	rootCmd.AddCommand(syncCmd)
}

// BackupData represents all data in the system
type BackupData struct {
	Version           string                    `json:"version"`
	CreatedAt         time.Time                 `json:"created_at"`
	Companies         []models.Company          `json:"companies"`
	Clients           []models.Client           `json:"clients"`
	Contracts         []models.Contract         `json:"contracts"`
	Invoices          []models.Invoice          `json:"invoices"`
	InvoiceRecipients []models.InvoiceRecipient `json:"invoice_recipients"`
	InvoiceLineItems  []models.InvoiceLineItem  `json:"invoice_line_items"`
	TrackingSessions  []models.TrackingSession  `json:"tracking_sessions"`
	Expenses          []models.Expense          `json:"expenses"`
	RecurringInvoices []models.RecurringInvoice `json:"recurring_invoices"`
}

func runSyncInteractive(cmd *cobra.Command, args []string) error {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Sync Action").
				Options(
					huh.NewOption("Create Backup", "backup"),
					huh.NewOption("Restore from Backup", "restore"),
					huh.NewOption("List Backups", "list"),
				).
				Value(&action),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	switch action {
	case "backup":
		return runSyncBackup(cmd, args)
	case "restore":
		return runSyncRestore(cmd, args)
	case "list":
		return runSyncList(cmd, args)
	}

	return nil
}

func runSyncBackup(cmd *cobra.Command, args []string) error {
	// Determine output path
	outputDir := syncOutputPath
	if outputDir == "" {
		outputDir = filepath.Join(os.Getenv("HOME"), ".ung", "backups")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Collect all data
	backup := BackupData{
		Version:   "1.0",
		CreatedAt: time.Now(),
	}

	// Load companies
	if err := db.DB.Find(&backup.Companies).Error; err != nil {
		return fmt.Errorf("failed to load companies: %w", err)
	}

	// Load clients
	if err := db.DB.Find(&backup.Clients).Error; err != nil {
		return fmt.Errorf("failed to load clients: %w", err)
	}

	// Load contracts
	if err := db.DB.Find(&backup.Contracts).Error; err != nil {
		return fmt.Errorf("failed to load contracts: %w", err)
	}

	// Load invoices
	if err := db.DB.Find(&backup.Invoices).Error; err != nil {
		return fmt.Errorf("failed to load invoices: %w", err)
	}

	// Load invoice recipients
	if err := db.DB.Find(&backup.InvoiceRecipients).Error; err != nil {
		return fmt.Errorf("failed to load invoice recipients: %w", err)
	}

	// Load invoice line items
	if err := db.DB.Find(&backup.InvoiceLineItems).Error; err != nil {
		return fmt.Errorf("failed to load invoice line items: %w", err)
	}

	// Load tracking sessions
	if err := db.DB.Unscoped().Find(&backup.TrackingSessions).Error; err != nil {
		return fmt.Errorf("failed to load tracking sessions: %w", err)
	}

	// Load expenses
	if err := db.DB.Find(&backup.Expenses).Error; err != nil {
		return fmt.Errorf("failed to load expenses: %w", err)
	}

	// Load recurring invoices
	if err := db.DB.Find(&backup.RecurringInvoices).Error; err != nil {
		return fmt.Errorf("failed to load recurring invoices: %w", err)
	}

	// Generate filename with timestamp
	filename := filepath.Join(outputDir, fmt.Sprintf("ung_backup_%s.json", time.Now().Format("2006-01-02_150405")))

	// Write backup file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backup); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	// Summary
	fmt.Println("\n✓ Backup created successfully!")
	fmt.Printf("  Location: %s\n", filename)
	fmt.Println("\n  Data backed up:")
	fmt.Printf("    • %d companies\n", len(backup.Companies))
	fmt.Printf("    • %d clients\n", len(backup.Clients))
	fmt.Printf("    • %d contracts\n", len(backup.Contracts))
	fmt.Printf("    • %d invoices\n", len(backup.Invoices))
	fmt.Printf("    • %d time entries\n", len(backup.TrackingSessions))
	fmt.Printf("    • %d expenses\n", len(backup.Expenses))
	fmt.Printf("    • %d recurring invoices\n", len(backup.RecurringInvoices))

	return nil
}

func runSyncRestore(cmd *cobra.Command, args []string) error {
	var backupFile string

	if len(args) > 0 {
		backupFile = args[0]
	} else {
		// List available backups
		backupDir := filepath.Join(os.Getenv("HOME"), ".ung", "backups")
		entries, err := os.ReadDir(backupDir)
		if err != nil || len(entries) == 0 {
			return fmt.Errorf("no backups found. Create one with: ung sync backup")
		}

		var options []huh.Option[string]
		for i := len(entries) - 1; i >= 0; i-- {
			entry := entries[i]
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				info, _ := entry.Info()
				label := fmt.Sprintf("%s (%s)", entry.Name(), info.ModTime().Format("Jan 2, 2006 3:04 PM"))
				options = append(options, huh.NewOption(label, filepath.Join(backupDir, entry.Name())))
			}
		}

		if len(options) == 0 {
			return fmt.Errorf("no backup files found")
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Backup to Restore").
					Options(options...).
					Value(&backupFile),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
	}

	// Read backup file
	file, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	var backup BackupData
	if err := json.NewDecoder(file).Decode(&backup); err != nil {
		return fmt.Errorf("failed to parse backup file: %w", err)
	}

	// Show what will be restored
	fmt.Println("\nBackup contents:")
	fmt.Printf("  • %d companies\n", len(backup.Companies))
	fmt.Printf("  • %d clients\n", len(backup.Clients))
	fmt.Printf("  • %d contracts\n", len(backup.Contracts))
	fmt.Printf("  • %d invoices\n", len(backup.Invoices))
	fmt.Printf("  • %d time entries\n", len(backup.TrackingSessions))
	fmt.Printf("  • %d expenses\n", len(backup.Expenses))
	fmt.Printf("  • %d recurring invoices\n", len(backup.RecurringInvoices))

	// Confirm restore
	if !syncForce {
		var confirm bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Restore this backup?").
					Description("This will merge data with your existing database").
					Affirmative("Restore").
					Negative("Cancel").
					Value(&confirm),
			),
		)

		if err := confirmForm.Run(); err != nil || !confirm {
			fmt.Println("Restore cancelled")
			return nil
		}
	}

	// Restore data (using upsert to avoid duplicates)
	restored := map[string]int{}

	// Restore companies
	for _, c := range backup.Companies {
		if err := db.DB.Save(&c).Error; err == nil {
			restored["companies"]++
		}
	}

	// Restore clients
	for _, c := range backup.Clients {
		if err := db.DB.Save(&c).Error; err == nil {
			restored["clients"]++
		}
	}

	// Restore contracts
	for _, c := range backup.Contracts {
		if err := db.DB.Save(&c).Error; err == nil {
			restored["contracts"]++
		}
	}

	// Restore invoices
	for _, inv := range backup.Invoices {
		if err := db.DB.Save(&inv).Error; err == nil {
			restored["invoices"]++
		}
	}

	// Restore invoice recipients
	for _, ir := range backup.InvoiceRecipients {
		if err := db.DB.Save(&ir).Error; err == nil {
			restored["invoice_recipients"]++
		}
	}

	// Restore invoice line items
	for _, li := range backup.InvoiceLineItems {
		if err := db.DB.Save(&li).Error; err == nil {
			restored["line_items"]++
		}
	}

	// Restore tracking sessions
	for _, t := range backup.TrackingSessions {
		if err := db.DB.Save(&t).Error; err == nil {
			restored["time_entries"]++
		}
	}

	// Restore expenses
	for _, e := range backup.Expenses {
		if err := db.DB.Save(&e).Error; err == nil {
			restored["expenses"]++
		}
	}

	// Restore recurring invoices
	for _, r := range backup.RecurringInvoices {
		if err := db.DB.Save(&r).Error; err == nil {
			restored["recurring"]++
		}
	}

	fmt.Println("\n✓ Restore completed!")
	fmt.Println("  Data restored:")
	for k, v := range restored {
		fmt.Printf("    • %d %s\n", v, k)
	}

	return nil
}

func runSyncList(cmd *cobra.Command, args []string) error {
	backupDir := filepath.Join(os.Getenv("HOME"), ".ung", "backups")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		fmt.Println("No backups found. Create one with: ung sync backup")
		return nil
	}

	fmt.Println("\nAvailable backups:")
	fmt.Println("==================")

	count := 0
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			info, _ := entry.Info()
			size := float64(info.Size()) / 1024 // KB
			fmt.Printf("  %s  %.1f KB  %s\n",
				info.ModTime().Format("2006-01-02 15:04"),
				size,
				entry.Name())
			count++
		}
	}

	if count == 0 {
		fmt.Println("  No backups found")
	} else {
		fmt.Printf("\nTotal: %d backups in %s\n", count, backupDir)
	}

	return nil
}
