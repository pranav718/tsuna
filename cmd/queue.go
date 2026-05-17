package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "manage the watch queue",
	Long:  "add, remove, and list episodes in the shared watch queue.",
}

var queueAddCmd = &cobra.Command{
	Use:   "add <file>",
	Short: "add a file to the watch queue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  queued: %s\n", args[0])
		fmt.Println(dim.Render("  (file will be added when connected to a room)"))
		return nil
	},
}

var queueListCmd = &cobra.Command{
	Use:   "list",
	Short: "show the current watch queue",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(dim.Render("  queue is empty (join a room first)"))
		return nil
	},
}

func init() {
	queueCmd.AddCommand(queueAddCmd)
	queueCmd.AddCommand(queueListCmd)
	rootCmd.AddCommand(queueCmd)
}
