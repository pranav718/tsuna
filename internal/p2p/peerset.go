package p2p

import (
	"log"
	"sync"
)

type PeerSet struct {
	peers  map[string]Sender
	mu     sync.RWMutex
	recvCh chan *Envelope
	stopCh chan struct{}
}

func NewPeerSet() *PeerSet {
	return &PeerSet{
		peers:  make(map[string]Sender),
		recvCh: make(chan *Envelope, 128),
		stopCh: make(chan struct{}),
	}
}

func (ps *PeerSet) Add(id string, s Sender) {
	ps.mu.Lock()
	ps.peers[id] = s
	ps.mu.Unlock()

	go ps.pipeRecv(id, s)
}

func (ps *PeerSet) Remove(id string) {
	ps.mu.Lock()
	s, ok := ps.peers[id]
	delete(ps.peers, id)
	ps.mu.Unlock()

	if ok {
		s.Close()
	}
}

func (ps *PeerSet) Send(msgType MsgType, payload any) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var lastErr error
	for id, s := range ps.peers {
		if err := s.Send(msgType, payload); err != nil {
			log.Printf("[peerset] send to %s failed: %v", id, err)
			lastErr = err
		}
	}
	return lastErr
}

func (ps *PeerSet) Recv() <-chan *Envelope {
	return ps.recvCh
}

func (ps *PeerSet) Close() {
	close(ps.stopCh)

	ps.mu.Lock()
	for id, s := range ps.peers {
		s.Close()
		delete(ps.peers, id)
	}
	ps.mu.Unlock()
}

func (ps *PeerSet) Count() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.peers)
}

func (ps *PeerSet) IDs() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	ids := make([]string, 0, len(ps.peers))
	for id := range ps.peers {
		ids = append(ids, id)
	}
	return ids
}

func (ps *PeerSet) pipeRecv(id string, s Sender) {
	for {
		select {
		case <-ps.stopCh:
			return
		case env, ok := <-s.Recv():
			if !ok {
				return
			}
			select {
			case ps.recvCh <- env:
			default:
				log.Printf("[peerset] recv channel full, dropping from %s", id)
			}
		}
	}
}
