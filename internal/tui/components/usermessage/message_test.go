package usermessage

import (
	"strings"
	"testing"
)

func TestUserMessageWrapsAndAddsPromptZone(t *testing.T) {
	msg := New("hello from user", Options{PaddingX: 1, PaddingY: 1})
	lines := msg.Render(20)
	if len(lines) < 3 {
		t.Fatalf("lines = %#v, want padded block", lines)
	}
	if !strings.HasPrefix(lines[0], osc133ZoneStart) {
		t.Fatalf("first line = %q, want OSC zone start", lines[0])
	}
	if !strings.Contains(strings.Join(lines, "\n"), "hello from user") {
		t.Fatalf("lines = %#v, want text", lines)
	}
}
