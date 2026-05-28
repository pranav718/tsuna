package cmd

import (
	"fmt"

	"github.com/pranav718/tsuna/internal/signal"
	"github.com/spf13/cobra"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "run the signaling server",
	Long:  "start the lightweight signaling server for room code to peer address resolution.",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		addr := fmt.Sprintf(":%d", port)

		fmt.Printf("starting signaling server on %s\n", addr)
		srv := signal.NewServer(addr)
		return srv.Start()
	},
}

func init() {
	signalCmd.Flags().IntP("port", "p", 8080, "port to listen on")
	rootCmd.AddCommand(signalCmd)
}
