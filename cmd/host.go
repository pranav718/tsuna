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

	sock, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return fmt.Errorf("failed to bind UDP socket: %w", err)
	}

	var pub p2p.PublicEndpoint
	if localMode {
		localAddr := sock.LocalAddr().(*net.UDPAddr)
		pub = p2p.PublicEndpoint{IP: net.ParseIP("127.0.0.1"), Port: localAddr.Port}
		printStepDone(1, fmt.Sprintf("local mode: %s", dim.Render(pub.String())))
	} else {
		err = runWithSpinner(1, "discovering public endpoint via STUN", func() error {
			var e error
			pub, e = p2p.DiscoverWithConn(sock)
			return e
		})
		if err != nil {
			sock.Close()
			printError("STUN discovery failed :(")
			return fmt.Errorf("STUN discovery failed: %w", err)
		}
		printStepDone(1, fmt.Sprintf("public endpoint: %s", dim.Render(pub.String())))
	}

	err = runWithSpinner(2, "registering with signaling server", func() error {
		_, e := sig.Register(signalServer, sig.RegisterRequest{
			RoomCode:   code,
			PeerID:     localID,
			PublicIP:   pub.IP.String(),
			PublicPort: pub.Port,
		})
		return e
	})
	if err != nil {
		sock.Close()
		printError("signaling registration failed :(")
		return fmt.Errorf("signaling registration failed: %w", err)
	}
	printStepDone(2, "registered")

	printRoomCode(code)
	printInfo("peer id", localID)
	printWaiting("share this code with your friends. waiting for them to join...")

	peerSet := p2p.NewPeerSet()

	// wait for first peer before launching TUI
	firstPeer, err := waitForFirstPeer(signalServer, code, localID, sock, peerSet)
	if err != nil {
		sock.Close()
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// keep polling for more peers in the background
	go acceptMorePeers(ctx, signalServer, code, localID, peerSet)

	uiCh := make(chan tui.UIEvent, 64)

	sess := session.New(session.Config{
		LocalID:   localID,
		RemoteID:  firstPeer,
		RoomCode:  code,
		IsHost:    true,
		MpvSocket: mpvSocket,
		Transport: peerSet,
		UIEvents:  uiCh,
	})

	go func() {
		sess.Run(ctx)
		close(uiCh)
	}()

	model := tui.NewModel(code, localID, firstPeer, true, uiCh, cancel)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	sig.Leave(signalServer, code, localID)
	peerSet.Close()
	return nil
}

func waitForFirstPeer(signalURL, code, localID string, sock *net.UDPConn, peerSet *p2p.PeerSet) (string, error) {
	for {
		peers, err := sig.GetPeers(signalURL, code)
		if err != nil {
			return "", fmt.Errorf("failed to poll peers: %w", err)
		}
		for _, p := range peers {
			if p.PeerID != localID {
				printStepDone(3, fmt.Sprintf("peer joined: %s", green.Render(p.PeerID)))
				if err := connectPeer(p, sock, signalURL, code, localID, peerSet); err != nil {
					printError(fmt.Sprintf("failed to connect %s: %v", p.PeerID, err))
					continue
				}
				return p.PeerID, nil
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func acceptMorePeers(ctx context.Context, signalURL, code, localID string, peerSet *p2p.PeerSet) {
	known := make(map[string]bool)
	known[localID] = true

	for _, id := range peerSet.IDs() {
		known[id] = true
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, err := sig.GetPeers(signalURL, code)
			if err != nil {
				continue
			}
			for _, p := range peers {
				if !known[p.PeerID] {
					known[p.PeerID] = true
					// connect via relay for additional peers (UDP socket already in use)
					relay, err := p2p.NewRelayTransport(signalURL, code, localID)
					if err != nil {
						continue
					}
					relay.Start()
					peerSet.Add(p.PeerID, relay)
				}
			}
		}
	}
}

func connectPeer(peer sig.PeerInfo, sock *net.UDPConn, signalURL, code, localID string, peerSet *p2p.PeerSet) error {
	remoteIP := net.ParseIP(peer.PublicIP)
	peerEP := p2p.PeerEndpoint{IP: remoteIP, Port: peer.PublicPort}

	var result p2p.PunchResult
	err := runWithSpinner(4, "punching through NAT", func() error {
		puncher := p2p.NewPuncherWithConn(sock, peerEP, p2p.DefaultPunchConfig())
		var e error
		result, e = puncher.Punch()
		return e
	})
	if err != nil {
		printError("hole punch failed, falling back to relay")

		relay, err := p2p.NewRelayTransport(signalURL, code, localID)
		if err != nil {
			return fmt.Errorf("relay failed: %w", err)
		}
		printStepDone(5, "relayed through signal server")
		relay.Start()
		peerSet.Add(peer.PeerID, relay)
		return nil
	}

	printConnected(result.RTT.String())
	transport := p2p.NewTransport(result.Conn, peerEP.UDPAddr(), localID, code)
	transport.Start()
	peerSet.Add(peer.PeerID, transport)
	return nil
}

func generatePeerID() string {
	host, _ := os.Hostname()
	if host == "" {
		host = "peer"
	}
	return fmt.Sprintf("%s-%04x", host, rand.Intn(0xFFFF))
}
