package tui

import (
	"fmt"
	"io"
	"strings"
)

const (
	repaintPrefix = "\x1b[H"
	syncStart     = "\x1b[?2026h"
	syncEnd       = "\x1b[?2026l"
	hideCursor    = "\x1b[?25l"
	showCursor    = "\x1b[?25h"
	bracketedOn   = "\x1b[?2004h"
	bracketedOff  = "\x1b[?2004l"
)

// Terminal writes rendered frames to an injected output.
type Terminal struct {
	w             io.Writer
	repaint       bool
	started       bool
	lastLines     int
	lastCursorRow int
}

// NewTerminal creates a frame writer.
func NewTerminal(w io.Writer, repaint bool) *Terminal {
	if w == nil {
		w = io.Discard
	}
	return &Terminal{w: w, repaint: repaint}
}

// Draw writes one frame.
func (t *Terminal) Draw(frame Frame) error {
	if !t.repaint {
		for _, line := range frame.Lines {
			if _, err := fmt.Fprintln(t.w, line); err != nil {
				return fmt.Errorf("write terminal frame line: %w", err)
			}
		}
		return nil
	}

	if err := t.start(frame); err != nil {
		return err
	}

	lines := frame.Lines
	if t.lastLines > len(lines) {
		blank := withReset(strings.Repeat(" ", frame.Width))
		for len(lines) < t.lastLines {
			lines = append(lines, blank)
		}
	}
	for i, line := range lines {
		if _, err := io.WriteString(t.w, line); err != nil {
			return fmt.Errorf("write terminal frame line: %w", err)
		}
		if i < len(lines)-1 {
			if _, err := io.WriteString(t.w, "\r\n"); err != nil {
				return fmt.Errorf("write terminal newline: %w", err)
			}
		}
	}
	if _, err := io.WriteString(t.w, syncEnd); err != nil {
		return fmt.Errorf("write terminal sync end: %w", err)
	}
	if up := len(lines) - 1 - frame.CursorRow; up > 0 {
		if _, err := fmt.Fprintf(t.w, "\x1b[%dA", up); err != nil {
			return fmt.Errorf("write terminal cursor row: %w", err)
		}
	}
	cursorCol := frame.CursorCol
	if cursorCol < 1 {
		cursorCol = 1
	}
	if _, err := fmt.Fprintf(t.w, "\x1b[%dG%s", cursorCol, hideCursor); err != nil {
		return fmt.Errorf("write terminal cursor col: %w", err)
	}
	t.lastLines = len(frame.Lines)
	t.lastCursorRow = frame.CursorRow
	return nil
}

// Close restores the terminal bits that Draw enables for repaint mode.
func (t *Terminal) Close() error {
	if !t.repaint || !t.started {
		return nil
	}
	if t.lastLines > 0 {
		if down := t.lastLines - 1 - t.lastCursorRow; down > 0 {
			if _, err := fmt.Fprintf(t.w, "\x1b[%dB", down); err != nil {
				return fmt.Errorf("move terminal below frame: %w", err)
			}
		}
		if _, err := io.WriteString(t.w, "\x1b[1G\r\n"); err != nil {
			return fmt.Errorf("write terminal close newline: %w", err)
		}
	}
	if _, err := io.WriteString(t.w, syncEnd+showCursor+bracketedOff); err != nil {
		return fmt.Errorf("restore terminal: %w", err)
	}
	return nil
}

// Size returns the current terminal size when the writer is a TTY.
func (t *Terminal) Size(fallbackWidth int, fallbackHeight int) (int, int) {
	return terminalSize(t.w, fallbackWidth, fallbackHeight)
}

func (t *Terminal) start(frame Frame) error {
	if !t.started {
		t.started = true
		title := firstNonEmpty(frame.Title, "Nu")
		// The terminal title is a side effect, so keep it beside the other TTY setup bytes.
		if _, err := io.WriteString(t.w, bracketedOn+hideCursor+"\x1b]0;"+title+"\x07"+syncStart); err != nil {
			return fmt.Errorf("start terminal repaint: %w", err)
		}
		return nil
	}
	if _, err := io.WriteString(t.w, syncStart+repaintPrefix); err != nil {
		return fmt.Errorf("write terminal repaint prefix: %w", err)
	}
	return nil
}
