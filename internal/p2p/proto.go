package p2p

import (
	"encoding/json"
	"fmt"
	"time"
)

type MsgType string

const (
	MsgHello     MsgType = "hello"
	MsgHeartbeat MsgType = "heartbeat"
	MsgBye       MsgType = "bye"

	MsgPing MsgType = "ping"
	MsgPong MsgType = "pong"

	MsgPlay  MsgType = "play"
	MsgPause MsgType = "pause"
	MsgSeek  MsgType = "seek"

	MsgStateUpdate MsgType = "state_update"
	MsgHold        MsgType = "hold"
	MsgResume      MsgType = "resume"

	MsgPeerList    MsgType = "peer_list"
	MsgQueueAdd    MsgType = "queue_add"
	MsgQueueRemove MsgType = "queue_remove"
	MsgQueueNext   MsgType = "queue_next"

	MsgChat     MsgType = "chat"
	MsgReaction MsgType = "reaction"
)

type Envelope struct {
	Type     MsgType         `json:"type"`
	SenderID string          `json:"sender_id"`
	RoomCode string          `json:"room_code"`
	Seq      uint64          `json:"seq"`
	SentAt   time.Time       `json:"sent_at"`
	Payload  json.RawMessage `json:"payload,omitempty"`
}

func (e *Envelope) Encode() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("proto: encode envelope: %w", err)
	}
	return data, nil
}

func DecodeEnvelope(data []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("proto: decode envelope: %w", err)
	}
	return &env, nil
}

func DecodePayload(env *Envelope, dst any) error {
	if err := json.Unmarshal(env.Payload, dst); err != nil {
		return fmt.Errorf("proto: decode payload for %s: %w", env.Type, err)
	}
	return nil
}

func EncodePayload(env *Envelope, src any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("proto: encode payload for %s: %w", env.Type, err)
	}
	env.Payload = data
	return nil
}

type HelloPayload struct {
	DisplayName string `json:"display_name"`
	Version     string `json:"version"`
	IsHost      bool   `json:"is_host"`
}

type HeartbeatPayload struct {
}

type ByePayload struct {
	Reason string `json:"reason,omitempty"`
}

type PingPayload struct {
	T1 time.Time `json:"t1"`
}

type PongPayload struct {
	T1 time.Time `json:"t1"`
	T2 time.Time `json:"t2"`
	T3 time.Time `json:"t3"`
}

type PlayPayload struct {
	Position float64   `json:"position"`
	At       time.Time `json:"at"`
}

type PausePayload struct {
	Position float64 `json:"position"`
}

type SeekPayload struct {
	Position float64 `json:"position"`
}

type StateUpdatePayload struct {
	Position  float64 `json:"position"`
	Paused    bool    `json:"paused"`
	Buffering bool    `json:"buffering"`

	SyncDelta int64 `json:"sync_delta_ms"`
}

type HoldPayload struct {
	Position float64 `json:"position"`
	Reason   string  `json:"reason,omitempty"`
}

type ResumePayload struct {
	Position float64 `json:"position"`
}

type PeerEntry struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	PublicAddr  string `json:"public_addr"`
	IsHost      bool   `json:"is_host"`
}

type PeerListPayload struct {
	Peers []PeerEntry `json:"peers"`
}

type QueueItem struct {
	ID       string  `json:"id"`
	Filename string  `json:"filename"`
	Duration float64 `json:"duration"`
	AddedBy  string  `json:"added_by"`
}

type QueueAddPayload struct {
	Item QueueItem `json:"item"`
}

type QueueRemovePayload struct {
	ItemID string `json:"item_id"`
}

type QueueNextPayload struct {
	NextItemID string `json:"next_item_id"`
}

type ChatPayload struct {
	Text string `json:"text"`
}

type ReactionPayload struct {
	Emoji string `json:"emoji"`
}

func NewEnvelope(msgType MsgType, senderID, roomCode string, seq uint64) *Envelope {
	return &Envelope{
		Type:     msgType,
		SenderID: senderID,
		RoomCode: roomCode,
		Seq:      seq,
		SentAt:   time.Now(),
	}
}

func Build(msgType MsgType, senderID, roomCode string, seq uint64, payload any) ([]byte, error) {
	env := NewEnvelope(msgType, senderID, roomCode, seq)
	if payload != nil {
		if err := EncodePayload(env, payload); err != nil {
			return nil, err
		}
	}
	return env.Encode()
}
