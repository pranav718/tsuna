package cmd

import (
	"os"

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
	rootCmd.PersistentFlags().StringVar(&signalServer, "signal-server", "http://localhost:8080", "signaling server URL")
	rootCmd.PersistentFlags().StringVar(&mpvSocket, "mpv-socket", "/tmp/tsuna-mpv.sock", "mpv IPC socket path")
	rootCmd.PersistentFlags().BoolVar(&localMode, "local", false, "use loopback for same-machine testing (skip STUN)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
