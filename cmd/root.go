package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pranav718/tsuna/internal/config"
	"github.com/spf13/cobra"
)

var (
	signalServer string
	mpvSocket    string
	localMode    bool
	noBrowser    bool
)

var asciiLines = []string{
	` ████████╗███████╗██╗   ██╗███╗   ██╗ █████╗ `,
	`    ██╔══╝██╔════╝██║   ██║████╗  ██║██╔══██╗`,
	`    ██║   ███████╗██║   ██║██╔██╗ ██║███████║`,
	`    ██║   ╚════██║██║   ██║██║╚██╗██║██╔══██║`,
	`    ██║   ███████║╚██████╔╝██║ ╚████║██║  ██║`,
	`    ╚═╝   ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═╝`,
}

var bannerGradient = []string{
	"#ffffff",
	"#f3e8ff",
	"#e9d8fd",
	"#d8b4fe",
	"#c084fc",
	"#a855f7",
}

func renderGradientBanner() string {
	lines := make([]string, len(asciiLines))
	for i, line := range asciiLines {
		color := bannerGradient[i%len(bannerGradient)]
		lines[i] = lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Bold(true).
			Render(line)
	}
	return strings.Join(lines, "\n")
}

var rootCmd = &cobra.Command{
	Use:   "tsuna",
	Short: "p2p synchronized video watching",
}

func init() {
	cfg := config.Load()

	rootCmd.Long = fmt.Sprintf("\n%s\n\n%s\n%s",
		renderGradientBanner(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render("  watch anything together. no servers. no accounts."),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#76778f")).Render("  just a 6-char room code and a udp packet."),
	)

	rootCmd.SetUsageTemplate(`usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

available commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

global flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	for _, cmd := range rootCmd.Commands() {
		switch cmd.Name() {
		case "help":
			cmd.Short = "help about any command"
		case "completion":
			cmd.Short = "generate the autocompletion script for the specified shell"
		}
	}

	rootCmd.PersistentFlags().StringVar(&signalServer, "signal-server", cfg.SignalServer, "signaling server url")
	rootCmd.PersistentFlags().StringVar(&mpvSocket, "mpv-socket", cfg.MpvSocket, "mpv ipc socket path")
	rootCmd.PersistentFlags().BoolVar(&localMode, "local", cfg.LocalMode, "use loopback for same-machine testing (skip stun)")
	rootCmd.PersistentFlags().BoolVar(&noBrowser, "no-browser", false, "disable auto-opening web dashboard")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
