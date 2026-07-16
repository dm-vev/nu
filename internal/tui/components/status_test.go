package components

import (
	"strings"
	"testing"
)

func TestStatusStatusAlwaysReservesOneLine(t *testing.T) {
	s := NewStatus(nil)
	if lines := s.Render(20); len(lines) != 1 || lines[0] != "                    " {
		t.Fatalf("idle lines = %#v, want one blank line", lines)
	}
	s.SetText("running")
	if lines := s.Render(20); len(lines) != 1 {
		t.Fatalf("busy lines = %#v, want one", lines)
	}
}

func TestStatusStatusStepAnimatesLabel(t *testing.T) {
	s := NewStatus(nil)
	s.SetText("bubbling")

	first := s.Render(20)[0]
	s.Step()
	second := s.Render(20)[0]

	if first == second {
		t.Fatalf("animated status did not change: %q", first)
	}
}

func TestStatusStatusUsesClaudeLikeFrames(t *testing.T) {
	s := NewStatus(nil)
	s.SetText("running")

	first := s.Render(20)[0]
	s.Step()
	second := s.Render(20)[0]

	if !strings.HasPrefix(first, " ·") || !strings.HasPrefix(second, " ✢") {
		t.Fatalf("frames = %q, %q; want Claude-like first frames", first, second)
	}
}

func TestStatusStatusCanUseASCIIFrames(t *testing.T) {
	s := NewStatus(nil, "-", "\\", "|", "/")
	s.SetText("running")

	first := s.Render(20)[0]
	s.Step()
	second := s.Render(20)[0]

	if !strings.HasPrefix(first, " -") || !strings.HasPrefix(second, " \\") {
		t.Fatalf("frames = %q, %q; want ASCII spinner frames", first, second)
	}
}
