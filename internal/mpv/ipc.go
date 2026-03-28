package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultSocketPath = "/tmp/tsuna-mpv.sock"

const PollInterval = 100 * time.Millisecond

type command struct {
	Command   []any  `json:"command"`
	RequestID uint64 `json:"request_id"`
}

type response struct {
	RequestID uint64 `json:"request_id"`
	Error     string `json:"error"`
	Data      any    `json:"data"`
}

type PlaybackState struct {
	Position time.Duration
	Paused   bool
	Sampled  time.Time
}

type Bridge struct {
	socketPath string

	conn   net.Conn
	writer *bufio.Writer
	reader *bufio.Scanner
	mu     sync.Mutex

	nextID  uint64
	pending map[uint64]chan response
	state   atomic.Pointer[PlaybackState]
	OnEvent func(event string, data map[string]any) 

	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewBridge(socketPath string) *Bridge {
	b := &Bridge{
		socketPath: socketPath,
		pending:    make(map[uint64]chan response),
		stopCh:     make(chan struct{}),
	}
	b.state.Store(&PlaybackState{})
	return b
}

func (b *Bridge) Connect() error {
	conn, err := net.Dial("unix", b.socketPath)
	if err != nil {
		return fmt.Errorf("mpv: dial %s: %w", b.socketPath, err)
	}

	b.mu.Lock()
	b.conn = conn
	b.writer = bufio.NewWriter(conn)
	b.reader = bufio.NewScanner(conn)
	b.mu.Unlock()

	b.wg.Add(2)
	go b.readLoop()
	go b.pollLoop()

	return nil
}

func (b *Bridge) Close() error {
	close(b.stopCh)
	b.wg.Wait()

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *Bridge) State() PlaybackState {
	if s := b.state.Load(); s != nil {
		return *s
	}
	return PlaybackState{}
}

func (b *Bridge) Play() error {
	return b.setProperty("pause", false)
}

func (b *Bridge) Pause() error {
	return b.setProperty("pause", true)
}

func (b *Bridge) Seek(pos time.Duration) error {
	seconds := pos.Seconds()
	_, err := b.sendCommand("seek", seconds, "absolute")
	return err
}

func (b *Bridge) SeekExact(pos time.Duration) error {
	seconds := pos.Seconds()
	_, err := b.sendCommand("seek", seconds, "absolute+exact")
	return err
}

func (b *Bridge) GetPosition() (time.Duration, error) {
	resp, err := b.sendCommand("get_property", "time-pos")
	if err != nil {
		return 0, err
	}
	seconds, ok := resp.Data.(float64)
	if !ok {
		return 0, fmt.Errorf("mpv: unexpected time-pos type %T", resp.Data)
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

func (b *Bridge) IsPaused() (bool, error) {
	resp, err := b.sendCommand("get_property", "pause")
	if err != nil {
		return false, err
	}
	paused, ok := resp.Data.(bool)
	if !ok {
		return false, fmt.Errorf("mpv: unexpected pause type %T", resp.Data)
	}
	return paused, nil
}

func (b *Bridge) LoadFile(path string) error {
	_, err := b.sendCommand("loadfile", path, "replace")
	return err
}

func (b *Bridge) sendCommand(args ...any) (response, error) {
	id := atomic.AddUint64(&b.nextID, 1)
	cmd := command{Command: args, RequestID: id}

	data, err := json.Marshal(cmd)
	if err != nil {
		return response{}, fmt.Errorf("mpv: marshal command: %w", err)
	}
	data = append(data, '\n')

	replyCh := make(chan response, 1)

	b.mu.Lock()
	b.pending[id] = replyCh
	_, err = b.writer.Write(data)
	if err == nil {
		err = b.writer.Flush()
	}
	b.mu.Unlock()

	if err != nil {
		b.mu.Lock()
		delete(b.pending, id)
		b.mu.Unlock()
		return response{}, fmt.Errorf("mpv: write command: %w", err)
	}

	select {
	case resp := <-replyCh:
		if resp.Error != "" && resp.Error != "success" {
			return resp, fmt.Errorf("mpv: command error: %s", resp.Error)
		}
		return resp, nil
	case <-time.After(3 * time.Second):
		b.mu.Lock()
		delete(b.pending, id)
		b.mu.Unlock()
		return response{}, fmt.Errorf("mpv: command %v timed out", args)
	case <-b.stopCh:
		return response{}, fmt.Errorf("mpv: bridge closed")
	}
}

func (b *Bridge) setProperty(name string, value any) error {
	_, err := b.sendCommand("set_property", name, value)
	return err
}

func (b *Bridge) readLoop() {
	defer b.wg.Done()

	for {
		select {
		case <-b.stopCh:
			return
		default:
		}

		if !b.reader.Scan() {
			return
		}

		line := b.reader.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp response
		if err := json.Unmarshal(line, &resp); err == nil && resp.RequestID != 0 {
			b.mu.Lock()
			ch, ok := b.pending[resp.RequestID]
			if ok {
				delete(b.pending, resp.RequestID)
			}
			b.mu.Unlock()
			if ok {
				ch <- resp
			}
			continue
		}

		if b.OnEvent != nil {
			var ev map[string]any
			if json.Unmarshal(line, &ev) == nil {
				if name, ok := ev["event"].(string); ok {
					b.OnEvent(name, ev)
				}
			}
		}
	}
}

func (b *Bridge) pollLoop() {
	defer b.wg.Done()

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			pos, err := b.GetPosition()
			if err != nil {
				continue
			}
			paused, err := b.IsPaused()
			if err != nil {
				continue
			}
			b.state.Store(&PlaybackState{
				Position: pos,
				Paused:   paused,
				Sampled:  time.Now(),
			})
		}
	}
}
