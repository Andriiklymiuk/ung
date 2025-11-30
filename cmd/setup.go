package cmd

import (
	"fmt"
	"time"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
	"github.com/Andriiklymiuk/ung/internal/repository"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Quick start wizard - get running in 2 minutes",
	Long: `Interactive setup wizard that guides you through the essential steps
to start tracking time and creating invoices.

This wizard will help you:
  1. Initialize your workspace (if needed)
  2. Add your company details
  3. Add your first client
  4. Set an income goal (optional)

After setup, you'll be ready to track time immediately!`,
	RunE: runSetup,
}

var setupSkipGoal bool

func init() {
	setupCmd.Flags().BoolVar(&setupSkipGoal, "skip-goal", false, "Skip the income goal setup")
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		MarginBottom(1)

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22C55E"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	// Welcome header
	fmt.Println()
	fmt.Println(headerStyle.Render("  Quick Start Wizard"))
	fmt.Println(mutedStyle.Render("  Get ready to track time and invoice clients in 2 minutes"))
	fmt.Println()

	// Step tracking
	totalSteps := 4
	currentStep := 0

	printStep := func(step int, name string) {
		fmt.Printf("\n%s Step %d/%d: %s\n", successStyle.Render(""), step, totalSteps, name)
		fmt.Println(mutedStyle.Render("─────────────────────────────────────"))
	}

	// ============ STEP 1: Initialize if needed ============
	currentStep++
	printStep(currentStep, "Workspace Setup")

	// Check if already initialized
	needsInit := false
	if err := db.Initialize(); err != nil {
		if err == db.ErrNotInitialized {
			needsInit = true
		} else {
			return fmt.Errorf("database error: %w", err)
		}
	}

	if needsInit {
		var initChoice string
		initForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Where should UNG store your data?").
					Description("Global is recommended for most users").
					Options(
						huh.NewOption("Global (~/.ung/) - Access from anywhere", "global"),
						huh.NewOption("Local (.ung/) - Just this project", "local"),
					).
					Value(&initChoice),
			),
		)

		if err := initForm.Run(); err != nil {
			return fmt.Errorf("setup cancelled")
		}

		isGlobal := initChoice == "global"
		config.SetForceGlobal(isGlobal)

		if err := config.Initialize(isGlobal, false, ""); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		if err := db.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		fmt.Printf("%s Workspace created!\n", successStyle.Render(""))
	} else {
		fmt.Printf("%s Workspace already configured\n", successStyle.Render(""))
	}

	// ============ STEP 2: Company Setup ============
	currentStep++
	printStep(currentStep, "Your Business")

	companyRepo := repository.NewCompanyRepository()
	companyCount, err := companyRepo.Count()
	if err != nil {
		fmt.Printf("%s Could not check company status. Continuing...\n", mutedStyle.Render("!"))
		companyCount = 0
	}

	if companyCount == 0 {
		var companyName, companyEmail, companyAddress string

		companyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("What's your business name?").
					Description("Appears on all your invoices").
					Placeholder("e.g., Jane Smith Consulting").
					Value(&companyName).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("business name is required")
						}
						return nil
					}),
				huh.NewInput().
					Title("Business email").
					Description("Where clients reply to invoices").
					Placeholder("e.g., hello@example.com").
					Value(&companyEmail).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("email is required")
						}
						return nil
					}),
				huh.NewInput().
					Title("Address (optional)").
					Description("Shows on invoices if your business requires it").
					Placeholder("e.g., 123 Main St, City, Country").
					Value(&companyAddress),
			),
		)

		if err := companyForm.Run(); err != nil {
			return fmt.Errorf("setup cancelled")
		}

		company := &models.Company{
			Name:    companyName,
			Email:   companyEmail,
			Address: companyAddress,
		}

		if err := companyRepo.Create(company); err != nil {
			return fmt.Errorf("failed to save company: %w", err)
		}

		fmt.Printf("%s Company \"%s\" saved!\n", successStyle.Render(""), companyName)
	} else {
		companies, err := companyRepo.List()
		if err == nil && len(companies) > 0 {
			fmt.Printf("%s Using existing company: %s\n", successStyle.Render(""), companies[0].Name)
		} else {
			fmt.Printf("%s Company configured\n", successStyle.Render(""))
		}
	}

	// ============ STEP 3: First Client ============
	currentStep++
	printStep(currentStep, "Your First Client")

	clientRepo := repository.NewClientRepository()
	clientCount, err := clientRepo.Count()
	if err != nil {
		fmt.Printf("%s Could not check client status. Continuing...\n", mutedStyle.Render("!"))
		clientCount = 0
	}

	if clientCount == 0 {
		var addClient bool
		addClientForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add a client now?").
					Description("You can skip and add clients later").
					Affirmative("Yes, add one").
					Negative("Skip for now").
					Value(&addClient),
			),
		)

		if err := addClientForm.Run(); err != nil {
			return fmt.Errorf("setup cancelled")
		}

		if addClient {
			var clientName, clientEmail string

			clientForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Client/Company name").
						Description("Who you're billing - appears on invoices").
						Placeholder("e.g., Acme Corp").
						Value(&clientName).
						Validate(func(s string) error {
							if s == "" {
								return fmt.Errorf("client name is required")
							}
							return nil
						}),
					huh.NewInput().
						Title("Client email").
						Description("Where to send invoices").
						Placeholder("e.g., billing@acme.com").
						Value(&clientEmail).
						Validate(func(s string) error {
							if s == "" {
								return fmt.Errorf("email is required")
							}
							return nil
						}),
				),
			)

			if err := clientForm.Run(); err != nil {
				return fmt.Errorf("setup cancelled")
			}

			client := &models.Client{
				Name:  clientName,
				Email: clientEmail,
			}

			if err := clientRepo.Create(client); err != nil {
				return fmt.Errorf("failed to save client: %w", err)
			}

			fmt.Printf("%s Client \"%s\" added!\n", successStyle.Render(""), clientName)
		} else {
			fmt.Printf("%s Skipped - add clients later with: ung client add\n", mutedStyle.Render(""))
		}
	} else {
		clients, err := clientRepo.List()
		if err == nil && len(clients) > 0 {
			fmt.Printf("%s You have %d client(s) already\n", successStyle.Render(""), len(clients))
		} else {
			fmt.Printf("%s Clients configured\n", successStyle.Render(""))
		}
	}

	// ============ STEP 4: Income Goal (Optional) ============
	currentStep++
	printStep(currentStep, "Monthly Goal (Optional)")

	if !setupSkipGoal {
		// Check for existing goal
		var existingGoal IncomeGoal
		year := time.Now().Year()
		month := int(time.Now().Month())

		hasGoal := db.GormDB.Where("period = ? AND year = ? AND month = ?", "monthly", year, month).First(&existingGoal).Error == nil

		if !hasGoal {
			var setGoal bool
			goalConfirmForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("Set a monthly income goal?").
						Description("See progress on your dashboard - motivates you to hit targets").
						Affirmative("Yes, set a goal").
						Negative("Skip for now").
						Value(&setGoal),
				),
			)

			if err := goalConfirmForm.Run(); err != nil {
				return fmt.Errorf("setup cancelled")
			}

			if setGoal {
				var goalAmount string
				goalForm := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title(fmt.Sprintf("Monthly target for %s %d ($)", time.Now().Month().String(), year)).
							Description("Enter amount in dollars (e.g., 5000). You can change this anytime.").
							Placeholder("e.g., 5000").
							Value(&goalAmount).
							Validate(func(s string) error {
								if s == "" {
									return fmt.Errorf("amount is required")
								}
								var amount float64
								if _, err := fmt.Sscanf(s, "%f", &amount); err != nil || amount <= 0 {
									return fmt.Errorf("enter a positive number like 5000")
								}
								return nil
							}),
					),
				)

				if err := goalForm.Run(); err != nil {
					return fmt.Errorf("setup cancelled")
				}

				var amount float64
				fmt.Sscanf(goalAmount, "%f", &amount)

				goal := IncomeGoal{
					Amount: amount,
					Period: "monthly",
					Year:   year,
					Month:  month,
				}

				if err := db.GormDB.Create(&goal).Error; err != nil {
					return fmt.Errorf("failed to save goal: %w", err)
				}

				fmt.Printf("%s Goal set: $%.0f for %s\n", successStyle.Render(""), amount, time.Now().Month().String())
			} else {
				fmt.Printf("%s Skipped - set goals later with: ung goal set <amount>\n", mutedStyle.Render(""))
			}
		} else {
			fmt.Printf("%s You already have a goal: $%.0f for %s\n", successStyle.Render(""), existingGoal.Amount, time.Now().Month().String())
		}
	} else {
		fmt.Printf("%s Skipped (--skip-goal)\n", mutedStyle.Render(""))
	}

	// ============ SUCCESS! ============
	fmt.Println()

	// Clean celebration (accessible - no decorative emoji spam)
	celebrationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22C55E")).
		Bold(true)

	// Focused success message with single clear next action
	successBox := boxStyle.Render(fmt.Sprintf(`%s

  Your freelance toolkit is ready!

%s

  ung track start "Project name"

  This will start a billable timer. Stop anytime with: ung track stop

%s
  ung next     Your personalized dashboard
  ung help     Explore all features`,
		celebrationStyle.Render("  Setup Complete!"),
		headerStyle.Render("Your Next Step:"),
		mutedStyle.Render("When you're ready:")))

	fmt.Println(successBox)
	fmt.Println()

	return nil
}
