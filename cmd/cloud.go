package cmd

import (
	"fmt"

	"github.com/Andriiklymiuk/ung/internal/cloud"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/spf13/cobra"
)

var cloudCmd = &cobra.Command{
	Use:     "cloud",
	Aliases: []string{"icloud"},
	Short:   "Manage cloud storage and sync",
	Long: `Manage iCloud Drive storage and synchronization.

iCloud Drive provides automatic sync across all your Apple devices (Mac, iPhone, iPad)
without needing a server. Your data is encrypted and synced automatically.

Examples:
  ung cloud status         # Check iCloud sync status
  ung cloud available      # Check if iCloud is available
  ung cloud path           # Show iCloud paths`,
}

var cloudStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show iCloud sync status",
	Long:  `Display the current iCloud synchronization status for your database and invoices.`,
	Run:   runCloudStatus,
}

var cloudAvailableCmd = &cobra.Command{
	Use:   "available",
	Short: "Check if iCloud Drive is available",
	Long:  `Check if iCloud Drive is available and accessible on this system.`,
	Run:   runCloudAvailable,
}

var cloudPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show iCloud Drive paths",
	Long:  `Display iCloud Drive paths for app data storage.`,
	Run:   runCloudPath,
}

func init() {
	cloudCmd.AddCommand(cloudStatusCmd)
	cloudCmd.AddCommand(cloudAvailableCmd)
	cloudCmd.AddCommand(cloudPathCmd)
	rootCmd.AddCommand(cloudCmd)
}

func runCloudStatus(cmd *cobra.Command, args []string) {
	dbPath := db.GetDBPath()
	invoicesDir := db.GetInvoicesDir()

	fmt.Println("☁️  iCloud Sync Status")

	// Check if database is in iCloud
	dbInCloud := cloud.IsSyncedToiCloud(dbPath)
	invoicesInCloud := cloud.IsSyncedToiCloud(invoicesDir)

	if !dbInCloud && !invoicesInCloud {
		fmt.Println("Status:     Not using iCloud")
		fmt.Println("Location:   Local storage")
		fmt.Println("\n💡 To enable iCloud sync:")
		fmt.Println("   ung config init --icloud")
		return
	}

	if dbInCloud {
		dbStatus := cloud.GetSyncStatus(dbPath)
		statusIcon := getStatusIcon(dbStatus)

		fmt.Printf("Database:   %s %s\n", statusIcon, dbStatus)
		fmt.Printf("Path:       %s\n", dbPath)
	}

	if invoicesInCloud {
		invoicesStatus := cloud.GetSyncStatus(invoicesDir)
		statusIcon := getStatusIcon(invoicesStatus)

		fmt.Printf("\nInvoices:   %s %s\n", statusIcon, invoicesStatus)
		fmt.Printf("Path:       %s\n", invoicesDir)
	}

	if dbInCloud || invoicesInCloud {
		fmt.Println("\n✨ Features:")
		fmt.Println("   • Automatic sync across Mac, iPhone, iPad")
		fmt.Println("   • No server required")
		fmt.Println("   • iOS app can access same data")
		fmt.Println("   • Works offline (queues changes)")
		fmt.Println("\n💡 Combine with encryption for extra security:")
		fmt.Println("   ung security enable")
	}
}

func runCloudAvailable(cmd *cobra.Command, args []string) {
	fmt.Println("☁️  iCloud Drive Availability")

	available := cloud.ICloudAvailable()

	if available {
		fmt.Println("Status:     ✅ Available")

		iCloudPath := cloud.GetICloudDrivePath()
		fmt.Printf("Location:   %s\n", iCloudPath)

		appPath, err := cloud.GetICloudAppPath()
		if err == nil {
			fmt.Printf("App Folder: %s\n", appPath)
		}

		fmt.Println("\n💡 You can use iCloud sync:")
		fmt.Println("   ung config init --icloud")
	} else {
		fmt.Println("Status:     ❌ Not Available")
		fmt.Println("\nReasons:")
		fmt.Println("   • Not running on macOS")
		fmt.Println("   • iCloud Drive not enabled")
		fmt.Println("   • iCloud Drive folder not accessible")

		fmt.Println("\n💡 To enable iCloud Drive:")
		fmt.Println("   1. Open System Settings")
		fmt.Println("   2. Go to Apple ID → iCloud")
		fmt.Println("   3. Enable 'iCloud Drive'")
	}
}

func runCloudPath(cmd *cobra.Command, args []string) {
	fmt.Println("☁️  iCloud Drive Paths")

	if !cloud.ICloudAvailable() {
		fmt.Println("❌ iCloud Drive is not available")
		return
	}

	iCloudRoot := cloud.GetICloudDrivePath()
	fmt.Printf("iCloud Drive:  %s\n", iCloudRoot)

	appPath, err := cloud.GetICloudAppPath()
	if err != nil {
		fmt.Printf("❌ Failed to get app path: %v\n", err)
		return
	}

	fmt.Printf("App Folder:    %s\n", appPath)

	dbPath, err := cloud.GetICloudDatabasePath("ung.db")
	if err == nil {
		fmt.Printf("Database:      %s\n", dbPath)
	}

	invoicesPath, err := cloud.GetICloudInvoicesPath()
	if err == nil {
		fmt.Printf("Invoices:      %s\n", invoicesPath)
	}

	fmt.Println("\n💡 These paths are shared across:")
	fmt.Println("   • All Macs signed in to your Apple ID")
	fmt.Println("   • iPhone (with iOS app)")
	fmt.Println("   • iPad (with iOS app)")
}

func getStatusIcon(status string) string {
	switch status {
	case "synced":
		return "✅"
	case "uploading":
		return "⬆️"
	case "downloading":
		return "⬇️"
	case "not-in-icloud":
		return "⚪"
	default:
		return "❓"
	}
}
