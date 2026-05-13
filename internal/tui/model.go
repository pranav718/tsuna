package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxLogs = 12

type Model struct {
	roomCode    string
	isHost      bool
	localID     string
	remoteID    string
	remoteName  string
	peerOnline  bool
	rtt         time.Duration
	offset      time.Duration
	syncDelta   time.Duration
	localPos    time.Duration
	localPaused bool
	roomState   string
	buffering   bool
	logs        []string
	width       int
	height      int
	eventCh     <-chan UIEvent
	cancelFunc  func()
}

func NewModel(roomCode, localID, remoteID string, isHost bool, eventCh <-chan UIEvent, cancel func()) Model {
	return Model{
		roomCode:   roomCode,
		isHost:     isHost,
		localID:    localID,
		remoteID:   remoteID,
		peerOnline: true,
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
			m.remoteName = d.DisplayName
			m.peerOnline = true
			m.addLog("peer connected: " + d.DisplayName)
		}

	case UIPeerBye:
		m.peerOnline = false
		m.addLog("peer disconnected")

	case UIClockSync:
		if d, ok := ev.Data.(ClockSyncData); ok {
			m.rtt = d.RTT
			m.offset = d.Offset
		}

	case UIStateUpdate:
		if d, ok := ev.Data.(StateData); ok {
			m.localPos = d.Position
			m.localPaused = d.Paused
			m.syncDelta = d.SyncDelta
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
		m.buffering = true
		m.roomState = "HOLDING"
		m.addLog("buffering detected. holding peers")

	case UIBufferingStop:
		m.buffering = false
		m.addLog("buffering resolved :D , resuming")

	case UILog:
		if s, ok := ev.Data.(string); ok {
			m.addLog(s)
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
	logs := m.renderLogs()

	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(peers)
	b.WriteString("\n")
	b.WriteString(playback)
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

	left := HeaderStyle.Render("津波 TSUNA")
	code := RoomCodeStyle.Render(m.roomCode)
	line := fmt.Sprintf("  %s    %s   %s", left, code, badge.String())

	w := m.width
	if w < 50 {
		w = 50
	}
	return BorderBox.Width(w - 4).Render(line)
}

func (m Model) renderPeers() string {
	title := SectionTitle.Render("  PEERS")

	role := "host"
	if !m.isHost {
		role = "peer"
	}
	local := fmt.Sprintf("  %s %s  %s",
		PeerOnline.String(),
		ValueText.Render(truncate(m.localID, 20)),
		DimText.Render("(you, "+role+")"),
	)

	remoteStatus := PeerOnline
	remoteMeta := ""
	if !m.peerOnline {
		remoteStatus = PeerOffline
		remoteMeta = DimText.Render("disconnected")
	} else {
		parts := []string{}
		if m.rtt > 0 {
			parts = append(parts, fmt.Sprintf("rtt %v", m.rtt.Round(time.Millisecond)))
		}
		if m.syncDelta != 0 {
			parts = append(parts, fmt.Sprintf("Δ %+dms", m.syncDelta.Milliseconds()))
		}
		if m.buffering {
			parts = append(parts, lipgloss.NewStyle().Foreground(yellow).Render("buffering"))
		}
		remoteMeta = DimText.Render(strings.Join(parts, "  "))
	}

	name := m.remoteName
	if name == "" {
		name = m.remoteID
	}
	remoteRole := "peer"
	if m.isHost {
		remoteRole = "peer"
	} else {
		remoteRole = "host"
	}
	remote := fmt.Sprintf("  %s %s  %s  %s",
		remoteStatus.String(),
		ValueText.Render(truncate(name, 20)),
		DimText.Render("("+remoteRole+")"),
		remoteMeta,
	)

	return fmt.Sprintf("%s\n%s\n%s", title, local, remote)
}

func (m Model) renderPlayback() string {
	title := SectionTitle.Render("  PLAYBACK")
	icon := ">"
	if m.localPaused {
		icon = "||"
	}
	pos := formatDuration(m.localPos)
	line := fmt.Sprintf("  %s  %s", icon, ValueText.Render(pos))
	return fmt.Sprintf("%s\n%s", title, line)
}

func (m Model) renderLogs() string {
	title := SectionTitle.Render("  LOG")
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
