package web

import "github.com/pranav718/tsuna/internal/tui"

func FanOut(src <-chan tui.UIEvent, dsts ...chan<- tui.UIEvent) {
	go func() {
		for ev := range src {
			for _, dst := range dsts {
				select {
				case dst <- ev:
				default:
				}
			}
		}
		for _, dst := range dsts {
			close(dst)
		}
	}()
}
