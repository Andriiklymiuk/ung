package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/pkg/email"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Manage email configuration",
	Long: `Configure SMTP settings for sending invoices and contracts via email.

Examples:
  ung email setup       # Interactive SMTP configuration
  ung email test        # Test email connection
  ung email show        # Show current email config`,
}

var emailSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure SMTP settings interactively",
	Long:  `Set up SMTP email configuration through an interactive form.`,
	Run:   runEmailSetup,
}

var emailTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test email configuration",
	Long:  `Test the SMTP connection with current configuration.`,
	Run:   runEmailTest,
}

var emailShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show email configuration",
	Long:  `Display the current email configuration (password hidden).`,
	Run:   runEmailShow,
}

var emailSendTestCmd = &cobra.Command{
	Use:   "send-test [recipient-email]",
	Short: "Send a test email",
	Long:  `Send a test email to verify the configuration is working.`,
	Args:  cobra.ExactArgs(1),
	Run:   runEmailSendTest,
}

func init() {
	emailCmd.AddCommand(emailSetupCmd)
	emailCmd.AddCommand(emailTestCmd)
	emailCmd.AddCommand(emailShowCmd)
	emailCmd.AddCommand(emailSendTestCmd)
	rootCmd.AddCommand(emailCmd)
}

func runEmailSetup(cmd *cobra.Command, args []string) {
	fmt.Println("📧 Email Configuration Setup\n")

	cfg, _ := config.Load()

	// Interactive form for email config
	var (
		smtpHost    = cfg.Email.SMTPHost
		smtpPortInt = cfg.Email.SMTPPort
		username    = cfg.Email.Username
		password    = cfg.Email.Password
		fromEmail   = cfg.Email.FromEmail
		fromName    = cfg.Email.FromName
		useTLS      = cfg.Email.UseTLS
	)

	// Set defaults if empty
	if smtpPortInt == 0 {
		smtpPortInt = 587
	}
	if useTLS == false && smtpPortInt == 587 {
		useTLS = true
	}

	// Convert port to string for form input
	smtpPort := fmt.Sprintf("%d", smtpPortInt)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("SMTP Host").
				Description("e.g., smtp.gmail.com, smtp.office365.com").
				Value(&smtpHost).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("SMTP host is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("SMTP Port").
				Description("Common: 587 (TLS), 465 (SSL), 25 (no encryption)").
				Value(&smtpPort).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("port is required")
					}
					port, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("port must be a number")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("port must be between 1 and 65535")
					}
					return nil
				}),

			huh.NewInput().
				Title("Username").
				Description("Your email address or SMTP username").
				Value(&username).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("username is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Password").
				Description("Use app-specific password for Gmail/Outlook").
				Value(&password).
				Password(true).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("password is required")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("From Email").
				Description("Email address to send from").
				Value(&fromEmail).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("from email is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("From Name").
				Description("Display name for sender").
				Value(&fromName),

			huh.NewConfirm().
				Title("Use TLS?").
				Description("Enable TLS encryption (recommended)").
				Value(&useTLS),
		),
	)

	if err := form.Run(); err != nil {
		fmt.Printf("❌ Configuration cancelled: %v\n", err)
		return
	}

	// Convert port string to int
	portInt, _ := strconv.Atoi(smtpPort) // Already validated

	// Update config
	cfg.Email = config.EmailConfig{
		SMTPHost:  smtpHost,
		SMTPPort:  portInt,
		Username:  username,
		Password:  password,
		FromEmail: fromEmail,
		FromName:  fromName,
		UseTLS:    useTLS,
	}

	// Ask if global or local
	var saveGlobal bool
	saveForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Save as global configuration?").
				Description("No = save to local workspace (.ung.yaml)").
				Value(&saveGlobal),
		),
	)

	if err := saveForm.Run(); err == nil {
		if err := config.Save(cfg, saveGlobal); err != nil {
			fmt.Printf("❌ Failed to save configuration: %v\n", err)
			return
		}

		if saveGlobal {
			home, _ := os.UserHomeDir()
			fmt.Printf("✅ Email configuration saved to ~/.ung/config.yaml\n")
			fmt.Printf("   Config: %s/.ung/config.yaml\n", home)
		} else {
			fmt.Println("✅ Email configuration saved to .ung.yaml")
			fmt.Println("   Config: ./.ung.yaml")
		}

		fmt.Println("\n💡 Test your configuration:")
		fmt.Println("   ung email test")
	}
}

func runEmailTest(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Testing email configuration...\n")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	if cfg.Email.SMTPHost == "" {
		fmt.Println("❌ Email not configured")
		fmt.Println("\n💡 Set up email first:")
		fmt.Println("   ung email setup")
		return
	}

	// Convert to email config
	emailCfg := &email.Config{
		SMTPHost:  cfg.Email.SMTPHost,
		SMTPPort:  cfg.Email.SMTPPort,
		Username:  cfg.Email.Username,
		Password:  cfg.Email.Password,
		FromEmail: cfg.Email.FromEmail,
		FromName:  cfg.Email.FromName,
		UseTLS:    cfg.Email.UseTLS,
	}

	// Validate config
	if err := email.ValidateConfig(emailCfg); err != nil {
		fmt.Printf("❌ Invalid configuration: %v\n", err)
		return
	}

	fmt.Printf("Connecting to %s:%d...\n", emailCfg.SMTPHost, emailCfg.SMTPPort)

	// Test connection
	if err := email.TestConnection(emailCfg); err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		fmt.Println("\n💡 Common issues:")
		fmt.Println("   - Wrong password (use app-specific password for Gmail/Outlook)")
		fmt.Println("   - SMTP host or port incorrect")
		fmt.Println("   - TLS settings mismatch")
		fmt.Println("   - Account security settings blocking access")
		return
	}

	fmt.Println("✅ Connection successful!")
	fmt.Println("\n💡 Send a test email:")
	fmt.Println("   ung email send-test recipient@example.com")
}

func runEmailShow(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	if cfg.Email.SMTPHost == "" {
		fmt.Println("❌ Email not configured")
		fmt.Println("\n💡 Set up email first:")
		fmt.Println("   ung email setup")
		return
	}

	fmt.Println("📧 Email Configuration\n")
	fmt.Printf("SMTP Host:  %s\n", cfg.Email.SMTPHost)
	fmt.Printf("SMTP Port:  %d\n", cfg.Email.SMTPPort)
	fmt.Printf("Username:   %s\n", cfg.Email.Username)
	fmt.Printf("Password:   %s\n", maskPassword(cfg.Email.Password))
	fmt.Printf("From Email: %s\n", cfg.Email.FromEmail)
	fmt.Printf("From Name:  %s\n", cfg.Email.FromName)
	fmt.Printf("Use TLS:    %v\n", cfg.Email.UseTLS)

	// Show config source
	configSource := "default"
	if _, err := os.Stat(".ung.yaml"); err == nil {
		configSource = ".ung.yaml (local workspace)"
	} else {
		home, _ := os.UserHomeDir()
		globalConfig := home + "/.ung/config.yaml"
		if _, err := os.Stat(globalConfig); err == nil {
			configSource = globalConfig + " (global)"
		}
	}
	fmt.Printf("\nSource:     %s\n", configSource)
}

func runEmailSendTest(cmd *cobra.Command, args []string) {
	recipient := args[0]

	fmt.Printf("📧 Sending test email to %s...\n\n", recipient)

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	if cfg.Email.SMTPHost == "" {
		fmt.Println("❌ Email not configured")
		return
	}

	// Convert to email config
	emailCfg := &email.Config{
		SMTPHost:  cfg.Email.SMTPHost,
		SMTPPort:  cfg.Email.SMTPPort,
		Username:  cfg.Email.Username,
		Password:  cfg.Email.Password,
		FromEmail: cfg.Email.FromEmail,
		FromName:  cfg.Email.FromName,
		UseTLS:    cfg.Email.UseTLS,
	}

	// Create test email
	testEmail := &email.Email{
		To:      []string{recipient},
		Subject: "Test Email from UNG",
		Body: `This is a test email from UNG - Universal Next-Gen Billing & Tracking.

If you received this email, your SMTP configuration is working correctly!

Best regards,
UNG CLI`,
		HTMLBody: `<html>
<body style="font-family: Arial, sans-serif;">
<h2>Test Email from UNG</h2>
<p>This is a test email from <strong>UNG - Universal Next-Gen Billing & Tracking</strong>.</p>
<p>If you received this email, your SMTP configuration is working correctly!</p>
<br>
<p>Best regards,<br><strong>UNG CLI</strong></p>
</body>
</html>`,
	}

	// Send email
	if err := email.Send(emailCfg, testEmail); err != nil {
		fmt.Printf("❌ Failed to send email: %v\n", err)
		return
	}

	fmt.Println("✅ Test email sent successfully!")
	fmt.Printf("   Check %s inbox\n", recipient)
}

func maskPassword(password string) string {
	if password == "" {
		return "(not set)"
	}
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}
