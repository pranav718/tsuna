package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show sync health and connected peers",
	Long:  "display real-time sync status, peer latencies, and playback deltas for the current room.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("tsuna status: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
