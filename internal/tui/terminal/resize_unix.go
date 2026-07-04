//go:build unix

package terminal

import (
	"os"
	"os/signal"
	"syscall"
)

// WatchResize calls render on SIGWINCH.
func WatchResize(render func()) func() {
	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signals, syscall.SIGWINCH)
	go func() {
		for {
			select {
			case <-signals:
				render()
			case <-done:
				signal.Stop(signals)
				close(signals)
				return
			}
		}
	}()
	return func() { close(done) }
}
