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
		return fmt.Errorf("invalid room code %q -- must be exactly 6 characters", code)
	}

	printBanner()

	localID := generatePeerID()

	var pub p2p.PublicEndpoint
	err := runWithSpinner(1, "discovering public endpoint via STUN", func() error {
		var e error
		pub, e = p2p.DiscoverPublicEndpoint()
		return e
	})
	if err != nil {
		printError("STUN discovery failed :(")
		return fmt.Errorf("STUN discovery failed: %w", err)
	}
	printStepDone(1, fmt.Sprintf("public endpoint: %s", dim.Render(pub.String())))

	var resp *sig.RegisterResponse
	err = runWithSpinner(2, "registering with signaling server", func() error {
		var e error
		resp, e = sig.Register(signalServer, sig.RegisterRequest{
			RoomCode:   code,
			PeerID:     localID,
			PublicIP:   pub.IP.String(),
			PublicPort: pub.Port,
		})
		return e
	})
	if err != nil {
		printError("signaling registration failed :(")
		return fmt.Errorf("signaling registration failed: %w", err)
	}
	printStepDone(2, fmt.Sprintf("joined room %s", pink.Render(code)))

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

	printStepDone(3, fmt.Sprintf("host found: %s", green.Render(hostPeer.PeerID)))

	remoteIP := net.ParseIP(hostPeer.PublicIP)
	peerEP := p2p.PeerEndpoint{IP: remoteIP, Port: hostPeer.PublicPort}

	var result p2p.PunchResult
	err = runWithSpinner(4, "punching through NAT", func() error {
		puncher, e := p2p.NewPuncher(peerEP, p2p.DefaultPunchConfig())
		if e != nil {
			return e
		}
		result, e = puncher.Punch()
		return e
	})
	if err != nil {
		printError("hole punch timed out :(")
		return fmt.Errorf("hole punch failed: %w", err)
	}

	printConnected(result.RTT.String())

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
