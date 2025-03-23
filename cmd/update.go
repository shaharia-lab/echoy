package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"os"
	"strings"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

// NewUpdateCmd creates a new update command
func NewUpdateCmd(appCfg *config.AppConfig) *cobra.Command {
	updateCmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "update",
		Short:   "Check for updates and update the CLI",
		Long:    "Check for updates and if a new version is available, download and install it",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(appCfg.Version.Version)
		},
	}

	return updateCmd
}

func runUpdate(currentAppVersion string) error {
	// Set the GitHub repository
	repoSlug := "shaharia-lab/echoy"

	fmt.Println("Checking for updates...")

	// Check for the latest version
	latest, found, err := selfupdate.DetectLatest(repoSlug)
	if err != nil {
		return fmt.Errorf("error detecting version: %s", err)
	}

	// Remove 'v' prefix for comparison if needed
	currentVersionNoV := strings.TrimPrefix(currentAppVersion, "v")
	latestVersionNoV := strings.TrimPrefix(latest.Version.String(), "v")

	// Check if we're already on the latest version
	if !found || latestVersionNoV == currentVersionNoV {
		fmt.Printf("Current version (%s) is the latest\n", currentAppVersion)
		return nil
	}

	// Confirm with the user
	fmt.Printf("New version available: %s (current: %s)\n", latest.Version, currentAppVersion)
	fmt.Printf("Release notes:\n%s\n", latest.ReleaseNotes)
	fmt.Print("Do you want to update? (y/n): ")

	var input string
	fmt.Scanln(&input)
	if strings.ToLower(input) != "y" && strings.ToLower(input) != "yes" {
		fmt.Println("Update cancelled")
		return nil
	}

	// Apply the update
	fmt.Println("Downloading and installing update...")
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %s", err)
	}

	// Update the binary
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("error updating binary: %s", err)
	}

	fmt.Printf("Successfully updated to version %s\n", latest.Version)
	return nil
}
