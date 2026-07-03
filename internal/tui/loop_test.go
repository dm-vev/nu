package tui

import (
	"bytes"
	"strings"
	"testing"

	"nu/internal/agent"
)

func TestTUIAppRenderUsesTerminalRepaint(t *testing.T) {
	var out bytes.Buffer
	app := NewApp(AppOptions{
		Stdout:  &out,
		CWD:     "/tmp/nu",
		Model:   "model",
		Repaint: true,
	})

	app.Emit(agent.Event{Type: "turn_start"})
	app.Emit(agent.Event{Type: "message_update", Data: map[string]string{"delta": "hello"}})

	got := out.String()
	if strings.Count(got, syncStart) != 2 || strings.Count(got, repaintPrefix) != 1 {
		t.Fatalf("output = %q, want in-place repaint for each emitted frame", got)
	}
	if strings.Contains(got, "\x1b[2J") {
		t.Fatalf("output = %q, should avoid clear-screen repaint", got)
	}
	if !strings.Contains(got, "hello") {
		t.Fatalf("output = %q, want assistant delta in frame", got)
	}
}
