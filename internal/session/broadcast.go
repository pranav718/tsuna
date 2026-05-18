package session

import (
	"context"
	"time"

	"github.com/pranav718/tsuna/internal/p2p"
	"github.com/pranav718/tsuna/internal/tui"
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
	buffering := false

	if s.mpvOK {
		st := s.mpv.State()
		pos = st.Position.Seconds()
		paused = st.Paused

		isBuffering, err := s.mpv.IsBuffering()
		if err == nil {
			buffering = isBuffering
		}

		s.mu.Lock()
		wasBuf := s.buffering
		s.buffering = buffering
		s.mu.Unlock()

		if buffering && !wasBuf {
			s.transport.Send(p2p.MsgHold, &p2p.HoldPayload{
				Position: pos,
				Reason:   "buffering",
			})
			s.emitUI(tui.UIEvent{Type: tui.UIBufferingStart})
		} else if !buffering && wasBuf {
			s.transport.Send(p2p.MsgResume, &p2p.ResumePayload{
				Position: pos,
			})
			s.emitUI(tui.UIEvent{Type: tui.UIBufferingStop})
		}
	}

	deltas := s.delta.Deltas()
	var syncDelta int64
	for _, d := range deltas {
		ms := d.Milliseconds()
		if ms < 0 {
			ms = -ms
		}
		if ms > syncDelta || syncDelta == 0 {
			syncDelta = d.Milliseconds()
		}
	}

	s.transport.Send(p2p.MsgStateUpdate, &p2p.StateUpdatePayload{
		Position:  pos,
		Paused:    paused,
		Buffering: buffering,
		SyncDelta: syncDelta,
	})

	s.emitUI(tui.UIEvent{Type: tui.UIStateUpdate, Data: tui.StateData{
		Position:  time.Duration(pos * float64(time.Second)),
		Paused:    paused,
		SyncDelta: time.Duration(syncDelta) * time.Millisecond,
	}})
}
