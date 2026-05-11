package session

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pranav718/tsuna/internal/mpv"
	"github.com/pranav718/tsuna/internal/p2p"
	"github.com/pranav718/tsuna/internal/room"
	tsync "github.com/pranav718/tsuna/internal/sync"
	"github.com/pranav718/tsuna/internal/tui"
)

type Config struct {
	LocalID    string
	RemoteID   string
	RoomCode   string
	IsHost     bool
	MpvSocket  string
	Transport  *p2p.Transport
	UIEvents   chan<- tui.UIEvent
}

type Session struct {
	cfg        Config
	transport  *p2p.Transport
	mpv        *mpv.Bridge
	room       *room.Room
	reconciler *tsync.Reconciler
	delta      *tsync.DeltaEngine
	mpvOK      bool

	mu          sync.RWMutex
	lastPeerMsg time.Time
	buffering   bool
}

func (s *Session) emitUI(ev tui.UIEvent) {
	if s.cfg.UIEvents == nil {
		return
	}
	select {
	case s.cfg.UIEvents <- ev:
	default:
	}
}

func New(cfg Config) *Session {
	return &Session{
		cfg:       cfg,
		transport: cfg.Transport,
	}
}

func (s *Session) Run(ctx context.Context) error {
	s.mpv = mpv.NewBridge(s.cfg.MpvSocket)
	if err := s.mpv.Connect(); err != nil {
		log.Printf("[session] mpv not connected: %v (sync will run without playback control)", err)
		s.mpvOK = false
	} else {
		s.mpvOK = true
		defer s.mpv.Close()
		log.Printf("[session] connected to mpv at %s", s.cfg.MpvSocket)
	}

	hostID := s.cfg.LocalID
	if !s.cfg.IsHost {
		hostID = s.cfg.RemoteID
	}

	s.room = room.NewRoom(s.cfg.RoomCode, hostID)
	defer s.room.Shutdown()

	s.room.Send(room.Event{
		PeerID: s.cfg.RemoteID,
		Type:   room.EvPeerJoined,
	})

	s.reconciler = tsync.NewReconciler()
	s.reconciler.Start()
	defer s.reconciler.Stop()

	s.delta = tsync.NewDeltaEngine(s.cfg.LocalID, hostID, s.cfg.IsHost, s.positionFunc())
	s.delta.Start()
	defer s.delta.Stop()

	s.transport.Send(p2p.MsgHello, &p2p.HelloPayload{
		DisplayName: s.cfg.LocalID,
		Version:     "0.1.0",
		IsHost:      s.cfg.IsHost,
	})

	go s.clockSyncLoop(ctx)
	go s.broadcastLoop(ctx)
	go s.heartbeatLoop(ctx)
	go s.watchdogLoop(ctx)

	log.Printf("[session] running — room=%s peer=%s host=%v", s.cfg.RoomCode, s.cfg.RemoteID, s.cfg.IsHost)

	return s.loop(ctx)
}

func (s *Session) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			s.transport.Send(p2p.MsgBye, &p2p.ByePayload{Reason: "session ended"})
			return nil

		case env := <-s.transport.RecvCh:
			s.handleMessage(env)

		case c := <-s.delta.CorrectionCh:
			s.applyCorrection(c)

		case cmd := <-s.room.CommandCh:
			s.executeCommand(cmd)
		}
	}
}

func (s *Session) handleMessage(env *p2p.Envelope) {
	s.touchPeerSeen()

	switch env.Type {
	case p2p.MsgHello:
		var pl p2p.HelloPayload
		if p2p.DecodePayload(env, &pl) == nil {
			log.Printf("[session] peer hello: %s (v%s, host=%v)", pl.DisplayName, pl.Version, pl.IsHost)
			s.room.Send(room.Event{PeerID: env.SenderID, Type: room.EvPeerReady})
			s.emitUI(tui.UIEvent{Type: tui.UIPeerHello, Data: tui.PeerData{
				PeerID: env.SenderID, DisplayName: pl.DisplayName,
			}})
		}

	case p2p.MsgHeartbeat:

	case p2p.MsgBye:
		log.Printf("[session] peer %s disconnected", env.SenderID)
		s.room.Send(room.Event{PeerID: env.SenderID, Type: room.EvPeerLeft})
		s.emitUI(tui.UIEvent{Type: tui.UIPeerBye})

	case p2p.MsgPing:
		s.handlePing(env)

	case p2p.MsgPong:
		s.handlePong(env)

	case p2p.MsgPlay:
		var pl p2p.PlayPayload
		if p2p.DecodePayload(env, &pl) == nil {
			log.Printf("[session] play at %.2fs (scheduled %v from now)", pl.Position, time.Until(pl.At))
			if s.mpvOK {
				s.mpv.Seek(time.Duration(pl.Position * float64(time.Second)))
				delay := time.Until(pl.At)
				if delay > 0 {
					time.Sleep(delay)
				}
				s.mpv.Play()
			}
		}

	case p2p.MsgPause:
		var pl p2p.PausePayload
		if p2p.DecodePayload(env, &pl) == nil {
			log.Printf("[session] pause at %.2fs", pl.Position)
			if s.mpvOK {
				s.mpv.Pause()
			}
		}

	case p2p.MsgSeek:
		var pl p2p.SeekPayload
		if p2p.DecodePayload(env, &pl) == nil {
			log.Printf("[session] seek to %.2fs", pl.Position)
			if s.mpvOK {
				s.mpv.Seek(time.Duration(pl.Position * float64(time.Second)))
			}
		}

	case p2p.MsgStateUpdate:
		var pl p2p.StateUpdatePayload
		if p2p.DecodePayload(env, &pl) == nil {
			s.delta.UpdatePeer(tsync.PeerPosition{
				PeerID:    env.SenderID,
				Position:  time.Duration(pl.Position * float64(time.Second)),
				Paused:    pl.Paused,
				SampledAt: env.SentAt,
			})
		}

	case p2p.MsgHold:
		log.Printf("[session] peer %s is buffering", env.SenderID)
		s.room.Send(room.Event{PeerID: env.SenderID, Type: room.EvPeerBuffering})

	case p2p.MsgResume:
		log.Printf("[session] peer %s ready to resume", env.SenderID)
		s.room.Send(room.Event{PeerID: env.SenderID, Type: room.EvPeerResumed})
	}
}

func (s *Session) applyCorrection(c tsync.Correction) {
	if !s.mpvOK {
		return
	}

	switch c.Type {
	case tsync.CorrectionMicroSeek:
		log.Printf("[session] micro-seek: delta=%v target=%v", c.Delta, c.TargetPos)
		s.mpv.SeekExact(c.TargetPos)
	case tsync.CorrectionHardResync:
		log.Printf("[session] hard resync: delta=%v target=%v", c.Delta, c.TargetPos)
		s.mpv.Seek(c.TargetPos)
	case tsync.CorrectionPause:
		log.Printf("[session] sync pause (peer %s too far ahead)", c.PeerID)
		s.mpv.Pause()
	case tsync.CorrectionResume:
		log.Printf("[session] sync resume to %v", c.TargetPos)
		s.mpv.Seek(c.TargetPos)
		s.mpv.Play()
	}

	s.emitUI(tui.UIEvent{Type: tui.UICorrection, Data: tui.CorrectionData{
		CorrType: c.Type.String(),
		Delta:    c.Delta,
		Target:   c.TargetPos,
	}})
}

func (s *Session) executeCommand(cmd room.Command) {
	switch cmd.Type {
	case room.CmdPlay:
		pos := float64(0)
		if s.mpvOK {
			if p, err := s.mpv.GetPosition(); err == nil {
				pos = p.Seconds()
			}
			s.mpv.Play()
		}
		s.transport.Send(p2p.MsgPlay, &p2p.PlayPayload{
			Position: pos,
			At:       cmd.At,
		})

	case room.CmdPause:
		pos := float64(0)
		if s.mpvOK {
			if p, err := s.mpv.GetPosition(); err == nil {
				pos = p.Seconds()
			}
			s.mpv.Pause()
		}
		s.transport.Send(p2p.MsgPause, &p2p.PausePayload{Position: pos})

	case room.CmdSeek:
		if s.mpvOK {
			s.mpv.Seek(cmd.Position)
		}
		s.transport.Send(p2p.MsgSeek, &p2p.SeekPayload{
			Position: cmd.Position.Seconds(),
		})
	}
}

func (s *Session) positionFunc() tsync.PositionFunc {
	return func() (time.Duration, bool, error) {
		if !s.mpvOK {
			return 0, true, nil
		}
		st := s.mpv.State()
		return st.Position, st.Paused, nil
	}
}
