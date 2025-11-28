package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage user preferences",
	Long: `Manage user preferences and dashboard settings.

Commands:
  target    Set weekly hours target
  streak    Show tracking streak
  get       Get a setting value
  set       Set a setting value
  ls        List all settings`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Run parent's PersistentPreRunE if exists
		if parent := cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
			if err := parent.PersistentPreRunE(parent, args); err != nil {
				return err
			}
		}
		// Migrate the schema
		if db.GormDB != nil {
			db.GormDB.AutoMigrate(&models.UserSettings{})
		}
		return nil
	},
}

var settingsTargetCmd = &cobra.Command{
	Use:   "target [hours]",
	Short: "Get or set weekly hours target",
	Long: `Get or set your weekly hours target for the dashboard progress bar.

Examples:
  ung settings target         Show current target
  ung settings target 40      Set target to 40 hours/week
  ung settings target 35.5    Set target to 35.5 hours/week`,
	RunE: runSettingsTarget,
}

var settingsStreakCmd = &cobra.Command{
	Use:   "streak",
	Short: "Show tracking streak",
	Long:  "Show your current tracking streak (consecutive days with time tracked)",
	RunE:  runSettingsStreak,
}

var settingsGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a setting value",
	Args:  cobra.ExactArgs(1),
	RunE:  runSettingsGet,
}

var settingsSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a setting value",
	Args:  cobra.ExactArgs(2),
	RunE:  runSettingsSet,
}

var settingsListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all settings",
	RunE:  runSettingsList,
}

func init() {
	rootCmd.AddCommand(settingsCmd)
	settingsCmd.AddCommand(settingsTargetCmd)
	settingsCmd.AddCommand(settingsStreakCmd)
	settingsCmd.AddCommand(settingsGetCmd)
	settingsCmd.AddCommand(settingsSetCmd)
	settingsCmd.AddCommand(settingsListCmd)
}

func runSettingsTarget(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// Get current target
		var setting models.UserSettings
		result := db.GormDB.Where("key = ?", models.SettingWeeklyHoursTarget).First(&setting)
		if result.Error != nil {
			fmt.Println("Weekly target: 40h (default)")
			fmt.Println("\nUse 'ung settings target <hours>' to customize")
			return nil
		}
		fmt.Printf("Weekly target: %sh\n", setting.Value)
		return nil
	}

	// Set new target
	hours, err := strconv.ParseFloat(args[0], 64)
	if err != nil || hours <= 0 {
		return fmt.Errorf("invalid hours value: %s (must be a positive number)", args[0])
	}

	if hours > 168 {
		return fmt.Errorf("weekly target cannot exceed 168 hours (7 days × 24 hours)")
	}

	var setting models.UserSettings
	result := db.GormDB.Where("key = ?", models.SettingWeeklyHoursTarget).First(&setting)
	if result.Error != nil {
		// Create new
		setting = models.UserSettings{
			Key:   models.SettingWeeklyHoursTarget,
			Value: fmt.Sprintf("%.1f", hours),
		}
		if err := db.GormDB.Create(&setting).Error; err != nil {
			return fmt.Errorf("failed to save setting: %w", err)
		}
	} else {
		// Update existing
		setting.Value = fmt.Sprintf("%.1f", hours)
		if err := db.GormDB.Save(&setting).Error; err != nil {
			return fmt.Errorf("failed to update setting: %w", err)
		}
	}

	fmt.Printf("Weekly target set to %.1f hours\n", hours)
	return nil
}

func runSettingsStreak(cmd *cobra.Command, args []string) error {
	streak, lastDate := calculateTrackingStreak()

	fmt.Println()
	if streak == 0 {
		fmt.Println("No tracking streak yet")
		fmt.Println("Start tracking to build your streak!")
	} else if streak == 1 {
		fmt.Printf("1 day streak\n")
		fmt.Printf("Last tracked: %s\n", lastDate.Format("Mon, Jan 2"))
	} else {
		fmt.Printf("%d day streak\n", streak)
		fmt.Printf("Last tracked: %s\n", lastDate.Format("Mon, Jan 2"))
	}
	fmt.Println()
	return nil
}

func runSettingsGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	var setting models.UserSettings
	result := db.GormDB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		return fmt.Errorf("setting not found: %s", key)
	}
	fmt.Println(setting.Value)
	return nil
}

func runSettingsSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	var setting models.UserSettings
	result := db.GormDB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		// Create new
		setting = models.UserSettings{
			Key:   key,
			Value: value,
		}
		if err := db.GormDB.Create(&setting).Error; err != nil {
			return fmt.Errorf("failed to save setting: %w", err)
		}
	} else {
		// Update existing
		setting.Value = value
		if err := db.GormDB.Save(&setting).Error; err != nil {
			return fmt.Errorf("failed to update setting: %w", err)
		}
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func runSettingsList(cmd *cobra.Command, args []string) error {
	var settings []models.UserSettings
	if err := db.GormDB.Find(&settings).Error; err != nil {
		return fmt.Errorf("failed to list settings: %w", err)
	}

	if len(settings) == 0 {
		fmt.Println("No custom settings configured.")
		fmt.Println("\nDefault values:")
		fmt.Println("  weekly_hours_target: 40")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tVALUE")
	for _, s := range settings {
		fmt.Fprintf(w, "%s\t%s\n", s.Key, s.Value)
	}
	w.Flush()

	return nil
}

// calculateTrackingStreak calculates consecutive days with time tracked
// Returns streak count and the last tracked date
func calculateTrackingStreak() (int, time.Time) {
	// Get all tracking sessions ordered by date descending
	var sessions []models.TrackingSession
	if err := db.GormDB.Order("start_time DESC").Find(&sessions).Error; err != nil {
		return 0, time.Time{}
	}

	if len(sessions) == 0 {
		return 0, time.Time{}
	}

	// Get unique dates with tracking (in local timezone)
	loc := time.Now().Location()
	trackedDates := make(map[string]bool)
	var lastDate time.Time

	for _, s := range sessions {
		dateStr := s.StartTime.In(loc).Format("2006-01-02")
		if !trackedDates[dateStr] {
			trackedDates[dateStr] = true
			if lastDate.IsZero() {
				lastDate = s.StartTime.In(loc)
			}
		}
	}

	// Count consecutive days starting from today or yesterday
	today := time.Now().In(loc)
	todayStr := today.Format("2006-01-02")
	yesterdayStr := today.AddDate(0, 0, -1).Format("2006-01-02")

	// Start counting from today or yesterday (allow 1 day gap)
	var startDate time.Time
	if trackedDates[todayStr] {
		startDate = today
	} else if trackedDates[yesterdayStr] {
		startDate = today.AddDate(0, 0, -1)
	} else {
		// Streak broken - no tracking today or yesterday
		return 0, lastDate
	}

	streak := 0
	checkDate := startDate
	for {
		dateStr := checkDate.Format("2006-01-02")
		if trackedDates[dateStr] {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, lastDate
}

// GetWeeklyTarget returns the user's weekly hours target (default 40)
func GetWeeklyTarget() float64 {
	var setting models.UserSettings
	result := db.GormDB.Where("key = ?", models.SettingWeeklyHoursTarget).First(&setting)
	if result.Error != nil {
		return 40.0 // Default
	}
	hours, err := strconv.ParseFloat(setting.Value, 64)
	if err != nil {
		return 40.0
	}
	return hours
}

// GetTrackingStreak returns the current tracking streak
func GetTrackingStreak() int {
	streak, _ := calculateTrackingStreak()
	return streak
}
