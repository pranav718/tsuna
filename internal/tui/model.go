package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxLogs = 12

type peerState struct {
	name      string
	online    bool
	rtt       time.Duration
	offset    time.Duration
	syncDelta time.Duration
	buffering bool
	isHost    bool
}

type Model struct {
	roomCode    string
	isHost      bool
	localID     string
	peers       map[string]*peerState
	localPos    time.Duration
	localPaused bool
	roomState   string
	logs        []string
	width       int
	height      int
	eventCh     <-chan UIEvent
	cancelFunc  func()

	queueItems   []QueueItemData
	queueCurrent int
}

func NewModel(roomCode, localID, remoteID string, isHost bool, eventCh <-chan UIEvent, cancel func()) Model {
	peers := make(map[string]*peerState)
	if remoteID != "" {
		peers[remoteID] = &peerState{name: remoteID, online: true, isHost: !isHost}
	}

	return Model{
		roomCode:   roomCode,
		isHost:     isHost,
		localID:    localID,
		peers:      peers,
		roomState:  "IDLE",
		logs:       make([]string, 0, maxLogs),
		eventCh:    eventCh,
		cancelFunc: cancel,
	}
}

func (m Model) Init() tea.Cmd {
	return waitForEvent(m.eventCh)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.cancelFunc()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case UIEvent:
		m.handleEvent(msg)
		return m, waitForEvent(m.eventCh)
	}

	return m, nil
}

func (m *Model) handleEvent(ev UIEvent) {
	switch ev.Type {
	case UIPeerHello:
		if d, ok := ev.Data.(PeerData); ok {
			if p, exists := m.peers[d.PeerID]; exists {
				p.name = d.DisplayName
				p.online = true
			} else {
				m.peers[d.PeerID] = &peerState{name: d.DisplayName, online: true}
			}
			m.addLog("peer connected: " + d.DisplayName)
		}

	case UIPeerBye:
		if d, ok := ev.Data.(PeerData); ok {
			if p, exists := m.peers[d.PeerID]; exists {
				p.online = false
			}
		}
		m.addLog("peer disconnected")

	case UIClockSync:
		if d, ok := ev.Data.(ClockSyncData); ok {
			if p, exists := m.peers[d.PeerID]; exists {
				p.rtt = d.RTT
				p.offset = d.Offset
			}
		}

	case UIStateUpdate:
		if d, ok := ev.Data.(StateData); ok {
			m.localPos = d.Position
			m.localPaused = d.Paused
			if d.Paused {
				m.roomState = "PAUSED"
			} else {
				m.roomState = "PLAYING"
			}
		}

	case UICorrection:
		if d, ok := ev.Data.(CorrectionData); ok {
			m.addLog(fmt.Sprintf("%s: Δ%v → %v", d.CorrType, d.Delta, d.Target))
		}

	case UIBufferingStart:
		m.roomState = "HOLDING"
		m.addLog("buffering detected. holding peers")

	case UIBufferingStop:
		m.roomState = "PLAYING"
		m.addLog("buffering resolved :D , resuming")

	case UILog:
		if s, ok := ev.Data.(string); ok {
			m.addLog(s)
		}

	case UIQueueUpdate:
		if d, ok := ev.Data.(QueueData); ok {
			m.queueItems = d.Items
			m.queueCurrent = d.Current
		}
	}
}

func (m *Model) addLog(msg string) {
	ts := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("%s  %s", DimText.Render(ts), msg)
	m.logs = append(m.logs, entry)
	if len(m.logs) > maxLogs {
		m.logs = m.logs[len(m.logs)-maxLogs:]
	}
}

func (m Model) View() string {
	var b strings.Builder

	header := m.renderHeader()
	peers := m.renderPeers()
	playback := m.renderPlayback()
	q := m.renderQueue()
	logs := m.renderLogs()

	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(peers)
	b.WriteString("\n")
	b.WriteString(playback)
	if q != "" {
		b.WriteString("\n")
		b.WriteString(q)
	}
	b.WriteString("\n")
	b.WriteString(logs)
	b.WriteString("\n\n")
	b.WriteString(DimText.Render("  press q to quit"))

	return b.String()
}

func (m Model) renderHeader() string {
	badge := BadgeIdle
	switch m.roomState {
	case "PLAYING":
		badge = BadgePlaying
	case "PAUSED":
		badge = BadgePaused
	case "HOLDING":
		badge = BadgeHolding
	}

	left := HeaderStyle.Render("tsuna")
	code := RoomCodeStyle.Render(m.roomCode)
	line := fmt.Sprintf("  %s    %s   %s", left, code, badge.String())

	w := m.width
	if w < 50 {
		w = 50
	}
	return BorderBox.Width(w - 4).Render(line)
}

func (m Model) renderPeers() string {
	title := SectionTitle.Render("  peers")

	role := "host"
	if !m.isHost {
		role = "peer"
	}
	local := fmt.Sprintf("  %s %s  %s",
		PeerOnline.String(),
		ValueText.Render(truncate(m.localID, 20)),
		DimText.Render("(you, "+role+")"),
	)

	lines := []string{title, local}

	for id, p := range m.peers {
		status := PeerOnline
		meta := ""
		if !p.online {
			status = PeerOffline
			meta = DimText.Render("disconnected")
		} else {
			parts := []string{}
			if p.rtt > 0 {
				parts = append(parts, fmt.Sprintf("rtt %v", p.rtt.Round(time.Millisecond)))
			}
			if p.syncDelta != 0 {
				parts = append(parts, fmt.Sprintf("Δ %+dms", p.syncDelta.Milliseconds()))
			}
			if p.buffering {
				parts = append(parts, lipgloss.NewStyle().Foreground(yellow).Render("buffering"))
			}
			meta = DimText.Render(strings.Join(parts, "  "))
		}

		name := p.name
		if name == "" {
			name = id
		}
		peerRole := "peer"
		if p.isHost {
			peerRole = "host"
		}
		line := fmt.Sprintf("  %s %s  %s  %s",
			status.String(),
			ValueText.Render(truncate(name, 20)),
			DimText.Render("("+peerRole+")"),
			meta,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderPlayback() string {
	title := SectionTitle.Render("  playback")
	icon := ">"
	if m.localPaused {
		icon = "||"
	}
	pos := formatDuration(m.localPos)
	line := fmt.Sprintf("  %s  %s", icon, ValueText.Render(pos))
	return fmt.Sprintf("%s\n%s", title, line)
}

func (m Model) renderLogs() string {
	title := SectionTitle.Render("  log")
	if len(m.logs) == 0 {
		return fmt.Sprintf("%s\n%s", title, DimText.Render("  waiting for events..."))
	}

	lines := make([]string, len(m.logs))
	for i, l := range m.logs {
		lines[i] = "  " + l
	}
	return fmt.Sprintf("%s\n%s", title, strings.Join(lines, "\n"))
}

func waitForEvent(ch <-chan UIEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return tea.Quit()
		}
		return ev
	}
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "."
}

func (m Model) renderQueue() string {
	if len(m.queueItems) == 0 {
		return ""
	}

	title := SectionTitle.Render("  queue")
	lines := make([]string, len(m.queueItems))
	for i, item := range m.queueItems {
		marker := "  "
		if i == m.queueCurrent {
			marker = "> "
		}

		name := item.Filename
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		name = truncate(name, 40)

		lines[i] = fmt.Sprintf("  %s%s  %s",
			marker,
			ValueText.Render(name),
			DimText.Render("("+item.AddedBy+")"),
		)
	}
	return fmt.Sprintf("%s\n%s", title, strings.Join(lines, "\n"))
}
