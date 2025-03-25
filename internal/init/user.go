package init

import (
	"github.com/AlecAivazis/survey/v2"
)

// ConfigureUser configures user information
func (i *Initializer) ConfigureUser() error {
	i.cliTheme.Primary().Println("\n📝 Your Information")

	var userName string
	promptUserName := &survey.Input{
		Message: "Name (optional):",
		Help:    "Your name will be used in conversations",
		Default: i.Config.User.Name,
	}
	err := survey.AskOne(promptUserName, &userName)
	if err != nil {
		return err
	}

	if userName == "" {
		userName = "User"
	}

	i.Config.User.Name = userName
	return nil
}
