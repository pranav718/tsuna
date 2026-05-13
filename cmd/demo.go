package cmd

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pranav718/tsuna/internal/tui"
	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "preview the tui dashboard with simulated data",
	RunE: func(cmd *cobra.Command, args []string) error {
		ch := make(chan tui.UIEvent, 64)

		go simulateEvents(ch)

		model := tui.NewModel("SAKURA", "local-host-01", "remote-peer-42", true, ch, func() {})
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)
}

func simulateEvents(ch chan<- tui.UIEvent) {
	time.Sleep(500 * time.Millisecond)

	ch <- tui.UIEvent{Type: tui.UIPeerHello, Data: tui.PeerData{
		PeerID: "remote-peer-42", DisplayName: "remote-peer-42",
	}}

	pos := 0.0
	tick := time.NewTicker(500 * time.Millisecond)
	clockTick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	defer clockTick.Stop()

	for {
		select {
		case <-tick.C:
			pos += 0.5
			delta := time.Duration(rand.Intn(30)-15) * time.Millisecond
			ch <- tui.UIEvent{Type: tui.UIStateUpdate, Data: tui.StateData{
				Position:  time.Duration(pos * float64(time.Second)),
				Paused:    false,
				SyncDelta: delta,
			}}

		case <-clockTick.C:
			rtt := time.Duration(8+rand.Intn(10)) * time.Millisecond
			offset := time.Duration(rand.Intn(6)-3) * time.Millisecond
			ch <- tui.UIEvent{Type: tui.UIClockSync, Data: tui.ClockSyncData{
				PeerID: "remote-peer-42",
				RTT:    rtt,
				Offset: offset,
			}}

			if rand.Intn(10) == 0 {
				ch <- tui.UIEvent{Type: tui.UICorrection, Data: tui.CorrectionData{
					CorrType: "micro-seek",
					Delta:    time.Duration(rand.Intn(20)) * time.Millisecond,
					Target:   time.Duration(pos * float64(time.Second)),
				}}
			}
		}
	}
}
