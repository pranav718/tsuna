package cmd

import (
	"os"

	"github.com/pranav718/tsuna/internal/config"
	"github.com/spf13/cobra"
)

var (
	signalServer string
	mpvSocket    string
	localMode    bool
)

var rootCmd = &cobra.Command{
	Use:   "tsuna",
	Short: "p2p synchronized video watching",
	Long: `
 ████████╗███████╗██╗   ██╗███╗   ██╗ █████╗
    ██╔══╝██╔════╝██║   ██║████╗  ██║██╔══██╗
    ██║   ███████╗██║   ██║██╔██╗ ██║███████║
    ██║   ╚════██║██║   ██║██║╚██╗██║██╔══██║
    ██║   ███████║╚██████╔╝██║ ╚████║██║  ██║
    ╚═╝   ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═╝

  watch anime together. no servers. no accounts.
  just a 6-char room code and a UDP packet.`,
}

func init() {
	cfg := config.Load()

	rootCmd.PersistentFlags().StringVar(&signalServer, "signal-server", cfg.SignalServer, "signaling server URL")
	rootCmd.PersistentFlags().StringVar(&mpvSocket, "mpv-socket", cfg.MpvSocket, "mpv IPC socket path")
	rootCmd.PersistentFlags().BoolVar(&localMode, "local", cfg.LocalMode, "use loopback for same-machine testing (skip STUN)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
