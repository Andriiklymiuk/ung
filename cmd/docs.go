package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation",
	Long:  "Generate markdown documentation for all CLI commands",
}

var docsGenerateCmd = &cobra.Command{
	Use:   "generate [output-dir]",
	Short: "Generate markdown documentation for all CLI commands",
	Long:  "Generate markdown documentation with Docusaurus frontmatter for all CLI commands",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDocsGenerate,
}

func init() {
	docsCmd.AddCommand(docsGenerateCmd)
}

func runDocsGenerate(cmd *cobra.Command, args []string) error {
	outputDir := "./ung-docs/docs/cli"
	if len(args) > 0 {
		outputDir = args[0]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Docusaurus frontmatter prepender
	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		name = strings.TrimSuffix(name, ".md")
		title := strings.ReplaceAll(name, "_", " ")
		return fmt.Sprintf("---\nid: %s\ntitle: %s\n---\n\n", name, title)
	}

	// Link handler for Docusaurus
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, ".md")
		return "./" + base
	}

	// Disable auto-generation header to keep docs clean
	rootCmd.DisableAutoGenTag = true

	err := doc.GenMarkdownTreeCustom(rootCmd, outputDir, filePrepender, linkHandler)
	if err != nil {
		return fmt.Errorf("failed to generate docs: %w", err)
	}

	fmt.Printf("✓ Documentation generated in %s\n", outputDir)
	return nil
}
