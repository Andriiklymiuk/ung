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

	// Post-process files to fix MDX compatibility issues
	err = postProcessDocs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to post-process docs: %w", err)
	}

	fmt.Printf("✓ Documentation generated in %s\n", outputDir)
	return nil
}

// postProcessDocs fixes MDX compatibility issues in generated markdown
// Converts tab-indented code to proper fenced code blocks
func postProcessDocs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		newContent := convertTabIndentedCode(string(content))

		// Only write if content changed
		if newContent != string(content) {
			if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// convertTabIndentedCode converts tab-indented code blocks to fenced code blocks
func convertTabIndentedCode(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var codeBlockLines []string
	inFencedBlock := false
	inTabBlock := false

	for _, line := range lines {
		// Track fenced code blocks
		if strings.HasPrefix(line, "```") {
			inFencedBlock = !inFencedBlock
			if inTabBlock {
				// Close tab block before fenced block
				result = append(result, "```bash")
				result = append(result, codeBlockLines...)
				result = append(result, "```")
				inTabBlock = false
				codeBlockLines = nil
			}
			result = append(result, line)
			continue
		}

		// Skip processing inside fenced blocks
		if inFencedBlock {
			result = append(result, line)
			continue
		}

		// Check for tab-indented lines (single tab, not double)
		isTabIndented := strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, "\t\t")

		if isTabIndented {
			if !inTabBlock {
				inTabBlock = true
				codeBlockLines = []string{}
			}
			// Remove leading tab
			codeBlockLines = append(codeBlockLines, strings.TrimPrefix(line, "\t"))
		} else {
			if inTabBlock {
				// Close the code block
				result = append(result, "```bash")
				result = append(result, codeBlockLines...)
				result = append(result, "```")
				inTabBlock = false
				codeBlockLines = nil
			}
			result = append(result, line)
		}
	}

	// Handle case where file ends with tab block
	if inTabBlock {
		result = append(result, "```bash")
		result = append(result, codeBlockLines...)
		result = append(result, "```")
	}

	return strings.Join(result, "\n")
}
