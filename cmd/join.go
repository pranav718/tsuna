package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pranav718/tsuna/internal/p2p"
	"github.com/pranav718/tsuna/internal/session"
	sig "github.com/pranav718/tsuna/internal/signal"
	"github.com/pranav718/tsuna/internal/tui"
	"github.com/spf13/cobra"
)

var joinCmd = &cobra.Command{
	Use:   "join <CODE>",
	Short: "join a watch party room",
	Long:  "join an existing tsuna room using a 6-character room code.",
	Args:  cobra.ExactArgs(1),
	RunE:  runJoin,
}

func init() {
	rootCmd.AddCommand(joinCmd)
}

func runJoin(cmd *cobra.Command, args []string) error {
	code := strings.ToUpper(strings.TrimSpace(args[0]))

	if len(code) != 6 {
		return fmt.Errorf("invalid room code %q — must be exactly 6 characters", code)
	}

	printBanner()
	printStep(1, fmt.Sprintf("joining room %s", pink.Render(code)))

	localID := generatePeerID()

	printStep(2, "discovering public endpoint via STUN...")
	pub, err := p2p.DiscoverPublicEndpoint()
	if err != nil {
		printError("STUN discovery failed")
		return fmt.Errorf("STUN discovery failed: %w", err)
	}
	printStepDone(2, fmt.Sprintf("public endpoint: %s", dim.Render(pub.String())))

	printStep(3, "registering with signaling server...")
	resp, err := sig.Register(signalServer, sig.RegisterRequest{
		RoomCode:   code,
		PeerID:     localID,
		PublicIP:   pub.IP.String(),
		PublicPort: pub.Port,
	})
	if err != nil {
		printError("signaling registration failed")
		return fmt.Errorf("signaling registration failed: %w", err)
	}
	printStepDone(3, "registered")

	var hostPeer sig.PeerInfo
	for _, p := range resp.Peers {
		if p.PeerID != localID {
			hostPeer = p
			break
		}
	}

	if hostPeer.PeerID == "" {
		printError("no host found")
		return fmt.Errorf("no host found in room %s", code)
	}

	printStepDone(4, fmt.Sprintf("host found: %s", green.Render(hostPeer.PeerID)))

	printStep(5, "punching through NAT...")
	remoteIP := net.ParseIP(hostPeer.PublicIP)
	peerEP := p2p.PeerEndpoint{IP: remoteIP, Port: hostPeer.PublicPort}

	puncher, err := p2p.NewPuncher(peerEP, p2p.DefaultPunchConfig())
	if err != nil {
		printError("punch setup failed")
		return fmt.Errorf("punch setup failed: %w", err)
	}

	result, err := puncher.Punch()
	if err != nil {
		printError("hole punch timed out")
		return fmt.Errorf("hole punch failed: %w", err)
	}

	printConnected(result.RTT.String())
	printStep(6, "launching dashboard...")
	fmt.Println()

	transport := p2p.NewTransport(result.Conn, peerEP.UDPAddr(), localID, code)
	transport.Start()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	uiCh := make(chan tui.UIEvent, 64)

	sess := session.New(session.Config{
		LocalID:   localID,
		RemoteID:  hostPeer.PeerID,
		RoomCode:  code,
		IsHost:    false,
		MpvSocket: mpvSocket,
		Transport: transport,
		UIEvents:  uiCh,
	})

	go func() {
		sess.Run(ctx)
		close(uiCh)
	}()

	model := tui.NewModel(code, localID, hostPeer.PeerID, false, uiCh, cancel)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	sig.Leave(signalServer, code, localID)
	transport.Close()
	return nil
}
