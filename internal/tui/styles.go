package tui

import "github.com/charmbracelet/lipgloss"

var (
	accent      = lipgloss.Color("#d8b4fe")
	accentBright = lipgloss.Color("#e9d8fd")
	command     = lipgloss.Color("#c084fc")
	success     = lipgloss.Color("#a7f3d0")
	warning     = lipgloss.Color("#fde047")
	danger      = lipgloss.Color("#fca5a5")
	dimGray     = lipgloss.Color("#9fa0b5")
	muted       = lipgloss.Color("#76778f")
	textColor   = lipgloss.Color("#e2e2ec")
	white       = lipgloss.Color("#ffffff")
	bg          = lipgloss.Color("#09090b")
	surface     = lipgloss.Color("#0f0f11")
	border      = lipgloss.Color("#27272a")
	borderBright = lipgloss.Color("#52525b")
)

var (
	cyan   = accent
	pink   = command
	green  = success
	yellow = warning
	red    = danger
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentBright).
			Padding(0, 1)

	RoomCodeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(command)

	BadgePlaying = lipgloss.NewStyle().
			Bold(true).
			Foreground(success).
			SetString("▸ playing")

	BadgePaused = lipgloss.NewStyle().
			Bold(true).
			Foreground(warning).
			SetString("‖ paused")

	BadgeHolding = lipgloss.NewStyle().
			Bold(true).
			Foreground(warning).
			SetString(".. holding")

	BadgeIdle = lipgloss.NewStyle().
			Bold(true).
			Foreground(muted).
			SetString("— idle")

	SectionTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			MarginTop(1)

	PeerOnline = lipgloss.NewStyle().
			Foreground(success).
			SetString("●")

	PeerBuffering = lipgloss.NewStyle().
			Foreground(warning).
			SetString("◐")

	PeerOffline = lipgloss.NewStyle().
			Foreground(danger).
			SetString("○")

	DimText = lipgloss.NewStyle().
		Foreground(dimGray)

	MutedText = lipgloss.NewStyle().
		Foreground(muted)

	ValueText = lipgloss.NewStyle().
			Foreground(textColor)

	BrightText = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	LogLine = lipgloss.NewStyle().
		Foreground(dimGray)

	BorderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			BorderBackground(bg).
			Padding(0, 1)

	FullScreen = lipgloss.NewStyle().
			Background(bg)
)
