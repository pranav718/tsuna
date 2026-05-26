package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	accentColor = lipgloss.Color("#d8b4fe") 
	pinkColor   = lipgloss.Color("#c084fc")
	greenColor  = lipgloss.Color("#a7f3d0")
	dimColor    = lipgloss.Color("#9fa0b5")
	warnColor   = lipgloss.Color("#fde047")
)

var (
	bannerStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	accent      = lipgloss.NewStyle().Foreground(accentColor)
	pink        = lipgloss.NewStyle().Foreground(pinkColor).Bold(true)
	green       = lipgloss.NewStyle().Foreground(greenColor)
	dim         = lipgloss.NewStyle().Foreground(dimColor)
	warn        = lipgloss.NewStyle().Foreground(warnColor)
)

var spinFrames = []string{"-", "\\", "|", "/"}

func printBanner() {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(0, 2)

	title := bannerStyle.Render("TSUNA") + "  " + dim.Render("p2p synchronized video watching")
	fmt.Println(box.Render(title))
	fmt.Println()
}

func runWithSpinner(num int, label string, fn func() error) error {
	marker := accent.Render(fmt.Sprintf("  [%d]", num))

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	frame := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-done:
			fmt.Fprintf(os.Stdout, "\r\033[K")
			if err != nil {
				fmt.Printf("%s %s %s\n", marker, label, warn.Render("failed"))
			}
			return err
		case <-ticker.C:
			spin := accent.Render(spinFrames[frame%len(spinFrames)])
			fmt.Fprintf(os.Stdout, "\r%s %s %s", marker, label, spin)
			frame++
		}
	}
}

func printStepDone(num int, msg string) {
	marker := green.Render(fmt.Sprintf("  [%d]", num))
	fmt.Printf("%s %s %s\n", marker, msg, green.Render("ok"))
}

func typewrite(s string, delay time.Duration) {
	for _, c := range s {
		fmt.Print(string(c))
		time.Sleep(delay)
	}
}

func printRoomCode(code string) {
	fmt.Println()

	label := dim.Render("room code") + "  "
	fmt.Print("      " + label)

	typewrite(pink.Render(code), 60*time.Millisecond)
	fmt.Println()
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
