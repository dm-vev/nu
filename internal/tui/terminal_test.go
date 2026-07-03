package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestTerminalDrawRepaintsWithANSI(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminal(&out, true)

	if err := term.Draw(Frame{Lines: []string{"one"}, Width: 3, CursorCol: 1}); err != nil {
		t.Fatalf("Draw first error = %v", err)
	}
	if err := term.Draw(Frame{Lines: []string{"two"}, Width: 3, CursorCol: 1}); err != nil {
		t.Fatalf("Draw second error = %v", err)
	}

	got := out.String()
	if !strings.HasPrefix(got, bracketedOn+hideCursor) {
		t.Fatalf("output = %q, want terminal setup prefix", got)
	}
	if strings.Count(got, syncStart) != 2 || strings.Count(got, repaintPrefix) != 1 {
		t.Fatalf("output = %q, want sync draw and one home repaint", got)
	}
	if strings.Contains(got, "\x1b[2J") {
		t.Fatalf("output = %q, should not clear the screen", got)
	}
	if !strings.Contains(got, "one") || !strings.Contains(got, "two") {
		t.Fatalf("output = %q, want frame lines", got)
	}
}

func TestTerminalDrawAppendModePlain(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminal(&out, false)

	if err := term.Draw(Frame{Lines: []string{"one"}}); err != nil {
		t.Fatalf("Draw error = %v", err)
	}

	if got := out.String(); got != "one\n" {
		t.Fatalf("output = %q, want plain frame", got)
	}
}

func TestTerminalCloseMovesBelowFrame(t *testing.T) {
	var out bytes.Buffer
	term := NewTerminal(&out, true)

	frame := Frame{
		Lines:     []string{"top", "editor", "footer"},
		Width:     6,
		CursorRow: 1,
		CursorCol: 1,
	}
	if err := term.Draw(frame); err != nil {
		t.Fatalf("Draw error = %v", err)
	}
	if err := term.Close(); err != nil {
		t.Fatalf("Close error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "\x1b[1B\x1b[1G\r\n") {
		t.Fatalf("output = %q, want cursor moved below rendered frame", got)
	}
	if !strings.Contains(got, showCursor+bracketedOff) {
		t.Fatalf("output = %q, want terminal restore bytes", got)
	}
}
