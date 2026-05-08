package sync

import (
	"sync"
	"time"
)

type ClockOffset struct {
	Delta   time.Duration
	RTT     time.Duration
	Sampled time.Time
}

type Reconciler struct {
	mu         sync.RWMutex
	offsets    map[string]*ClockOffset
	correction time.Duration
	sampleRate time.Duration
	tolerance  time.Duration
	stopCh     chan struct{}
}

func NewReconciler() *Reconciler {
	return &Reconciler{
		offsets:    make(map[string]*ClockOffset),
		sampleRate: 2 * time.Second,
		tolerance:  40 * time.Millisecond,
		stopCh:     make(chan struct{}),
	}
}

func (r *Reconciler) Start() {
	go r.loop()
}

func (r *Reconciler) Stop() {
	close(r.stopCh)
}

func (r *Reconciler) RecordSample(peerID string, rtt time.Duration, offset time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.offsets[peerID] = &ClockOffset{
		Delta:   offset,
		RTT:     rtt,
		Sampled: time.Now(),
	}
}

func (r *Reconciler) Correction() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.correction
}

func (r *Reconciler) NeedsSync() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, o := range r.offsets {
		if abs(o.Delta) > r.tolerance {
			return true
		}
	}
	return false
}

func (r *Reconciler) RemovePeer(peerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.offsets, peerID)
}

func (r *Reconciler) loop() {
	ticker := time.NewTicker(r.sampleRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.recompute()
		case <-r.stopCh:
			return
		}
	}
}

func (r *Reconciler) recompute() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.offsets) == 0 {
		return
	}

	var total time.Duration
	var count int

	now := time.Now()
	staleThreshold := 10 * time.Second

	for _, o := range r.offsets {
		if now.Sub(o.Sampled) > staleThreshold {
			continue
		}
		total += o.Delta
		count++
	}

	if count == 0 {
		return
	}

	r.correction = total / time.Duration(count)
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
