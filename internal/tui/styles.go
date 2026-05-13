package tui

import "github.com/charmbracelet/lipgloss"

var (
	cyan    = lipgloss.Color("#00d4ff")
	pink    = lipgloss.Color("#ff6b9d")
	green   = lipgloss.Color("#00e676")
	yellow  = lipgloss.Color("#ffd600")
	red     = lipgloss.Color("#ff5252")
	dimGray = lipgloss.Color("#666666")
	white   = lipgloss.Color("#e0e0e0")
	bg      = lipgloss.Color("#1a1a2e")
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			Padding(0, 1)

	RoomCodeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(pink)

	BadgePlaying = lipgloss.NewStyle().
			Bold(true).
			Foreground(green).
			SetString("> PLAYING")

	BadgePaused = lipgloss.NewStyle().
			Bold(true).
			Foreground(yellow).
			SetString("|| PAUSED")

	BadgeHolding = lipgloss.NewStyle().
			Bold(true).
			Foreground(yellow).
			SetString(".. HOLDING")

	BadgeIdle = lipgloss.NewStyle().
			Bold(true).
			Foreground(dimGray).
			SetString("-- IDLE")

	SectionTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			MarginTop(1)

	PeerOnline = lipgloss.NewStyle().
			Foreground(green).
			SetString("*")

	PeerBuffering = lipgloss.NewStyle().
			Foreground(yellow).
			SetString("*")

	PeerOffline = lipgloss.NewStyle().
			Foreground(red).
			SetString("*")

	DimText = lipgloss.NewStyle().
		Foreground(dimGray)

	ValueText = lipgloss.NewStyle().
			Foreground(white)

	LogLine = lipgloss.NewStyle().
		Foreground(dimGray)

	BorderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cyan).
			Padding(0, 1)
)
