package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// These variables are set at build time using ldflags
var (
	Version   = "1.0.18"
	GitCommit = "dev"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display detailed version information including version number, git commit, build date, and Go version.`,
	Run:   runVersion,
}

var versionBumpCmd = &cobra.Command{
	Use:   "bump [patch|minor|major]",
	Short: "Bump version and create git tag",
	Long: `Bump the version number and create a git tag.

Examples:
  ung version bump patch  # 1.2.3 -> 1.2.4
  ung version bump minor  # 1.2.3 -> 1.3.0
  ung version bump major  # 1.2.3 -> 2.0.0

This command will:
1. Get the latest git tag
2. Increment the version number
3. Create a new git tag
4. Display the new version`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"patch", "minor", "major"},
	Run: runVersionBump,
}

func init() {
	versionCmd.AddCommand(versionBumpCmd)
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("🐾 UNG - Universal Next-Gen Billing & Tracking\n\n")
	fmt.Printf("Version:    %s\n", Version)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Build Date: %s\n", BuildDate)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func runVersionBump(cmd *cobra.Command, args []string) {
	bumpType := args[0]

	// Check if git is available
	if !isGitAvailable() {
		fmt.Println("❌ Git is not available. Please install git to use version bump.")
		return
	}

	// Check if we're in a git repository
	if !isGitRepo() {
		fmt.Println("❌ Not a git repository. Initialize git first with: git init")
		return
	}

	// Get the latest tag
	latestTag := getLatestGitTag()
	fmt.Printf("Current version: %s\n", latestTag)

	// Parse and bump version
	newVersion, err := bumpVersion(latestTag, bumpType)
	if err != nil {
		fmt.Printf("❌ Failed to bump version: %v\n", err)
		return
	}

	fmt.Printf("New version:     %s\n\n", newVersion)

	// Ask for confirmation
	fmt.Print("❓ Create git tag? (y/N): ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		fmt.Println("Cancelled")
		return
	}

	// Create git tag
	if err := createGitTag(newVersion); err != nil {
		fmt.Printf("❌ Failed to create tag: %v\n", err)
		return
	}

	fmt.Printf("✅ Created tag: %s\n", newVersion)
	fmt.Println("\n💡 To push the tag to remote:")
	fmt.Printf("   git push origin %s\n", newVersion)
}

func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func getLatestGitTag() string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "v0.0.0"
	}
	return strings.TrimSpace(string(output))
}

func bumpVersion(currentVersion, bumpType string) (string, error) {
	// Remove 'v' prefix if present
	version := strings.TrimPrefix(currentVersion, "v")

	// Parse version using regex
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) != 4 {
		return "", fmt.Errorf("invalid version format: %s", currentVersion)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	// Bump the appropriate component
	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("invalid bump type: %s (must be patch, minor, or major)", bumpType)
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}

func createGitTag(tag string) error {
	// Create annotated tag
	cmd := exec.Command("git", "tag", "-a", tag, "-m", fmt.Sprintf("Release %s", tag))
	return cmd.Run()
}
