package init

import (
	"github.com/AlecAivazis/survey/v2"
)

// ConfigureUser configures user information
func (i *Initializer) ConfigureUser() error {
	var userName string
	promptUserName := &survey.Input{
		Message: "Your name (optional):",
		Help:    "Your name will be used in conversations",
		Default: i.Config.User.Name,
	}
	err := survey.AskOne(promptUserName, &userName)
	if err != nil {
		return err
	}
	i.Config.User.Name = userName

	return nil
}
