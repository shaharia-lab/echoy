package init

import (
	"github.com/AlecAivazis/survey/v2"
)

// ConfigureUser configures user information
func (i *Initializer) ConfigureUser() {
	var userName string
	promptUserName := &survey.Input{
		Message: "What's your name (optional):",
		Help:    "Your name will be used in conversations",
		Default: i.Config.User.Name,
	}
	survey.AskOne(promptUserName, &userName)
	i.Config.User.Name = userName

	var userEmail string
	promptUserEmail := &survey.Input{
		Message: "Your email address (optional):",
		Help:    "Your email may be used for notifications",
		Default: i.Config.User.Email,
	}
	survey.AskOne(promptUserEmail, &userEmail)
	i.Config.User.Email = userEmail
}
