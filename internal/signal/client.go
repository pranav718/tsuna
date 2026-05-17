package signal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

func Register(serverURL string, req RegisterRequest) (*RegisterResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("signal: marshal register: %w", err)
	}

	resp, err := httpClient.Post(serverURL+"/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("signal: register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signal: register: status %d", resp.StatusCode)
	}

	var out RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("signal: decode register response: %w", err)
	}
	return &out, nil
}

func GetPeers(serverURL, roomCode string) ([]PeerInfo, error) {
	resp, err := httpClient.Get(fmt.Sprintf("%s/peers?room=%s", serverURL, roomCode))
	if err != nil {
		return nil, fmt.Errorf("signal: get peers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("signal: room not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signal: get peers: status %d", resp.StatusCode)
	}

	var peers []PeerInfo
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return nil, fmt.Errorf("signal: decode peers: %w", err)
	}
	return peers, nil
}

func Leave(serverURL, roomCode, peerID string) error {
	body, _ := json.Marshal(map[string]string{
		"room_code": roomCode,
		"peer_id":   peerID,
	})

	resp, err := httpClient.Post(serverURL+"/leave", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("signal: leave: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func GetRooms(serverURL string) ([]RoomSummary, error) {
	resp, err := httpClient.Get(serverURL + "/rooms")
	if err != nil {
		return nil, fmt.Errorf("signal: get rooms: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signal: get rooms: status %d", resp.StatusCode)
	}

	var rooms []RoomSummary
	if err := json.NewDecoder(resp.Body).Decode(&rooms); err != nil {
		return nil, fmt.Errorf("signal: decode rooms: %w", err)
	}
	return rooms, nil
}
