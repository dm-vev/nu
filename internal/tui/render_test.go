package tui

import "testing"

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
