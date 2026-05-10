package session

import (
	"context"
	"log"
	"time"

	"github.com/pranav718/tsuna/internal/p2p"
)

const clockSyncInterval = 2 * time.Second

func (s *Session) clockSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(clockSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.transport.Send(p2p.MsgPing, &p2p.PingPayload{
				T1: time.Now(),
			})
		}
	}
}

func (s *Session) handlePing(env *p2p.Envelope) {
	var pl p2p.PingPayload
	if err := p2p.DecodePayload(env, &pl); err != nil {
		return
	}

	now := time.Now()
	s.transport.Send(p2p.MsgPong, &p2p.PongPayload{
		T1: pl.T1,
		T2: now,
		T3: time.Now(),
	})
}

func (s *Session) handlePong(env *p2p.Envelope) {
	t4 := time.Now()

	var pl p2p.PongPayload
	if err := p2p.DecodePayload(env, &pl); err != nil {
		return
	}

	rtt := (t4.Sub(pl.T1)) - (pl.T3.Sub(pl.T2))
	offset := (pl.T2.Sub(pl.T1) + pl.T3.Sub(t4)) / 2

	if rtt < 0 {
		rtt = t4.Sub(pl.T1)
	}

	s.reconciler.RecordSample(env.SenderID, rtt, offset)
	log.Printf("[clocksync] peer=%s rtt=%v offset=%v", env.SenderID, rtt, offset)
}
