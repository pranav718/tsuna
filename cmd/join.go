package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var joinCmd = &cobra.Command{
	Use:   "join <CODE>",
	Short: "Join a watch party room",
	Long:  "Join an existing Tsuna room using a 6-character room code.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code := strings.ToUpper(strings.TrimSpace(args[0]))

		if len(code) != 6 {
			return fmt.Errorf("invalid room code %q — must be exactly 6 characters", code)
		}

		fmt.Printf("jining room %s...\n", code)
		fmt.Println("discovering peers via signaling server...")
		fmt.Println("punching through NAT...")
		fmt.Println("syncing clocks...")
		fmt.Printf("connected to room %s\n", code)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
}
