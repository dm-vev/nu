package engine

import (
	"nu/internal/tui/core"
	"nu/internal/tui/terminal"
)

// TUI manages a component tree and terminal diff rendering.
type TUI struct {
	core.Container
	terminal *terminal.Terminal
	opts     Options

	previousLines []string
	previousWidth int
	previousRows  int
	cursorRow     int
	cursorCol     int
	scrollOffset  int
	started       bool
}

// New creates a TUI engine.
func New(term *terminal.Terminal, opts Options) *TUI {
	if opts.Title == "" {
		opts.Title = "Nu"
	}
	if opts.MinRenderRows <= 0 {
		opts.MinRenderRows = 1
	}
	if !opts.SynchronizedDraw {
		opts.SynchronizedDraw = true
	}
	return &TUI{terminal: term, opts: opts}
}

// Terminal returns the underlying terminal.
func (t *TUI) Terminal() *terminal.Terminal {
	return t.terminal
}

// ScrollBy moves the viewport away from or toward the bottom.
func (t *TUI) ScrollBy(delta int) bool {
	old := t.scrollOffset
	t.scrollOffset += delta
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
	return t.scrollOffset != old
}

// ScrollToBottom restores automatic bottom following.
func (t *TUI) ScrollToBottom() bool {
	if t.scrollOffset == 0 {
		return false
	}
	t.scrollOffset = 0
	return true
}
