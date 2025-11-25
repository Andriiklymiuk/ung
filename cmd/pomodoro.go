package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var pomodoroCmd = &cobra.Command{
	Use:     "pomodoro",
	Aliases: []string{"pomo", "focus"},
	Short:   "Start a Pomodoro focus timer",
	Long: `Start a Pomodoro timer for focused work sessions.

The Pomodoro Technique uses 25-minute work sessions followed by
5-minute breaks. After 4 sessions, take a longer 15-minute break.

Examples:
  ung pomodoro                Start default 25-minute session
  ung pomodoro --work 30      Start 30-minute work session
  ung pomodoro --break 10     Set break to 10 minutes
  ung pomodoro --client acme  Track time to a client when done`,
	RunE: runPomodoro,
}

var (
	pomodoroWorkMinutes    int
	pomodoroBreakMinutes   int
	pomodoroLongBreak      int
	pomodoroSessions       int
	pomodoroClient         string
	pomodoroProject        string
	pomodoroAutoTrack      bool
)

func init() {
	pomodoroCmd.Flags().IntVarP(&pomodoroWorkMinutes, "work", "w", 25, "Work session duration in minutes")
	pomodoroCmd.Flags().IntVarP(&pomodoroBreakMinutes, "break", "b", 5, "Short break duration in minutes")
	pomodoroCmd.Flags().IntVar(&pomodoroLongBreak, "long-break", 15, "Long break duration (after 4 sessions)")
	pomodoroCmd.Flags().IntVarP(&pomodoroSessions, "sessions", "s", 4, "Number of sessions before long break")
	pomodoroCmd.Flags().StringVarP(&pomodoroClient, "client", "c", "", "Client to track time to")
	pomodoroCmd.Flags().StringVarP(&pomodoroProject, "project", "p", "", "Project name")
	pomodoroCmd.Flags().BoolVar(&pomodoroAutoTrack, "track", false, "Auto-log time when session completes")

	rootCmd.AddCommand(pomodoroCmd)
}

func runPomodoro(cmd *cobra.Command, args []string) error {
	fmt.Println("🍅 Pomodoro Timer")
	fmt.Println("─────────────────")
	fmt.Printf("Work: %d min | Break: %d min | Long break: %d min\n", pomodoroWorkMinutes, pomodoroBreakMinutes, pomodoroLongBreak)
	if pomodoroClient != "" {
		fmt.Printf("Tracking to: %s", pomodoroClient)
		if pomodoroProject != "" {
			fmt.Printf(" / %s", pomodoroProject)
		}
		fmt.Println()
	}
	fmt.Println()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sessionCount := 0
	totalMinutes := 0

	for {
		sessionCount++
		fmt.Printf("📍 Session %d/%d\n", sessionCount, pomodoroSessions)

		// Work session
		if !runTimer("🔴 FOCUS", pomodoroWorkMinutes, sigChan) {
			break // User cancelled
		}

		totalMinutes += pomodoroWorkMinutes

		// Play notification sound (if available)
		playNotification()

		// Determine break type
		isLongBreak := sessionCount%pomodoroSessions == 0
		breakMinutes := pomodoroBreakMinutes
		breakLabel := "☕ SHORT BREAK"
		if isLongBreak {
			breakMinutes = pomodoroLongBreak
			breakLabel = "🌴 LONG BREAK"
		}

		// Ask to continue
		fmt.Printf("\n✓ Session complete! Total: %d minutes\n", totalMinutes)

		var continueSession bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Take a %d-minute %s?", breakMinutes, breakLabel)).
					Value(&continueSession),
			),
		)

		if err := form.Run(); err != nil || !continueSession {
			break
		}

		// Break
		if !runTimer(breakLabel, breakMinutes, sigChan) {
			break
		}

		playNotification()

		// After long break, reset session count
		if isLongBreak {
			sessionCount = 0
			fmt.Println("\n🎉 Completed a full Pomodoro cycle!")
		}

		// Ask to continue
		var startNext bool
		nextForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Start next work session?").
					Value(&startNext),
			),
		)

		if err := nextForm.Run(); err != nil || !startNext {
			break
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("\n─────────────────")
	fmt.Printf("✓ Total focused time: %d minutes (%.1f hours)\n", totalMinutes, float64(totalMinutes)/60)

	// Auto-track time if enabled
	if pomodoroAutoTrack && pomodoroClient != "" && totalMinutes > 0 {
		hours := float64(totalMinutes) / 60
		fmt.Printf("\n📝 Logging %.2f hours to %s...\n", hours, pomodoroClient)

		// Use tracking command to log time
		err := logPomodoroTime(pomodoroClient, pomodoroProject, hours)
		if err != nil {
			fmt.Printf("⚠ Failed to log time: %v\n", err)
		} else {
			fmt.Println("✓ Time logged successfully!")
		}
	} else if totalMinutes > 0 && pomodoroClient == "" {
		fmt.Println("\n💡 Tip: Use --client <name> --track to auto-log time")
	}

	return nil
}

// runTimer displays a countdown timer and returns true if completed, false if cancelled
func runTimer(label string, minutes int, sigChan chan os.Signal) bool {
	duration := time.Duration(minutes) * time.Minute
	endTime := time.Now().Add(duration)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\n⏹ Timer stopped")
			return false

		case <-ticker.C:
			remaining := time.Until(endTime)
			if remaining <= 0 {
				fmt.Printf("\r%s [████████████████████] 00:00 ✓\n", label)
				return true
			}

			// Calculate progress
			elapsed := duration - remaining
			progress := float64(elapsed) / float64(duration)
			barWidth := 20
			filled := int(progress * float64(barWidth))

			// Build progress bar
			bar := ""
			for i := 0; i < barWidth; i++ {
				if i < filled {
					bar += "█"
				} else {
					bar += "░"
				}
			}

			// Format remaining time
			mins := int(remaining.Minutes())
			secs := int(remaining.Seconds()) % 60

			fmt.Printf("\r%s [%s] %02d:%02d ", label, bar, mins, secs)
		}
	}
}

// playNotification attempts to play a notification sound
func playNotification() {
	// Print bell character (works in most terminals)
	fmt.Print("\a")
}

// logPomodoroTime logs the pomodoro session to time tracking
func logPomodoroTime(clientName, project string, hours float64) error {
	// Find client
	clientID, _, err := FindClientByName(clientName)
	if err != nil {
		return err
	}

	// Create tracking session
	now := time.Now()
	startTime := now.Add(-time.Duration(hours*60) * time.Minute)
	durationSeconds := int(hours * 3600)

	projectName := project
	if projectName == "" {
		projectName = "Pomodoro session"
	}

	_, err = db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, project_name, start_time, end_time, duration, hours, billable, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?)
	`, clientID, projectName, startTime, now, durationSeconds, hours, "Pomodoro focus session", now, now)

	return err
}
