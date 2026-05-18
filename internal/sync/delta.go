package sync

import (
	"fmt"
	"sync"
	"time"
)

const SyncTolerance = 40 * time.Millisecond

const MicroSeekThreshold = 5 * time.Millisecond

const MaxSeekDelta = 2 * time.Second

type PeerPosition struct {
	PeerID    string
	Position  time.Duration
	Paused    bool
	SampledAt time.Time
}

type CorrectionType uint8

const (
	CorrectionNone CorrectionType = iota
	CorrectionMicroSeek
	CorrectionHardResync
	CorrectionPause
	CorrectionResume
)

func (c CorrectionType) String() string {
	switch c {
	case CorrectionNone:
		return "none"
	case CorrectionMicroSeek:
		return "micro-seek"
	case CorrectionHardResync:
		return "hard-resync"
	case CorrectionPause:
		return "pause"
	case CorrectionResume:
		return "resume"
	default:
		return "unknown"
	}
}

type Correction struct {
	Type      CorrectionType
	TargetPos time.Duration
	Delta     time.Duration
	PeerID    string
	EmittedAt time.Time
}

type PositionFunc func() (pos time.Duration, paused bool, err error)

type DeltaEngine struct {
	isHost   bool
	localID  string
	getPos   PositionFunc
	tickRate time.Duration

	mu          sync.RWMutex
	peers       map[string]*PeerPosition
	hostID      string
	lastCorrect time.Time
	cooldown    time.Duration

	CorrectionCh chan Correction
	stopCh       chan struct{}
}

func NewDeltaEngine(localID, hostID string, isHost bool, getPos PositionFunc) *DeltaEngine {
	return &DeltaEngine{
		isHost:       isHost,
		localID:      localID,
		hostID:       hostID,
		getPos:       getPos,
		tickRate:     50 * time.Millisecond,
		peers:        make(map[string]*PeerPosition),
		cooldown:     150 * time.Millisecond,
		CorrectionCh: make(chan Correction, 32),
		stopCh:       make(chan struct{}),
	}
}

func (e *DeltaEngine) Start() {
	go e.loop()
}

func (e *DeltaEngine) Stop() {
	close(e.stopCh)
}

func (e *DeltaEngine) UpdatePeer(p PeerPosition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.peers[p.PeerID] = &p
}

func (e *DeltaEngine) RemovePeer(peerID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.peers, peerID)
}

func (e *DeltaEngine) Deltas() map[string]time.Duration {
	localPos, _, err := e.getPos()
	if err != nil {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	out := make(map[string]time.Duration, len(e.peers))
	for id, p := range e.peers {

		estimated := p.Position
		if !p.Paused {
			estimated += time.Since(p.SampledAt)
		}
		out[id] = localPos - estimated
	}
	return out
}

func (e *DeltaEngine) loop() {
	ticker := time.NewTicker(e.tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.tick()
		}
	}
}

func (e *DeltaEngine) tick() {
	localPos, localPaused, err := e.getPos()
	if err != nil {

		return
	}

	if time.Since(e.lastCorrect) < e.cooldown {
		return
	}

	if e.isHost {
		e.tickHost(localPos, localPaused)
	} else {
		e.tickPeer(localPos, localPaused)
	}
}

func (e *DeltaEngine) tickPeer(localPos time.Duration, localPaused bool) {
	e.mu.RLock()
	host, ok := e.peers[e.hostID]
	e.mu.RUnlock()

	if !ok {

		return
	}

	hostPos := estimatedPos(host)

	delta := localPos - hostPos
	absDelta := abs(delta)

	switch {
	case absDelta < MicroSeekThreshold:

	case absDelta <= SyncTolerance:

	case absDelta > MaxSeekDelta:

		e.emit(Correction{
			Type:      CorrectionHardResync,
			TargetPos: hostPos,
			Delta:     delta,
			PeerID:    e.hostID,
			EmittedAt: time.Now(),
		})

	default:

		target := hostPos + (delta / 2)
		e.emit(Correction{
			Type:      CorrectionMicroSeek,
			TargetPos: target,
			Delta:     delta,
			PeerID:    e.hostID,
			EmittedAt: time.Now(),
		})
	}

	if host.Paused && !localPaused {
		e.emit(Correction{
			Type:      CorrectionPause,
			Delta:     0,
			PeerID:    e.hostID,
			EmittedAt: time.Now(),
		})
	} else if !host.Paused && localPaused {
		e.emit(Correction{
			Type:      CorrectionResume,
			TargetPos: hostPos,
			Delta:     0,
			PeerID:    e.hostID,
			EmittedAt: time.Now(),
		})
	}
}

func (e *DeltaEngine) tickHost(localPos time.Duration, _ bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for id, p := range e.peers {
		estimated := estimatedPos(p)
		delta := localPos - estimated

		if delta > MaxSeekDelta {

			e.emit(Correction{
				Type:      CorrectionPause,
				Delta:     delta,
				PeerID:    id,
				EmittedAt: time.Now(),
			})
			return
		}
	}
}

func (e *DeltaEngine) emit(c Correction) {
	e.lastCorrect = time.Now()
	select {
	case e.CorrectionCh <- c:
	default:

	}
}

type SyncReport struct {
	LocalPos time.Duration
	Peers    []PeerSyncStatus
}

type PeerSyncStatus struct {
	PeerID    string
	Delta     time.Duration
	Synced    bool
	Buffering bool
}

func (e *DeltaEngine) Report() (SyncReport, error) {
	localPos, _, err := e.getPos()
	if err != nil {
		return SyncReport{}, fmt.Errorf("delta: get local pos: %w", err)
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	report := SyncReport{LocalPos: localPos}
	for id, p := range e.peers {
		estimated := estimatedPos(p)
		delta := localPos - estimated
		report.Peers = append(report.Peers, PeerSyncStatus{
			PeerID:    id,
			Delta:     delta,
			Synced:    abs(delta) <= SyncTolerance,
			Buffering: p.Paused,
		})
	}
	return report, nil
}

func estimatedPos(p *PeerPosition) time.Duration {
	if p.Paused {
		return p.Position
	}
	return p.Position + time.Since(p.SampledAt)
}
