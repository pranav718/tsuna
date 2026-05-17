package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type RelayTransport struct {
	conn     *websocket.Conn
	localID  string
	roomCode string
	seq      atomic.Uint64
	RecvCh   chan *Envelope
	stopCh   chan struct{}
	once     sync.Once
}

func NewRelayTransport(signalURL, roomCode, localID string) (*RelayTransport, error) {
	u, err := url.Parse(signalURL)
	if err != nil {
		return nil, fmt.Errorf("relay: bad url: %w", err)
	}

	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	u.Path = "/relay"
	q := u.Query()
	q.Set("room", roomCode)
	q.Set("peer", localID)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("relay: dial: %w", err)
	}

	return &RelayTransport{
		conn:     conn,
		localID:  localID,
		roomCode: roomCode,
		RecvCh:   make(chan *Envelope, 64),
		stopCh:   make(chan struct{}),
	}, nil
}

func (rt *RelayTransport) Start() {
	go rt.readLoop()
}

func (rt *RelayTransport) Send(msgType MsgType, payload any) error {
	seq := rt.seq.Add(1)
	data, err := Build(msgType, rt.localID, rt.roomCode, seq, payload)
	if err != nil {
		return fmt.Errorf("relay: build: %w", err)
	}

	return rt.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (rt *RelayTransport) readLoop() {
	for {
		select {
		case <-rt.stopCh:
			return
		default:
		}

		_, data, err := rt.conn.ReadMessage()
		if err != nil {
			select {
			case <-rt.stopCh:
			default:
				log.Printf("[relay] read error: %v", err)
			}
			return
		}

		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}

		if env.SenderID == rt.localID {
			continue
		}

		select {
		case rt.RecvCh <- &env:
		default:
		}
	}
}

func (rt *RelayTransport) Recv() <-chan *Envelope {
	return rt.RecvCh
}

func (rt *RelayTransport) Close() {
	rt.once.Do(func() {
		close(rt.stopCh)
		rt.conn.Close()
	})
}
