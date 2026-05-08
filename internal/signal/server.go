package signal

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type PeerInfo struct {
	PeerID     string    `json:"peer_id"`
	PublicIP   string    `json:"public_ip"`
	PublicPort int       `json:"public_port"`
	JoinedAt   time.Time `json:"joined_at"`
}

type Room struct {
	Code      string              `json:"code"`
	Peers     map[string]PeerInfo `json:"peers"`
	CreatedAt time.Time           `json:"created_at"`
	mu        sync.RWMutex
}

type Server struct {
	rooms map[string]*Room
	mu    sync.RWMutex
	addr  string
}

type RegisterRequest struct {
	RoomCode   string `json:"room_code"`
	PeerID     string `json:"peer_id"`
	PublicIP   string `json:"public_ip"`
	PublicPort int    `json:"public_port"`
}

type RegisterResponse struct {
	OK    bool       `json:"ok"`
	Peers []PeerInfo `json:"peers"`
}

func NewServer(addr string) *Server {
	return &Server{
		rooms: make(map[string]*Room),
		addr:  addr,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", s.handleRegister)
	mux.HandleFunc("/peers", s.handlePeers)
	mux.HandleFunc("/leave", s.handleLeave)
	mux.HandleFunc("/health", s.handleHealth)

	log.Printf("[signal] server listening on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	room, exists := s.rooms[req.RoomCode]
	if !exists {
		room = &Room{
			Code:      req.RoomCode,
			Peers:     make(map[string]PeerInfo),
			CreatedAt: time.Now(),
		}
		s.rooms[req.RoomCode] = room
		log.Printf("[signal] room created: %s", req.RoomCode)
	}
	s.mu.Unlock()

	room.mu.Lock()
	room.Peers[req.PeerID] = PeerInfo{
		PeerID:     req.PeerID,
		PublicIP:   req.PublicIP,
		PublicPort: req.PublicPort,
		JoinedAt:   time.Now(),
	}
	peers := peersSlice(room.Peers)
	room.mu.Unlock()

	log.Printf("[signal] peer %s registered in room %s (%d peers)", req.PeerID, req.RoomCode, len(peers))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RegisterResponse{OK: true, Peers: peers})
}

func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("room")
	if code == "" {
		http.Error(w, "missing room param", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	room, exists := s.rooms[code]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	room.mu.RLock()
	peers := peersSlice(room.Peers)
	room.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peers)
}

func (s *Server) handleLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoomCode string `json:"room_code"`
		PeerID   string `json:"peer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[req.RoomCode]
	if !exists {
		w.WriteHeader(http.StatusOK)
		return
	}

	room.mu.Lock()
	delete(room.Peers, req.PeerID)
	remaining := len(room.Peers)
	room.mu.Unlock()

	if remaining == 0 {
		delete(s.rooms, req.RoomCode)
		log.Printf("[signal] room %s dissolved (no peers remaining)", req.RoomCode)
	}

	log.Printf("[signal] peer %s left room %s", req.PeerID, req.RoomCode)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func peersSlice(m map[string]PeerInfo) []PeerInfo {
	out := make([]PeerInfo, 0, len(m))
	for _, p := range m {
		out = append(out, p)
	}
	return out
}
