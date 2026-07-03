package tui

import (
	"strings"
	"testing"
)

func TestNUF100RendererDoesNotOverflowWidth(t *testing.T) {
	frame := Render(State{
		Title:  "Nu",
		CWD:    "/very/long/path/that/must/not/overflow",
		Model:  "test-model-with-a-long-name",
		Status: "working",
		Messages: []Message{
			{Role: "user", Text: "this is a very long user message that should be clipped to width"},
			{Role: "assistant", Text: "this is a very long assistant message that should be clipped to width"},
		},
		Editor: EditorSnapshot{Text: "draft text that is also too long"},
	}, 24, 8)

	if len(frame.Lines) > 8 {
		t.Fatalf("frame lines = %d, want <= 8", len(frame.Lines))
	}
	for _, line := range frame.Lines {
		if got := len([]rune(StripANSI(line))); got > 24 {
			t.Fatalf("line visible width = %d, want <= 24: %q", got, line)
		}
	}
}

func TestNUF100ResizeInvalidatesLayout(t *testing.T) {
	state := State{
		Title:    "Nu",
		Messages: []Message{{Role: "assistant", Text: "abcdef"}},
	}

	wide := Render(state, 20, 4)
	narrow := Render(state, 8, 4)

	if len(wide.Lines) != len(narrow.Lines) {
		t.Fatalf("line counts = %d/%d, want same height clamp", len(wide.Lines), len(narrow.Lines))
	}
	if wide.Lines[1] == narrow.Lines[1] {
		t.Fatalf("wide and narrow message lines are equal, want resized layout")
	}
}

func TestNUF100RendererUsesDarkGreenPalette(t *testing.T) {
	frame := Render(State{
		Title:    "Nu",
		Status:   "idle",
		Messages: []Message{{Role: "assistant", Text: "hello"}},
	}, 40, 8)

	got := strings.Join(frame.Lines, "\n")
	if !strings.Contains(got, ansiDarkGreen) {
		t.Fatalf("frame = %q, want dark green accent", got)
	}
	if strings.Contains(got, "\x1b[35m") || strings.Contains(got, "\x1b[36m") {
		t.Fatalf("frame = %q, should avoid purple/cyan palette", got)
	}
}

func TestNUF100RendererTruncatesVisibleTextWithoutBreakingANSI(t *testing.T) {
	frame := Render(State{
		Title:    "Nu",
		Messages: []Message{{Role: "assistant", Text: "abcdef"}},
	}, 6, 4)

	for _, line := range frame.Lines {
		if strings.Contains(line, "\x1b…") {
			t.Fatalf("line = %q, truncated inside ANSI escape", line)
		}
		if got := len([]rune(StripANSI(line))); got > 6 {
			t.Fatalf("visible width = %d, want <= 6: %q", got, line)
		}
	}
}

func TestNUF100RendererWrapsLongMessages(t *testing.T) {
	frame := Render(State{
		Title: "Nu",
		Messages: []Message{{
			Role: "assistant",
			Text: "alpha beta gamma delta epsilon zeta eta theta",
		}},
	}, 20, 20)

	plain := strings.Join(stripLines(frame.Lines), "\n")
	if !strings.Contains(plain, "alpha beta gamma") || !strings.Contains(plain, "delta epsilon zeta") {
		t.Fatalf("plain frame = %q, want wrapped assistant message", plain)
	}
}

func TestNUF100RendererWrapsEditorText(t *testing.T) {
	frame := Render(State{
		Title:  "Nu",
		Editor: EditorSnapshot{Text: "one two three four five six"},
	}, 16, 20)

	plain := strings.Join(stripLines(frame.Lines), "\n")
	if !strings.Contains(plain, "one two three") || !strings.Contains(plain, "four five six") {
		t.Fatalf("plain frame = %q, want wrapped editor text", plain)
	}
}

func stripLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = strings.TrimSpace(StripANSI(line))
	}
	return out
}
