package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pranav718/tsuna/internal/p2p"
	"github.com/pranav718/tsuna/internal/room"
	"github.com/pranav718/tsuna/internal/session"
	sig "github.com/pranav718/tsuna/internal/signal"
	"github.com/pranav718/tsuna/internal/tui"
	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "create a new watch room",
	Long:  "spin up a new tsuna room and get a 6-character code to share with your friends :D",
	RunE:  runHost,
}

func init() {
	rootCmd.AddCommand(hostCmd)
}

func runHost(cmd *cobra.Command, args []string) error {
	printBanner()

	code, err := room.GenerateCode()
	if err != nil {
		return fmt.Errorf("failed to generate room code: %w", err)
	}

	localID := generatePeerID()

	printStep(1, "discovering public endpoint via STUN... :)")
	pub, err := p2p.DiscoverPublicEndpoint()
	if err != nil {
		printError("STUN discovery failed :(")
		return fmt.Errorf("STUN discovery failed: %w", err)
	}
	printStepDone(1, fmt.Sprintf("public endpoint: %s", dim.Render(pub.String())))

	printStep(2, "registering with signaling server... :O")
	_, err = sig.Register(signalServer, sig.RegisterRequest{
		RoomCode:   code,
		PeerID:     localID,
		PublicIP:   pub.IP.String(),
		PublicPort: pub.Port,
	})
	if err != nil {
		printError("signaling registration failed :(")
		return fmt.Errorf("signaling registration failed: %w", err)
	}
	printStepDone(2, "registered :)")

	printRoomCode(code)
	printInfo("peer id", localID)
	printWaiting("share this code with your friend. waiting for them to join... :)")

	var remotePeer sig.PeerInfo
	for {
		peers, err := sig.GetPeers(signalServer, code)
		if err != nil {
			return fmt.Errorf("failed to poll peers: %w", err)
		}
		for _, p := range peers {
			if p.PeerID != localID {
				remotePeer = p
				goto peerFound
			}
		}
		time.Sleep(1 * time.Second)
	}

peerFound:
	printStepDone(3, fmt.Sprintf("peer joined: %s", green.Render(remotePeer.PeerID)))

	printStep(4, "punching through NAT... :)")
	remoteIP := net.ParseIP(remotePeer.PublicIP)
	peerEP := p2p.PeerEndpoint{IP: remoteIP, Port: remotePeer.PublicPort}

	puncher, err := p2p.NewPuncher(peerEP, p2p.DefaultPunchConfig())
	if err != nil {
		printError("punch setup failed :(")
		return fmt.Errorf("punch setup failed: %w", err)
	}

	result, err := puncher.Punch()
	if err != nil {
		printError("hole punch timed out :(")
		return fmt.Errorf("hole punch failed: %w", err)
	}

	printConnected(result.RTT.String())
	printStep(5, "launching dashboard... :)")
	fmt.Println()

	transport := p2p.NewTransport(result.Conn, peerEP.UDPAddr(), localID, code)
	transport.Start()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	uiCh := make(chan tui.UIEvent, 64)

	sess := session.New(session.Config{
		LocalID:   localID,
		RemoteID:  remotePeer.PeerID,
		RoomCode:  code,
		IsHost:    true,
		MpvSocket: mpvSocket,
		Transport: transport,
		UIEvents:  uiCh,
	})

	go func() {
		sess.Run(ctx)
		close(uiCh)
	}()

	model := tui.NewModel(code, localID, remotePeer.PeerID, true, uiCh, cancel)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	sig.Leave(signalServer, code, localID)
	transport.Close()
	return nil
}

func generatePeerID() string {
	host, _ := os.Hostname()
	if host == "" {
		host = "peer"
	}
	return fmt.Sprintf("%s-%04x", host, rand.Intn(0xFFFF))
}
