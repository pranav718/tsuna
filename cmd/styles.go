package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	accentColor = lipgloss.Color("#00d4ff")
	pinkColor   = lipgloss.Color("#ff6b9d")
	greenColor  = lipgloss.Color("#00e676")
	dimColor    = lipgloss.Color("#666666")
	warnColor   = lipgloss.Color("#ffd600")
)

var (
	banner = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	accent = lipgloss.NewStyle().Foreground(accentColor)
	pink   = lipgloss.NewStyle().Foreground(pinkColor).Bold(true)
	green  = lipgloss.NewStyle().Foreground(greenColor)
	dim    = lipgloss.NewStyle().Foreground(dimColor)
	warn   = lipgloss.NewStyle().Foreground(warnColor)
)

func printBanner() {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(0, 2)

	title := banner.Render("TSUNA") + "  " + dim.Render("p2p synchronized video watching")
	fmt.Println(box.Render(title))
	fmt.Println()
}

func printStep(num int, msg string) {
	marker := accent.Render(fmt.Sprintf("  [%d]", num))
	fmt.Printf("%s %s\n", marker, msg)
}

func printStepDone(num int, msg string) {
	marker := green.Render(fmt.Sprintf("  [%d]", num))
	fmt.Printf("%s %s %s\n", marker, msg, green.Render("ok"))
}

func printRoomCode(code string) {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(pinkColor).
		Padding(0, 2).
		MarginLeft(4)

	inner := fmt.Sprintf("%s  %s", dim.Render("room code"), pink.Render(code))
	fmt.Println()
	fmt.Println(box.Render(inner))
	fmt.Println()
}

func printInfo(label, value string) {
	fmt.Printf("       %s %s\n", dim.Render(label), value)
}

func printWaiting(msg string) {
	fmt.Printf("\n  %s %s\n\n", accent.Render(".."), dim.Render(msg))
}

func printConnected(rtt string) {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(greenColor).
		Padding(0, 2).
		MarginLeft(4)

	inner := fmt.Sprintf("%s  %s", green.Render("connected"), dim.Render("rtt "+rtt))
	fmt.Println(box.Render(inner))
	fmt.Println()
}

func printError(msg string) {
	fmt.Printf("  %s %s\n", warn.Render("x"), msg)
}
