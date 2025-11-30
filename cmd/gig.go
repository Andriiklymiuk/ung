package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var gigCmd = &cobra.Command{
	Use:   "gig",
	Short: "Manage your gigs (kanban board)",
	Long:  "Create, view, and manage gigs through a workflow: pipeline → negotiating → active → delivered → invoiced → complete",
}

var gigListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all gigs",
	RunE:  runGigList,
}

var gigAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new gig",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runGigAdd,
}

var gigMoveCmd = &cobra.Command{
	Use:   "move <gig-id> <status>",
	Short: "Move a gig to a different status",
	Long: `Move a gig through the workflow. Available statuses:
  - pipeline     (potential work)
  - negotiating  (in discussion)
  - active       (currently working)
  - delivered    (work done, awaiting approval)
  - invoiced     (invoice sent)
  - complete     (paid and done)
  - on_hold      (paused)
  - cancelled    (cancelled)`,
	Args: cobra.ExactArgs(2),
	RunE: runGigMove,
}

var gigShowCmd = &cobra.Command{
	Use:   "show <gig-id>",
	Short: "Show gig details and tasks",
	Args:  cobra.ExactArgs(1),
	RunE:  runGigShow,
}

var gigDeleteCmd = &cobra.Command{
	Use:   "delete <gig-id>",
	Short: "Delete a gig",
	Args:  cobra.ExactArgs(1),
	RunE:  runGigDelete,
}

// Task subcommands
var gigTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks within gigs",
}

var gigTaskAddCmd = &cobra.Command{
	Use:   "add <gig-id> <title>",
	Short: "Add a task to a gig",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runGigTaskAdd,
}

var gigTaskDoneCmd = &cobra.Command{
	Use:   "done <task-id>",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),
	RunE:  runGigTaskDone,
}

var gigTaskDeleteCmd = &cobra.Command{
	Use:   "delete <task-id>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE:  runGigTaskDelete,
}

// Flags
var (
	gigClientID int
	gigStatus   string
	gigType     string
	gigRate     float64
)

func init() {
	rootCmd.AddCommand(gigCmd)

	// Add subcommands
	gigCmd.AddCommand(gigListCmd)
	gigCmd.AddCommand(gigAddCmd)
	gigCmd.AddCommand(gigMoveCmd)
	gigCmd.AddCommand(gigShowCmd)
	gigCmd.AddCommand(gigDeleteCmd)
	gigCmd.AddCommand(gigTaskCmd)

	// Task subcommands
	gigTaskCmd.AddCommand(gigTaskAddCmd)
	gigTaskCmd.AddCommand(gigTaskDoneCmd)
	gigTaskCmd.AddCommand(gigTaskDeleteCmd)

	// Flags for add
	gigAddCmd.Flags().IntVarP(&gigClientID, "client", "c", 0, "Client ID")
	gigAddCmd.Flags().StringVarP(&gigStatus, "status", "s", "pipeline", "Initial status")
	gigAddCmd.Flags().StringVarP(&gigType, "type", "t", "hourly", "Gig type: hourly, fixed, retainer")
	gigAddCmd.Flags().Float64VarP(&gigRate, "rate", "r", 0, "Hourly rate")

	// Flags for list
	gigListCmd.Flags().StringVarP(&gigStatus, "status", "s", "", "Filter by status")
}

func runGigList(cmd *cobra.Command, args []string) error {
	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3366FF"))

	statusStyle := func(status models.GigStatus) lipgloss.Style {
		colors := map[models.GigStatus]string{
			models.GigStatusPipeline:    "#888888",
			models.GigStatusNegotiating: "#9966FF",
			models.GigStatusActive:      "#3366FF",
			models.GigStatusDelivered:   "#FF9933",
			models.GigStatusInvoiced:    "#33CCFF",
			models.GigStatusComplete:    "#22CC55",
			models.GigStatusOnHold:      "#FFCC00",
			models.GigStatusCancelled:   "#FF3333",
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(colors[status]))
	}

	var gigs []models.Gig
	query := db.GormDB.Preload("Client")

	if gigStatus != "" {
		query = query.Where("status = ?", gigStatus)
	}

	if err := query.Order("priority DESC, updated_at DESC").Find(&gigs).Error; err != nil {
		return fmt.Errorf("failed to fetch gigs: %w", err)
	}

	if len(gigs) == 0 {
		fmt.Println("No gigs found. Create one with: ung gig add <name>")
		return nil
	}

	fmt.Println()
	fmt.Println(headerStyle.Render("GIGS"))
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCLIENT\tSTATUS\tHOURS\tTYPE")
	fmt.Fprintln(w, "--\t----\t------\t------\t-----\t----")

	for _, gig := range gigs {
		clientName := "-"
		if gig.Client != nil {
			clientName = gig.Client.Name
		}

		status := statusStyle(gig.Status).Render(string(gig.Status))
		hours := fmt.Sprintf("%.1f", gig.TotalHoursTracked)

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			gig.ID,
			truncate(gig.Name, 25),
			truncate(clientName, 15),
			status,
			hours,
			gig.GigType,
		)
	}

	w.Flush()
	fmt.Println()

	// Show summary by status
	fmt.Println(headerStyle.Render("SUMMARY"))
	statusCounts := make(map[models.GigStatus]int)
	for _, gig := range gigs {
		statusCounts[gig.Status]++
	}

	statuses := []models.GigStatus{
		models.GigStatusPipeline,
		models.GigStatusNegotiating,
		models.GigStatusActive,
		models.GigStatusDelivered,
		models.GigStatusInvoiced,
		models.GigStatusComplete,
	}

	var parts []string
	for _, s := range statuses {
		if count := statusCounts[s]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s: %d", statusStyle(s).Render(string(s)), count))
		}
	}
	fmt.Println("  " + strings.Join(parts, "  |  "))
	fmt.Println()

	return nil
}

func runGigAdd(cmd *cobra.Command, args []string) error {
	name := strings.Join(args, " ")

	gig := models.Gig{
		Name:    name,
		Status:  models.GigStatus(gigStatus),
		GigType: models.GigType(gigType),
	}

	if gigClientID > 0 {
		clientID := uint(gigClientID)
		gig.ClientID = &clientID
	}

	if gigRate > 0 {
		gig.HourlyRate = &gigRate
	}

	if err := db.GormDB.Create(&gig).Error; err != nil {
		return fmt.Errorf("failed to create gig: %w", err)
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC55"))
	fmt.Printf("%s Created gig #%d: %s\n", successStyle.Render("✓"), gig.ID, gig.Name)

	return nil
}

func runGigMove(cmd *cobra.Command, args []string) error {
	gigID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid gig ID: %w", err)
	}

	newStatus := models.GigStatus(args[1])

	// Validate status
	validStatuses := []models.GigStatus{
		models.GigStatusPipeline,
		models.GigStatusNegotiating,
		models.GigStatusActive,
		models.GigStatusDelivered,
		models.GigStatusInvoiced,
		models.GigStatusComplete,
		models.GigStatusOnHold,
		models.GigStatusCancelled,
	}

	valid := false
	for _, s := range validStatuses {
		if s == newStatus {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	var gig models.Gig
	if err := db.GormDB.First(&gig, gigID).Error; err != nil {
		return fmt.Errorf("gig not found: %w", err)
	}

	oldStatus := gig.Status
	gig.Status = newStatus

	if err := db.GormDB.Save(&gig).Error; err != nil {
		return fmt.Errorf("failed to update gig: %w", err)
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC55"))
	fmt.Printf("%s Moved gig #%d from %s → %s\n",
		successStyle.Render("✓"),
		gig.ID,
		oldStatus,
		newStatus,
	)

	return nil
}

func runGigShow(cmd *cobra.Command, args []string) error {
	gigID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid gig ID: %w", err)
	}

	var gig models.Gig
	if err := db.GormDB.Preload("Client").First(&gig, gigID).Error; err != nil {
		return fmt.Errorf("gig not found: %w", err)
	}

	// Get tasks
	var tasks []models.GigTask
	db.GormDB.Where("gig_id = ?", gigID).Order("sort_order ASC").Find(&tasks)

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3366FF"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	valueStyle := lipgloss.NewStyle().
		Bold(true)

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22CC55"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("GIG #%d: %s", gig.ID, gig.Name)))
	fmt.Println()

	// Details
	clientName := "-"
	if gig.Client != nil {
		clientName = gig.Client.Name
	}

	fmt.Printf("  %s %s\n", labelStyle.Render("Status:"), valueStyle.Render(string(gig.Status)))
	fmt.Printf("  %s %s\n", labelStyle.Render("Client:"), clientName)
	fmt.Printf("  %s %s\n", labelStyle.Render("Type:"), string(gig.GigType))
	fmt.Printf("  %s %.1f hours\n", labelStyle.Render("Tracked:"), gig.TotalHoursTracked)

	if gig.HourlyRate != nil {
		fmt.Printf("  %s $%.0f/hr\n", labelStyle.Render("Rate:"), *gig.HourlyRate)
	}

	if gig.Description != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Description:"), gig.Description)
	}

	// Tasks
	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("TASKS (%d)", len(tasks))))

	if len(tasks) == 0 {
		fmt.Printf("  %s\n", mutedStyle.Render("No tasks. Add one with: ung gig task add "+args[0]+" <title>"))
	} else {
		for _, task := range tasks {
			checkbox := "○"
			titleStyle := lipgloss.NewStyle()
			if task.Completed {
				checkbox = successStyle.Render("✓")
				titleStyle = mutedStyle.Copy().Strikethrough(true)
			}
			fmt.Printf("  %s [%d] %s\n", checkbox, task.ID, titleStyle.Render(task.Title))
		}
	}

	fmt.Println()

	return nil
}

func runGigDelete(cmd *cobra.Command, args []string) error {
	gigID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid gig ID: %w", err)
	}

	var gig models.Gig
	if err := db.GormDB.First(&gig, gigID).Error; err != nil {
		return fmt.Errorf("gig not found: %w", err)
	}

	// Delete tasks first
	db.GormDB.Where("gig_id = ?", gigID).Delete(&models.GigTask{})

	// Delete gig
	if err := db.GormDB.Delete(&gig).Error; err != nil {
		return fmt.Errorf("failed to delete gig: %w", err)
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC55"))
	fmt.Printf("%s Deleted gig #%d: %s\n", successStyle.Render("✓"), gig.ID, gig.Name)

	return nil
}

func runGigTaskAdd(cmd *cobra.Command, args []string) error {
	gigID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid gig ID: %w", err)
	}

	title := strings.Join(args[1:], " ")

	// Verify gig exists
	var gig models.Gig
	if err := db.GormDB.First(&gig, gigID).Error; err != nil {
		return fmt.Errorf("gig not found: %w", err)
	}

	// Get max sort order
	var maxOrder int
	db.GormDB.Model(&models.GigTask{}).
		Where("gig_id = ?", gigID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder)

	task := models.GigTask{
		GigID:     uint(gigID),
		Title:     title,
		SortOrder: maxOrder + 1,
	}

	if err := db.GormDB.Create(&task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC55"))
	fmt.Printf("%s Added task #%d to gig #%d: %s\n",
		successStyle.Render("✓"),
		task.ID,
		gigID,
		task.Title,
	)

	return nil
}

func runGigTaskDone(cmd *cobra.Command, args []string) error {
	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	var task models.GigTask
	if err := db.GormDB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	task.Completed = !task.Completed
	if task.Completed {
		now := db.GormDB.NowFunc()
		task.CompletedAt = &now
	} else {
		task.CompletedAt = nil
	}

	if err := db.GormDB.Save(&task).Error; err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC55"))
	status := "completed"
	if !task.Completed {
		status = "uncompleted"
	}
	fmt.Printf("%s Task #%d marked as %s\n", successStyle.Render("✓"), task.ID, status)

	return nil
}

func runGigTaskDelete(cmd *cobra.Command, args []string) error {
	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	var task models.GigTask
	if err := db.GormDB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if err := db.GormDB.Delete(&task).Error; err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#22CC55"))
	fmt.Printf("%s Deleted task #%d: %s\n", successStyle.Render("✓"), task.ID, task.Title)

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
