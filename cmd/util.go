package cmd

import (
	"fmt"
	"github.com/fatih/color"
)

// PrintColorfulBanner prints a colorful banner
func PrintColorfulBanner() {
	banner := `
 _____  _                         _    ___ 
|  __ \| |                       /_\  |_ _|
| |__) | | ___  __ _ ___  ___   //_\\ | | 
|  ___/| |/ _ \/ _' / __|/ _ \ / _ \ | | 
| |    | |  __/ (_| \__ \  __/| (_) || | 
|_|    |_|\___|\__,_|___/\___| \___/|___|
`
	color.New(color.FgCyan, color.Bold).Println(banner)
	color.New(color.FgHiYellow).Println("Your AI assistant for the command line!")
	fmt.Println()
}
