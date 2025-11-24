package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Time tracking utilities",
	Long:  "Start, stop, and manage time tracking sessions",
}

var trackStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new tracking session",
	RunE:  runTrackStart,
}

var trackStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the current tracking session",
	RunE:  runTrackStop,
}

var trackNowCmd = &cobra.Command{
	Use:   "now",
	Short: "Show current tracking session",
	RunE:  runTrackNow,
}

var trackListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all tracking sessions",
	RunE:    runTrackList,
}

var trackLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Log time manually (without timer)",
	Long:  "Manually log hours worked on a contract or client",
	RunE:  runTrackLog,
}

var (
	trackClientID   int
	trackClientName string
	trackContractID int
	trackProject    string
	trackBillable   bool
	trackNotes      string
	trackHours      float64
)

func init() {
	// Add subcommands
	trackCmd.AddCommand(trackStartCmd)
	trackCmd.AddCommand(trackStopCmd)
	trackCmd.AddCommand(trackNowCmd)
	trackCmd.AddCommand(trackLogCmd)
	trackCmd.AddCommand(trackListCmd)

	// Start flags
	trackStartCmd.Flags().IntVar(&trackClientID, "client", 0, "Client ID")
	trackStartCmd.Flags().StringVar(&trackProject, "project", "", "Project name")
	trackStartCmd.Flags().BoolVar(&trackBillable, "billable", true, "Billable session")
	trackStartCmd.Flags().StringVar(&trackNotes, "notes", "", "Session notes")

	// Log flags
	trackLogCmd.Flags().IntVar(&trackContractID, "contract", 0, "Contract ID")
	trackLogCmd.Flags().StringVar(&trackClientName, "client", "", "Client name (e.g., humbrella)")
	trackLogCmd.Flags().Float64Var(&trackHours, "hours", 0, "Hours worked")
	trackLogCmd.Flags().StringVar(&trackProject, "project", "", "Project name")
	trackLogCmd.Flags().StringVar(&trackNotes, "notes", "", "Session notes")
}

func runTrackStart(cmd *cobra.Command, args []string) error {
	// Check if there's already an active session
	var activeCount int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE end_time IS NULL").Scan(&activeCount)
	if err != nil {
		return fmt.Errorf("failed to check active sessions: %w", err)
	}

	if activeCount > 0 {
		return fmt.Errorf("there is already an active tracking session. Stop it first with 'ung track stop'")
	}

	// Insert new session
	query := `
		INSERT INTO tracking_sessions (client_id, project_name, start_time, billable, notes)
		VALUES (?, ?, ?, ?, ?)
	`

	var clientIDPtr *int
	if trackClientID > 0 {
		clientIDPtr = &trackClientID
	}

	result, err := db.DB.Exec(query, clientIDPtr, trackProject, time.Now(), trackBillable, trackNotes)
	if err != nil {
		return fmt.Errorf("failed to start tracking session: %w", err)
	}

	id, _ := result.LastInsertId()
	fmt.Printf("✓ Tracking session started (ID: %d)\n", id)
	if trackProject != "" {
		fmt.Printf("  Project: %s\n", trackProject)
	}
	if trackClientID > 0 {
		fmt.Printf("  Client ID: %d\n", trackClientID)
	}
	fmt.Printf("  Billable: %v\n", trackBillable)
	return nil
}

func runTrackStop(cmd *cobra.Command, args []string) error {
	// Find active session
	var session models.TrackingSession
	err := db.DB.QueryRow(`
		SELECT id, start_time, project_name
		FROM tracking_sessions
		WHERE end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`).Scan(&session.ID, &session.StartTime, &session.ProjectName)

	if err != nil {
		return fmt.Errorf("no active tracking session found")
	}

	// Stop the session
	endTime := time.Now()
	duration := int(endTime.Sub(session.StartTime).Seconds())

	_, err = db.DB.Exec(`
		UPDATE tracking_sessions
		SET end_time = ?, duration = ?, updated_at = ?
		WHERE id = ?
	`, endTime, duration, time.Now(), session.ID)

	if err != nil {
		return fmt.Errorf("failed to stop tracking session: %w", err)
	}

	hours := duration / 3600
	minutes := (duration % 3600) / 60
	seconds := duration % 60

	fmt.Printf("✓ Tracking session stopped\n")
	if session.ProjectName != "" {
		fmt.Printf("  Project: %s\n", session.ProjectName)
	}
	fmt.Printf("  Duration: %dh %dm %ds\n", hours, minutes, seconds)
	return nil
}

func runTrackNow(cmd *cobra.Command, args []string) error {
	var session models.TrackingSession
	var clientName *string

	err := db.DB.QueryRow(`
		SELECT ts.id, ts.client_id, ts.project_name, ts.start_time, ts.billable, ts.notes,
		       c.name as client_name
		FROM tracking_sessions ts
		LEFT JOIN clients c ON ts.client_id = c.id
		WHERE ts.end_time IS NULL
		ORDER BY ts.start_time DESC
		LIMIT 1
	`).Scan(&session.ID, &session.ClientID, &session.ProjectName,
		&session.StartTime, &session.Billable, &session.Notes, &clientName)

	if err != nil {
		fmt.Println("No active tracking session")
		return nil
	}

	elapsed := time.Since(session.StartTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60

	fmt.Println("Active Tracking Session:")
	fmt.Printf("  ID: %d\n", session.ID)
	if session.ProjectName != "" {
		fmt.Printf("  Project: %s\n", session.ProjectName)
	}
	if clientName != nil {
		fmt.Printf("  Client: %s (ID: %d)\n", *clientName, *session.ClientID)
	}
	fmt.Printf("  Started: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Elapsed: %dh %dm %ds\n", hours, minutes, seconds)
	fmt.Printf("  Billable: %v\n", session.Billable)
	if session.Notes != "" {
		fmt.Printf("  Notes: %s\n", session.Notes)
	}

	return nil
}

func runTrackList(cmd *cobra.Command, args []string) error {
	query := `
		SELECT ts.id, ts.project_name, ts.start_time, ts.end_time, ts.duration, ts.billable,
		       c.name as client_name
		FROM tracking_sessions ts
		LEFT JOIN clients c ON ts.client_id = c.id
		ORDER BY ts.start_time DESC
		LIMIT 50
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tracking sessions: %w", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPROJECT\tCLIENT\tSTART\tDURATION\tBILLABLE")

	for rows.Next() {
		var id int
		var projectName, clientName *string
		var startTime time.Time
		var endTime *time.Time
		var duration *int
		var billable bool

		if err := rows.Scan(&id, &projectName, &startTime, &endTime, &duration, &billable, &clientName); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		project := "-"
		if projectName != nil {
			project = *projectName
		}

		client := "-"
		if clientName != nil {
			client = *clientName
		}

		durationStr := "ongoing"
		if duration != nil {
			hours := *duration / 3600
			minutes := (*duration % 3600) / 60
			durationStr = fmt.Sprintf("%dh %dm", hours, minutes)
		}

		billableStr := "No"
		if billable {
			billableStr = "Yes"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			id, project, client, startTime.Format("2006-01-02 15:04"), durationStr, billableStr)
	}

	w.Flush()
	return nil
}

func runTrackLog(cmd *cobra.Command, args []string) error {
	// Ensure prerequisites exist (company, client, contract)
	if err := ensureContractExists(); err != nil {
		return err
	}

	// If client name provided, look up client and find their active contract
	if trackClientName != "" && trackContractID == 0 {
		var clientID int
		err := db.DB.QueryRow(`
			SELECT id FROM clients
			WHERE LOWER(name) LIKE LOWER(?)
			LIMIT 1
		`, "%"+trackClientName+"%").Scan(&clientID)

		if err != nil {
			return fmt.Errorf("client '%s' not found", trackClientName)
		}

		// Find active contract for this client
		err = db.DB.QueryRow(`
			SELECT id FROM contracts
			WHERE client_id = ? AND active = 1
			ORDER BY id DESC
			LIMIT 1
		`, clientID).Scan(&trackContractID)

		if err != nil {
			return fmt.Errorf("no active contract found for client '%s'", trackClientName)
		}
	}

	// Interactive mode if parameters not provided
	if trackContractID == 0 || trackHours == 0 {
		// Fetch active contracts with client info and hourly rates
		query := `
			SELECT c.id, c.name, c.contract_type, c.hourly_rate, c.currency, cl.name
			FROM contracts c
			JOIN clients cl ON c.client_id = cl.id
			WHERE c.active = 1
			ORDER BY cl.name, c.name
		`

		rows, err := db.DB.Query(query)
		if err != nil {
			return fmt.Errorf("failed to query contracts: %w", err)
		}
		defer rows.Close()

		type contractOption struct {
			id           int
			name         string
			contractType string
			hourlyRate   *float64
			currency     string
			clientName   string
		}

		var contracts []contractOption
		for rows.Next() {
			var c contractOption
			if err := rows.Scan(&c.id, &c.name, &c.contractType, &c.hourlyRate, &c.currency, &c.clientName); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			contracts = append(contracts, c)
		}

		if len(contracts) == 0 {
			return fmt.Errorf("no active contracts found. Add a contract first with 'ung contract add'")
		}

		// Build contract options with rates
		contractOptions := make([]huh.Option[int], len(contracts))
		for i, c := range contracts {
			label := fmt.Sprintf("%s - %s (%s)", c.clientName, c.name, c.contractType)
			if c.contractType == "hourly" && c.hourlyRate != nil {
				label = fmt.Sprintf("%s - %s (%.0f %s/hr)", c.clientName, c.name, *c.hourlyRate, c.currency)
			}
			contractOptions[i] = huh.NewOption(label, c.id)
		}

		var selectedContractID int
		var selectedHoursStr string
		var selectedProject string
		var selectedNotes string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select Contract").
					Description("Choose which contract/client you worked on").
					Options(contractOptions...).
					Value(&selectedContractID),

				huh.NewInput().
					Title("Hours Worked").
					Placeholder("e.g., 2.5").
					Value(&selectedHoursStr).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("hours is required")
						}
						hours, err := strconv.ParseFloat(s, 64)
						if err != nil || hours <= 0 {
							return fmt.Errorf("hours must be a positive number")
						}
						return nil
					}),

				huh.NewInput().
					Title("Project/Task Name (optional)").
					Placeholder("e.g., Homepage redesign").
					Value(&selectedProject),

				huh.NewText().
					Title("Notes (optional)").
					Placeholder("What did you work on?").
					Value(&selectedNotes),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("form cancelled: %w", err)
		}

		// Convert hours string to float64
		hours, err := strconv.ParseFloat(selectedHoursStr, 64)
		if err != nil {
			return fmt.Errorf("invalid hours: %w", err)
		}

		trackContractID = selectedContractID
		trackHours = hours
		trackProject = selectedProject
		trackNotes = selectedNotes
	}

	// Get contract and client info
	var clientID int
	var contractName, clientName string
	var hourlyRate *float64
	var currency string

	err := db.DB.QueryRow(`
		SELECT c.client_id, c.name, c.hourly_rate, c.currency, cl.name
		FROM contracts c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.id = ?
	`, trackContractID).Scan(&clientID, &contractName, &hourlyRate, &currency, &clientName)

	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	// Calculate duration in seconds and set hours
	durationSecs := int(trackHours * 3600)

	// Insert tracking session
	query := `
		INSERT INTO tracking_sessions
		(client_id, contract_id, project_name, start_time, end_time, duration, hours, billable, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?)
	`

	now := time.Now()
	endTime := now.Add(time.Duration(durationSecs) * time.Second)

	result, err := db.DB.Exec(query,
		clientID, trackContractID, trackProject,
		now, endTime, durationSecs, trackHours, trackNotes)

	if err != nil {
		return fmt.Errorf("failed to log time: %w", err)
	}

	id, _ := result.LastInsertId()

	fmt.Printf("✓ Time logged successfully (Session ID: %d)\n", id)
	fmt.Printf("  Client: %s\n", clientName)
	fmt.Printf("  Contract: %s\n", contractName)
	fmt.Printf("  Hours: %.2f\n", trackHours)
	if hourlyRate != nil {
		total := trackHours * (*hourlyRate)
		fmt.Printf("  Billable Amount: %.2f %s\n", total, currency)
	}
	if trackProject != "" {
		fmt.Printf("  Project: %s\n", trackProject)
	}

	return nil
}
