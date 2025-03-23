package init

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// ConfigureTools configures the available tools
func (i *Initializer) ConfigureTools() {
	color.New(color.FgHiCyan, color.Bold).Println("\n🧰 Configure Tools")
	color.New(color.FgHiWhite).Println("Select which tools you want to enable for your assistant")

	// Define tool options and their mapping to config properties
	type toolDefinition struct {
		displayName   string
		enabledPtr    *bool
		configureFunc func()
	}

	toolDefinitions := []toolDefinition{
		{
			displayName:   "Docker - LLM can interact with your local docker",
			enabledPtr:    &i.Config.Tools.Docker.Enabled,
			configureFunc: nil,
		},
		{
			displayName:   "Git - Assistant will use your local Git to perform version control operation",
			enabledPtr:    &i.Config.Tools.Git.Enabled,
			configureFunc: i.configureGitSettings,
		},
		{
			displayName:   "Bash - Assistant can execute bash commands",
			enabledPtr:    &i.Config.Tools.Bash.Enabled,
			configureFunc: nil,
		},
		{
			displayName:   "Sed - Assistant can execute sed commands",
			enabledPtr:    &i.Config.Tools.Sed.Enabled,
			configureFunc: nil,
		},
		{
			displayName:   "Grep - Assistant can execute grep commands",
			enabledPtr:    &i.Config.Tools.Grep.Enabled,
			configureFunc: nil,
		},
		{
			displayName:   "Cat - Assistant can execute cat commands",
			enabledPtr:    &i.Config.Tools.Cat.Enabled,
			configureFunc: nil,
		},
	}

	// Build options and defaults arrays
	toolOptions := make([]string, len(toolDefinitions))
	defaultSelections := []string{}

	for j, tool := range toolDefinitions {
		toolOptions[j] = tool.displayName
		if *tool.enabledPtr {
			defaultSelections = append(defaultSelections, tool.displayName)
		}
	}

	// Ask for tool selection
	var selectedTools []string
	promptTools := &survey.MultiSelect{
		Message: "Which tools do you want to enable?",
		Options: toolOptions,
		Default: defaultSelections,
		Help:    "Select tools using space, navigate with arrow keys",
	}
	survey.AskOne(promptTools, &selectedTools)

	// Reset all tool enabled flags
	for _, tool := range toolDefinitions {
		*tool.enabledPtr = false
	}

	// Set the selected tools to enabled and call their configure function if needed
	for _, selected := range selectedTools {
		for _, tool := range toolDefinitions {
			if selected == tool.displayName {
				*tool.enabledPtr = true
				if tool.configureFunc != nil {
					tool.configureFunc()
				}
				break
			}
		}
	}
}

// configureGitSettings configures Git-specific settings
func (i *Initializer) configureGitSettings() {
	// Handle whitelisted repositories
	i.configureGitWhitelistedRepos()

	// Handle blocked operations
	i.configureGitBlockedOperations()
}

// configureGitWhitelistedRepos configures Git whitelisted repositories
func (i *Initializer) configureGitWhitelistedRepos() {
	if i.IsUpdateMode && len(i.Config.Tools.Git.WhitelistedRepoPaths) > 0 {
		color.New(color.FgHiYellow).Println("\nCurrent whitelisted Git repositories:")
		for j, path := range i.Config.Tools.Git.WhitelistedRepoPaths {
			fmt.Printf("%d. %s\n", j+1, path)
		}

		var modifyRepos bool
		promptModifyRepos := &survey.Confirm{
			Message: "Do you want to modify the whitelisted repositories?",
			Default: false,
		}
		survey.AskOne(promptModifyRepos, &modifyRepos)

		if !modifyRepos {
			return // Keep existing repositories
		}
		// Clear existing repos to recollect them
		i.Config.Tools.Git.WhitelistedRepoPaths = []string{}
	}

	// Ask if they want to whitelist repositories
	var whitelistRepo bool
	promptWhitelist := &survey.Confirm{
		Message: "Do you want to whitelist any specific repository directory in your local machine?",
		Default: false,
	}
	survey.AskOne(promptWhitelist, &whitelistRepo)

	if whitelistRepo {
		i.collectGitRepoPaths()
	}
}

// collectGitRepoPaths collects Git repository paths
func (i *Initializer) collectGitRepoPaths() {
	addMoreRepos := true

	for addMoreRepos {
		var repoPath string
		promptRepoPath := &survey.Input{
			Message: "Provide full path of repository directory:",
			Help:    "Example: /home/username/projects/repo",
		}
		survey.AskOne(promptRepoPath, &repoPath)

		if repoPath != "" {
			i.Config.Tools.Git.WhitelistedRepoPaths = append(i.Config.Tools.Git.WhitelistedRepoPaths, repoPath)
		}

		promptAddMore := &survey.Confirm{
			Message: "Do you want to add more whitelisted repository directories?",
			Default: false,
		}
		survey.AskOne(promptAddMore, &addMoreRepos)
	}
}

// configureGitBlockedOperations configures Git blocked operations
func (i *Initializer) configureGitBlockedOperations() {
	// Similar to whitelisted repos, but for blocked operations
	// Implementation would go here, following the same pattern
}
