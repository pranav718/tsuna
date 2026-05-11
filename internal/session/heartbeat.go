package session

import (
	"context"
	"log"
	"time"

	"github.com/pranav718/tsuna/internal/p2p"
	"github.com/pranav718/tsuna/internal/room"
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
				log.Printf("[session] peer %s timed out (no messages for %v)", s.cfg.RemoteID, peerTimeout)
				s.room.Send(room.Event{PeerID: s.cfg.RemoteID, Type: room.EvPeerLeft})
			}
		}
	}
}

func (s *Session) touchPeerSeen() {
	s.mu.Lock()
	s.lastPeerMsg = time.Now()
	s.mu.Unlock()
}
