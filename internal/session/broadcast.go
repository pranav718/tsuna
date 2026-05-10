package session

import (
	"context"
	"time"

	"github.com/pranav718/tsuna/internal/p2p"
)

const broadcastInterval = 500 * time.Millisecond

func (s *Session) broadcastLoop(ctx context.Context) {
	ticker := time.NewTicker(broadcastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.broadcastState()
		}
	}
}

func (s *Session) broadcastState() {
	pos := float64(0)
	paused := true

	if s.mpvOK {
		st := s.mpv.State()
		pos = st.Position.Seconds()
		paused = st.Paused
	}

	deltas := s.delta.Deltas()
	var syncDelta int64
	if d, ok := deltas[s.cfg.RemoteID]; ok {
		syncDelta = d.Milliseconds()
	}

	s.transport.Send(p2p.MsgStateUpdate, &p2p.StateUpdatePayload{
		Position:  pos,
		Paused:    paused,
		Buffering: false,
		SyncDelta: syncDelta,
	})
}
