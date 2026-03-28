package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Create a new watch room",
	Long:  "Spin up a new Tsuna room and get a 6-character code to share with your friends.",
	RunE:  runHost,
}

func init() {
	rootCmd.AddCommand(hostCmd)
}

func runHost(cmd *cobra.Command, args []string) error {
	fmt.Println("hosting...")
	return nil
}
