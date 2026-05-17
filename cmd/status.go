package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	sig "github.com/pranav718/tsuna/internal/signal"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [CODE]",
	Short: "show sync health and connected peers",
	Long:  "display active rooms and peers on the signaling server. pass a room code to see details.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		return showRoom(args[0])
	}
	return showAllRooms()
}

func showAllRooms() error {
	rooms, err := sig.GetRooms(signalServer)
	if err != nil {
		printError(fmt.Sprintf("cannot reach signal server at %s", signalServer))
		return err
	}

	if len(rooms) == 0 {
		fmt.Println()
		fmt.Printf("  %s\n\n", dim.Render("no active rooms"))
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)

	fmt.Println()
	fmt.Printf("  %s  %s\n\n", headerStyle.Render("ACTIVE ROOMS"), dim.Render(fmt.Sprintf("(%d)", len(rooms))))

	for _, r := range rooms {
		age := time.Since(r.CreatedAt).Round(time.Second)
		code := pink.Render(r.Code)
		peers := green.Render(fmt.Sprintf("%d peers", r.PeerCount))
		uptime := dim.Render(fmt.Sprintf("up %s", age))

		fmt.Printf("  %s  %s  %s\n", code, peers, uptime)

		for _, p := range r.Peers {
			joined := time.Since(p.JoinedAt).Round(time.Second)
			fmt.Printf("    %s %s  %s\n",
				green.Render("*"),
				p.PeerID,
				dim.Render(fmt.Sprintf("joined %s ago", joined)),
			)
		}
		fmt.Println()
	}

	return nil
}

func showRoom(code string) error {
	peers, err := sig.GetPeers(signalServer, code)
	if err != nil {
		printError(fmt.Sprintf("room %s not found", code))
		return err
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(accentColor)

	fmt.Println()
	fmt.Printf("  %s  %s  %s\n\n",
		headerStyle.Render("ROOM"),
		pink.Render(code),
		dim.Render(fmt.Sprintf("%d peers", len(peers))),
	)

	for _, p := range peers {
		joined := time.Since(p.JoinedAt).Round(time.Second)
		addr := fmt.Sprintf("%s:%d", p.PublicIP, p.PublicPort)
		fmt.Printf("  %s %s\n", green.Render("*"), p.PeerID)
		fmt.Printf("    %s  %s\n",
			dim.Render("addr"),
			addr,
		)
		fmt.Printf("    %s  %s ago\n",
			dim.Render("joined"),
			joined,
		)
		fmt.Println()
	}

	return nil
}
