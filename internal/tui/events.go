package tui

import "time"

type UIEventType int

const (
	UIPeerHello      UIEventType = iota
	UIPeerBye
	UIClockSync
	UIStateUpdate
	UICorrection
	UIBufferingStart
	UIBufferingStop
	UIQueueUpdate
	UILog
)

type UIEvent struct {
	Type UIEventType
	Data any
}

type ClockSyncData struct {
	PeerID string
	RTT    time.Duration
	Offset time.Duration
}

type StateData struct {
	Position  time.Duration
	Paused    bool
	SyncDelta time.Duration
}

type CorrectionData struct {
	CorrType string
	Delta    time.Duration
	Target   time.Duration
}

type PeerData struct {
	PeerID      string
	DisplayName string
}

type QueueItemData struct {
	ID       string
	Filename string
	AddedBy  string
}

type QueueData struct {
	Items   []QueueItemData
	Current int
}
