package p2p

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

const maxUDPPacket = 65507

type Transport struct {
	conn     *net.UDPConn
	peerAddr *net.UDPAddr
	localID  string
	roomCode string
	seq      atomic.Uint64
	RecvCh   chan *Envelope
	stopCh   chan struct{}
}

func NewTransport(conn *net.UDPConn, peerAddr *net.UDPAddr, localID, roomCode string) *Transport {
	return &Transport{
		conn:     conn,
		peerAddr: peerAddr,
		localID:  localID,
		roomCode: roomCode,
		RecvCh:   make(chan *Envelope, 64),
		stopCh:   make(chan struct{}),
	}
}

func (t *Transport) Start() {
	go t.readLoop()
}

func (t *Transport) Send(msgType MsgType, payload any) error {
	seq := t.seq.Add(1)
	data, err := Build(msgType, t.localID, t.roomCode, seq, payload)
	if err != nil {
		return fmt.Errorf("transport: build message: %w", err)
	}

	_, err = t.conn.WriteToUDP(data, t.peerAddr)
	if err != nil {
		return fmt.Errorf("transport: send: %w", err)
	}
	return nil
}

func (t *Transport) Recv() <-chan *Envelope {
	return t.RecvCh
}

func (t *Transport) Close() {
	close(t.stopCh)
	t.conn.Close()
}

func (t *Transport) readLoop() {
	buf := make([]byte, maxUDPPacket)
	for {
		select {
		case <-t.stopCh:
			return
		default:
		}

		n, _, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-t.stopCh:
				return
			default:
				continue
			}
		}

		env, err := DecodeEnvelope(buf[:n])
		if err != nil {
			continue
		}

		select {
		case t.RecvCh <- env:
		default:
			log.Printf("[transport] recv channel full, dropping %s", env.Type)
		}
	}
}
