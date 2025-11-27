package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/pkg/template"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage PDF templates",
	Long: `Manage PDF templates for invoices and contracts.

Templates define the layout, colors, and blocks used when generating PDFs.
You can create custom templates or use the built-in default template.

Examples:
  ung template list                    List available templates
  ung template show default            Show template configuration
  ung template create my-template      Create a new template from default
  ung template preview my-template     Generate a preview PDF
  ung template use my-template         Set as active template`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long:  `List all available invoice and contract templates.`,
	RunE:  runTemplateList,
}

var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template configuration",
	Long:  `Display the configuration of a template in JSON format.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new template",
	Long: `Create a new template based on the default template.

The template will be saved to the templates directory and can be customized.`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateCreate,
}

var templatePreviewCmd = &cobra.Command{
	Use:   "preview [name]",
	Short: "Generate a template preview PDF",
	Long: `Generate a preview PDF using sample data to see how the template looks.

If no name is provided, uses the default template.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTemplatePreview,
}

var templateUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set active template",
	Long:  `Set a template as the active template for generating PDFs.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateUse,
}

var (
	templateType   string
	templateOutput string
)

func init() {
	rootCmd.AddCommand(templateCmd)

	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templatePreviewCmd)
	templateCmd.AddCommand(templateUseCmd)

	// Flags for create
	templateCreateCmd.Flags().StringVarP(&templateType, "type", "t", "invoice", "Template type (invoice or contract)")

	// Flags for preview
	templatePreviewCmd.Flags().StringVarP(&templateOutput, "output", "o", "", "Output file path")
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	templatesDir := getTemplatesDir(cfg)
	templates, err := template.ListTemplates(templatesDir)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	fmt.Println("Available Templates:")
	fmt.Println("====================")
	fmt.Println()

	for _, t := range templates {
		status := ""
		if t.IsBuiltin {
			status = " (built-in)"
		} else if t.IsCustom {
			status = " (custom)"
		}

		fmt.Printf("  %s%s\n", t.Name, status)
		fmt.Printf("    Type: %s\n", t.Type)
		fmt.Printf("    %s\n", t.Description)
		if t.Path != "" {
			fmt.Printf("    Path: %s\n", t.Path)
		}
		fmt.Println()
	}

	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	var tmpl *template.TemplateDefinition

	if name == "default" {
		tmpl = template.GetDefaultInvoiceTemplate()
	} else {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		templatesDir := getTemplatesDir(cfg)
		templatePath := filepath.Join(templatesDir, name+".json")

		tmpl, err = template.LoadTemplate(templatePath)
		if err != nil {
			return fmt.Errorf("template '%s' not found: %w", name, err)
		}
	}

	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format template: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	templatesDir := getTemplatesDir(cfg)
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	templatePath := filepath.Join(templatesDir, name+".json")

	// Check if template already exists
	if _, err := os.Stat(templatePath); err == nil {
		return fmt.Errorf("template '%s' already exists at %s", name, templatePath)
	}

	// Create from default
	tmpl := template.GetDefaultInvoiceTemplate()
	tmpl.Name = name
	tmpl.Description = fmt.Sprintf("Custom %s template", templateType)

	if err := template.SaveTemplate(tmpl, templatePath); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	fmt.Printf("✓ Created template '%s' at %s\n", name, templatePath)
	fmt.Println()
	fmt.Println("Edit the JSON file to customize your template, then use:")
	fmt.Printf("  ung template preview %s    # Preview your changes\n", name)
	fmt.Printf("  ung template use %s        # Set as active template\n", name)

	return nil
}

func runTemplatePreview(cmd *cobra.Command, args []string) error {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var tmpl *template.TemplateDefinition

	if name == "default" {
		tmpl = template.GetDefaultInvoiceTemplate()
	} else {
		templatesDir := getTemplatesDir(cfg)
		templatePath := filepath.Join(templatesDir, name+".json")

		tmpl, err = template.LoadTemplate(templatePath)
		if err != nil {
			return fmt.Errorf("template '%s' not found: %w", name, err)
		}
	}

	// Generate preview
	outputPath := templateOutput
	if outputPath == "" {
		outputPath = filepath.Join(os.TempDir(), fmt.Sprintf("ung-template-preview-%s.pdf", name))
	}

	renderer := template.NewRenderer(tmpl)
	sampleData := template.GetSampleInvoiceData(cfg)

	if err := renderer.RenderInvoice(sampleData, outputPath); err != nil {
		return fmt.Errorf("failed to generate preview: %w", err)
	}

	fmt.Printf("✓ Preview generated: %s\n", outputPath)
	return nil
}

func runTemplateUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if name == "default" {
		// Clear custom template path
		cfg.Templates.InvoiceHTML = ""
		fmt.Println("✓ Using default built-in template")
	} else {
		templatesDir := getTemplatesDir(cfg)
		templatePath := filepath.Join(templatesDir, name+".json")

		// Verify template exists
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("template '%s' not found at %s", name, templatePath)
		}

		cfg.Templates.InvoiceHTML = templatePath
		fmt.Printf("✓ Active template set to '%s'\n", name)
	}

	// Save config
	isGlobal := config.GetConfigSource() == config.SourceGlobal
	if err := config.Save(cfg, isGlobal); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func getTemplatesDir(cfg *config.Config) string {
	// Use templates directory alongside config
	switch config.GetConfigSource() {
	case config.SourceLocal:
		return filepath.Join(config.GetLocalUngDir(), "templates")
	case config.SourceGlobal:
		return filepath.Join(config.GetGlobalUngDir(), "templates")
	default:
		return filepath.Join(config.GetLocalUngDir(), "templates")
	}
}
