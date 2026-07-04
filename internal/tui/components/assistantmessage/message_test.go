package assistantmessage

import (
	"strings"
	"testing"

	"nu/internal/tui/ansi"
	tuimessage "nu/internal/tui/message"
)

func TestAssistantMessageRendersTextAndZone(t *testing.T) {
	msg := New("assistant text", Options{PaddingX: 1})
	lines := msg.Render(40)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "assistant text") {
		t.Fatalf("lines = %#v, want text", lines)
	}
	if !strings.HasPrefix(lines[0], osc133ZoneStart) {
		t.Fatalf("first line = %q, want OSC zone start", lines[0])
	}
}

func TestAssistantMessageRendersPartsWithoutZoneWhenToolExists(t *testing.T) {
	value := tuimessage.NewAssistant()
	value.AppendThinking("checking")
	value.AddTool("call-1", "bash", `{"command":"pwd"}`)
	value.FinishTool("call-1", `{"output":"/tmp\n","exit_code":0}`, false)
	value.AppendText("done")

	msg := NewMessage(value, Options{
		PaddingX:      1,
		TextStyle:     func(value string) string { return value },
		ThinkingStyle: func(value string) string { return ansi.Italic + value + ansi.ItalicOff },
		ToolSuccessBg: func(value string) string { return ansi.ToolSuccessBG + value + ansi.DefaultBG },
	})
	lines := msg.Render(60)
	joined := strings.Join(lines, "\n")
	plain := ansi.Strip(joined)
	if strings.HasPrefix(lines[0], osc133ZoneStart) {
		t.Fatalf("first line = %q, want no OSC zone when tool part exists", lines[0])
	}
	for _, want := range []string{"checking", "$ pwd", "/tmp", "done"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("lines = %#v, want %q", lines, want)
		}
	}
}
