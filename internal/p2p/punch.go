package p2p

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type PunchState int

const (
	StateIdle        PunchState = iota
	StateProbing               
	StateEstablished            
	StateFailed                 
)

func (s PunchState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateProbing:
		return "probing"
	case StateEstablished:
		return "established"
	case StateFailed:
		return "failed (relay fallback)"
	default:
		return "unknown"
	}
}

type PeerEndpoint struct {
	IP   net.IP
	Port int
}

func (e PeerEndpoint) String() string {
	return fmt.Sprintf("%s:%d", e.IP, e.Port)
}

func (e PeerEndpoint) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{IP: e.IP, Port: e.Port}
}

type PunchConfig struct {
	MaxAttempts int
	ProbeInterval time.Duration
	Timeout time.Duration
	LocalPort int
}

func DefaultPunchConfig() PunchConfig {
	return PunchConfig{
		MaxAttempts:   20,
		ProbeInterval: 150 * time.Millisecond,
		Timeout:       10 * time.Second,
		LocalPort:     0,
	}
}

type PunchResult struct {
	State    PunchState
	Conn     *net.UDPConn  
	RTT      time.Duration 
	Endpoint PeerEndpoint 
}

type Puncher struct {
	cfg      PunchConfig
	local    *net.UDPConn
	remote   PeerEndpoint
	state    PunchState
	mu       sync.Mutex
	resultCh chan PunchResult
	stopCh   chan struct{}
}

func NewPuncher(remote PeerEndpoint, cfg PunchConfig) (*Puncher, error) {
	addr := &net.UDPAddr{Port: cfg.LocalPort}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("p2p: bind local UDP: %w", err)
	}

	return &Puncher{
		cfg:      cfg,
		local:    conn,
		remote:   remote,
		state:    StateIdle,
		resultCh: make(chan PunchResult, 1),
		stopCh:   make(chan struct{}),
	}, nil
}

func (p *Puncher) Punch() (PunchResult, error) {
	p.setState(StateProbing)

	go p.sendProbes()
	go p.listenForProbe()

	select {
	case result := <-p.resultCh:
		return result, nil
	case <-time.After(p.cfg.Timeout):
		p.setState(StateFailed)
		return PunchResult{State: StateFailed}, fmt.Errorf("p2p: punch timed out after %s", p.cfg.Timeout)
	case <-p.stopCh:
		return PunchResult{State: StateFailed}, fmt.Errorf("p2p: puncher closed")
	}
}

func (p *Puncher) sendProbes() {
	probe := []byte("tsuna:punch:probe")

	for i := 0; i < p.cfg.MaxAttempts; i++ {
		select {
		case <-p.stopCh:
			return
		default:
		}

		_, err := p.local.WriteToUDP(probe, p.remote.UDPAddr())
		if err != nil {
			_ = err
		}

		time.Sleep(p.cfg.ProbeInterval)
	}
}

func (p *Puncher) listenForProbe() {
	buf := make([]byte, 256)

	for {
		select {
		case <-p.stopCh:
			return
		default:
		}

		_ = p.local.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, addr, err := p.local.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := string(buf[:n])
		if msg != "tsuna:punch:probe" || addr == nil {
			continue
		}

		start := time.Now()
		_, _ = p.local.WriteToUDP([]byte("tsuna:punch:ack"), addr)
		rtt := time.Since(start)

		p.setState(StateEstablished)
		p.resultCh <- PunchResult{
			State:    StateEstablished,
			Conn:     p.local,
			RTT:      rtt,
			Endpoint: p.remote,
		}
		return
	}
}

func (p *Puncher) setState(s PunchState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = s
}

func (p *Puncher) State() PunchState {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *Puncher) Close() error {
	close(p.stopCh)
	return p.local.Close()
}
