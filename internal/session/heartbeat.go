package session

import (
	"context"
	"log"
	"time"

	"github.com/pranav718/tsuna/internal/p2p"
	"github.com/pranav718/tsuna/internal/room"
	"github.com/pranav718/tsuna/internal/tui"
)

const (
	heartbeatInterval = 1 * time.Second
	peerTimeout       = 5 * time.Second
	watchdogInterval  = 2 * time.Second
)

func (s *Session) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.transport.Send(p2p.MsgHeartbeat, &p2p.HeartbeatPayload{})
		}
	}
}

func (s *Session) watchdogLoop(ctx context.Context) {
	ticker := time.NewTicker(watchdogInterval)
	defer ticker.Stop()

	peerLost := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			lastSeen := s.lastPeerMsg
			s.mu.RUnlock()

			if lastSeen.IsZero() {
				continue
			}

			if time.Since(lastSeen) > peerTimeout {
				if !peerLost {
					peerLost = true
					log.Printf("[session] peer %s timed out (no messages for %v)", s.cfg.RemoteID, peerTimeout)
					s.room.Send(room.Event{PeerID: s.cfg.RemoteID, Type: room.EvPeerLeft})
					s.emitUI(tui.UIEvent{Type: tui.UIPeerBye})
					s.emitUI(tui.UIEvent{Type: tui.UILog, Data: "peer lost. waiting for reconnection"})
				}
			} else if peerLost {
				peerLost = false
				log.Printf("[session] peer %s recovered", s.cfg.RemoteID)
				s.emitUI(tui.UIEvent{Type: tui.UIPeerHello, Data: tui.PeerData{
					PeerID: s.cfg.RemoteID, DisplayName: s.cfg.RemoteID,
				}})
				s.emitUI(tui.UIEvent{Type: tui.UILog, Data: "peer reconnected"})
			}
		}
	}
}

func (s *Session) touchPeerSeen() {
	s.mu.Lock()
	s.lastPeerMsg = time.Now()
	s.mu.Unlock()
}
