package room

import (
	"sync"
	"sync/atomic"
	"time"
)

type PeerState uint32

const (
	PeerConnecting  PeerState = iota
	PeerReady
	PeerPlaying
	PeerBuffering
	PeerSeeking
	PeerDisconnected
)

func (s PeerState) String() string {
	switch s {
	case PeerConnecting:
		return "connecting"
	case PeerReady:
		return "ready"
	case PeerPlaying:
		return "playing"
	case PeerBuffering:
		return "buffering"
	case PeerSeeking:
		return "seeking"
	case PeerDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

type RoomState uint32

const (
	RoomIdle     RoomState = iota
	RoomReady
	RoomPlaying
	RoomHolding
	RoomSeeking
	RoomFinished
)

type Peer struct {
	ID        string
	Addr      string
	state     uint32
	ClockOffset int64
	LastSeen  time.Time
	RTT       time.Duration
}

func (p *Peer) State() PeerState {
	return PeerState(atomic.LoadUint32(&p.state))
}
func (p *Peer) SetState(s PeerState) {
	atomic.StoreUint32(&p.state, uint32(s))
}

type Event struct {
	PeerID  string
	Type    EventType
	Payload any
}

type EventType uint8

const (
	EvPeerJoined      EventType = iota
	EvPeerLeft
	EvPeerReady
	EvPeerBuffering
	EvPeerResumed
	EvPlayRequested
	EvPauseRequested
	EvSeekRequested
	EvSyncTick
	EvShutdown
)

type Room struct {
	Code  string
	Host  string

	peers     map[string]*Peer
	mu        sync.RWMutex
	roomState uint32

	eventCh   chan Event
	CommandCh chan Command
	done      chan struct{}
}

type Command struct {
	Type     CommandType
	Position time.Duration
	At       time.Time
}

type CommandType uint8

const (
	CmdPlay  CommandType = iota
	CmdPause
	CmdSeek
)

func NewRoom(code, hostID string) *Room {
	r := &Room{
		Code:      code,
		Host:      hostID,
		peers:     make(map[string]*Peer),
		eventCh:   make(chan Event, 64),
		CommandCh: make(chan Command, 16),
		done:      make(chan struct{}),
	}
	atomic.StoreUint32(&r.roomState, uint32(RoomIdle))
	go r.loop()
	return r
}

func (r *Room) RoomStateVal() RoomState {
	return RoomState(atomic.LoadUint32(&r.roomState))
}
func (r *Room) Send(ev Event) {
	select {
	case r.eventCh <- ev:
	default:
	}
}

func (r *Room) Shutdown() {
	close(r.done)
}

func (r *Room) Peers() []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
		out = append(out, p)
	}
	return out
}

func (r *Room) loop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.done:
			return

		case <-ticker.C:
			r.Send(Event{Type: EvSyncTick})

		case ev := <-r.eventCh:
			r.handle(ev)
		}
	}
}

func (r *Room) handle(ev Event) {
	switch ev.Type {
	case EvPeerJoined:
		r.mu.Lock()
		r.peers[ev.PeerID] = &Peer{ID: ev.PeerID, LastSeen: time.Now()}
		r.peers[ev.PeerID].SetState(PeerConnecting)
		r.mu.Unlock()

	case EvPeerLeft, EvShutdown:
		r.mu.Lock()
		if p, ok := r.peers[ev.PeerID]; ok {
			p.SetState(PeerDisconnected)
			delete(r.peers, ev.PeerID)
		}
		r.mu.Unlock()
		r.recalcRoomState()

	case EvPeerReady:
		r.setPeerState(ev.PeerID, PeerReady)
		r.recalcRoomState()

	case EvPeerBuffering:
		r.setPeerState(ev.PeerID, PeerBuffering)
		atomic.StoreUint32(&r.roomState, uint32(RoomHolding))
		r.CommandCh <- Command{Type: CmdPause, At: time.Now()}

	case EvPeerResumed:
		r.setPeerState(ev.PeerID, PeerPlaying)
		r.recalcRoomState()

	case EvPlayRequested:
		if r.RoomStateVal() == RoomReady || r.RoomStateVal() == RoomHolding {
			atomic.StoreUint32(&r.roomState, uint32(RoomPlaying))
			r.CommandCh <- Command{Type: CmdPlay, At: time.Now().Add(200 * time.Millisecond)}
		}

	case EvPauseRequested:
		atomic.StoreUint32(&r.roomState, uint32(RoomIdle))
		r.CommandCh <- Command{Type: CmdPause, At: time.Now()}

	case EvSeekRequested:
		atomic.StoreUint32(&r.roomState, uint32(RoomSeeking))
		if pos, ok := ev.Payload.(time.Duration); ok {
			r.CommandCh <- Command{Type: CmdSeek, Position: pos, At: time.Now()}
		}

	case EvSyncTick:
		r.recalcRoomState()
	}
}
func (r *Room) setPeerState(id string, s PeerState) {
	r.mu.RLock()
	p, ok := r.peers[id]
	r.mu.RUnlock()
	if ok {
		p.SetState(s)
		p.LastSeen = time.Now()
	}
}

func (r *Room) recalcRoomState() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.peers) == 0 {
		atomic.StoreUint32(&r.roomState, uint32(RoomIdle))
		return
	}

	allReady := true
	allPlaying := true

	for _, p := range r.peers {
		s := p.State()
		if s == PeerBuffering {
			return
		}
		if s != PeerReady && s != PeerPlaying {
			allReady = false
			allPlaying = false
		}
		if s != PeerPlaying {
			allPlaying = false
		}
	}

	switch {
	case allPlaying:
		atomic.StoreUint32(&r.roomState, uint32(RoomPlaying))
	case allReady:
		atomic.StoreUint32(&r.roomState, uint32(RoomReady))
	}
}
