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

	localID := generatePeerID()

	fmt.Printf("discovering public endpoint via STUN...\n")
	pub, err := p2p.DiscoverPublicEndpoint()
	if err != nil {
		return fmt.Errorf("STUN discovery failed: %w", err)
	}
	fmt.Printf("public endpoint: %s\n", pub)

	fmt.Printf("registering with signaling server...\n")
	resp, err := sig.Register(signalServer, sig.RegisterRequest{
		RoomCode:   code,
		PeerID:     localID,
		PublicIP:   pub.IP.String(),
		PublicPort: pub.Port,
	})
	if err != nil {
		return fmt.Errorf("signaling registration failed: %w", err)
	}

	var hostPeer sig.PeerInfo
	for _, p := range resp.Peers {
		if p.PeerID != localID {
			hostPeer = p
			break
		}
	}

	if hostPeer.PeerID == "" {
		return fmt.Errorf("no host found in room %s — is anyone hosting?", code)
	}

	fmt.Printf("host found: %s (%s:%d)\n", hostPeer.PeerID, hostPeer.PublicIP, hostPeer.PublicPort)
	fmt.Printf("punching through NAT...\n")

	remoteIP := net.ParseIP(hostPeer.PublicIP)
	peerEP := p2p.PeerEndpoint{IP: remoteIP, Port: hostPeer.PublicPort}

	puncher, err := p2p.NewPuncher(peerEP, p2p.DefaultPunchConfig())
	if err != nil {
		return fmt.Errorf("punch setup failed: %w", err)
	}

	result, err := puncher.Punch()
	if err != nil {
		return fmt.Errorf("hole punch failed: %w", err)
	}

	fmt.Printf("connected! rtt=%v\n", result.RTT)
	fmt.Printf("launching dashboard...\n\n")

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
