package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pranav718/tsuna/internal/tui"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Bridge struct {
	addr    string
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
	eventCh <-chan tui.UIEvent
	cmdCh   chan Command
}

type Command struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func NewBridge(addr string, eventCh <-chan tui.UIEvent) *Bridge {
	return &Bridge{
		addr:    addr,
		clients: make(map[*websocket.Conn]bool),
		eventCh: eventCh,
		cmdCh:   make(chan Command, 32),
	}
}

func (b *Bridge) Commands() <-chan Command {
	return b.cmdCh
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (b *Bridge) Start() {
	go b.eventLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", b.handleWS)

	go func() {
		log.Printf("[web] bridge listening on %s", b.addr)
		if err := http.ListenAndServe(b.addr, mux); err != nil {
			log.Printf("[web] bridge error: %v", err)
		}
	}()
}

func (b *Bridge) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[web] upgrade error: %v", err)
		return
	}

	b.mu.Lock()
	b.clients[conn] = true
	b.mu.Unlock()

	log.Printf("[web] client connected (%d total)", len(b.clients))

	go b.readClient(conn)
}

func (b *Bridge) readClient(conn *websocket.Conn) {
	defer func() {
		b.mu.Lock()
		delete(b.clients, conn)
		b.mu.Unlock()
		conn.Close()
		log.Printf("[web] client disconnected (%d total)", len(b.clients))
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var cmd Command
		if err := json.Unmarshal(data, &cmd); err != nil {
			continue
		}

		select {
		case b.cmdCh <- cmd:
		default:
		}
	}
}

func (b *Bridge) eventLoop() {
	for ev := range b.eventCh {
		webEv := convertEvent(ev)
		if webEv == nil {
			continue
		}

		data, err := json.Marshal(webEv)
		if err != nil {
			continue
		}

		b.mu.RLock()
		for conn := range b.clients {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				conn.Close()
				delete(b.clients, conn)
			}
		}
		b.mu.RUnlock()
	}
}

func convertEvent(ev tui.UIEvent) *Event {
	switch ev.Type {
	case tui.UIPeerHello:
		return &Event{Type: "peer_hello", Data: ev.Data}
	case tui.UIPeerBye:
		return &Event{Type: "peer_bye", Data: ev.Data}
	case tui.UIClockSync:
		return &Event{Type: "clock_sync", Data: ev.Data}
	case tui.UIStateUpdate:
		return &Event{Type: "state_update", Data: ev.Data}
	case tui.UICorrection:
		return &Event{Type: "correction", Data: ev.Data}
	case tui.UIBufferingStart:
		return &Event{Type: "buffering_start", Data: nil}
	case tui.UIBufferingStop:
		return &Event{Type: "buffering_stop", Data: nil}
	case tui.UIQueueUpdate:
		return &Event{Type: "queue_update", Data: ev.Data}
	case tui.UILog:
		return &Event{Type: "log", Data: ev.Data}
	default:
		return nil
	}
}

func (b *Bridge) Broadcast(ev Event) {
	data, err := json.Marshal(ev)
	if err != nil {
		return
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for conn := range b.clients {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}
