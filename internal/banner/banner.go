// Package banner contains the design of the CLI banner.
package banner

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
)

// Banner is a struct that contains the methods to display the application banner
type Banner struct {
	appConfig *config.AppConfig
}

func CLIBanner(appCfg *config.AppConfig) *Banner {
	return &Banner{
		appConfig: appCfg,
	}
}

// Display prints the application banner
func (b *Banner) Display() {
	bannerWidth := 40
	padding := (bannerWidth - len(b.appConfig.Name) - 10) / 2
	spaces := ""
	for i := 0; i < padding; i++ {
		spaces += " "
	}

	color.New(color.FgHiCyan, color.Bold).Println("╔════════════════════════════════════════╗")
	color.New(color.FgHiCyan, color.Bold).Println(fmt.Sprintf("║%sWelcome to %s%s║", spaces, b.appConfig.Name, spaces))
	color.New(color.FgHiCyan, color.Bold).Println("╚════════════════════════════════════════╝")
}
