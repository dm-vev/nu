package engine

import (
	"bytes"
	"strings"
	"testing"

	"nu/internal/tui/components/text"
	"nu/internal/tui/core"
	"nu/internal/tui/terminal"
)

func TestEngineRendersComponentTreeAndDiffs(t *testing.T) {
	var out bytes.Buffer
	term := terminal.New(nil, &out, 20, 10)
	ui := New(term, Options{Title: "Nu"})
	label := text.New("hello", text.Options{})
	ui.AddChild(label)

	if err := ui.Start(); err != nil {
		t.Fatalf("Start error = %v", err)
	}
	if err := ui.RenderNow(); err != nil {
		t.Fatalf("RenderNow first error = %v", err)
	}
	label.SetText("hello world")
	if err := ui.RenderNow(); err != nil {
		t.Fatalf("RenderNow second error = %v", err)
	}

	got := out.String()
	if strings.Contains(got, "\x1b[?1000h") || strings.Contains(got, "\x1b[?1006h") {
		t.Fatalf("output = %q, should not enable mouse reporting because it breaks terminal text selection", got)
	}
	if strings.Count(got, terminal.SyncStart) != 2 {
		t.Fatalf("output = %q, want two synchronized renders", got)
	}
	if strings.Count(got, "\x1b[2J") != 1 {
		t.Fatalf("output = %q, want one initial clear", got)
	}
}

func TestEngineRendersBottomViewportWhenContentOverflows(t *testing.T) {
	var out bytes.Buffer
	term := terminal.New(nil, &out, 20, 3)
	ui := New(term, Options{Title: "Nu"})
	ui.AddChild(text.New("top", text.Options{}))
	ui.AddChild(text.New("middle", text.Options{}))
	ui.AddChild(text.New("bottom", text.Options{}))
	ui.AddChild(text.New("footer", text.Options{}))

	if err := ui.Start(); err != nil {
		t.Fatalf("Start error = %v", err)
	}
	if err := ui.RenderNow(); err != nil {
		t.Fatalf("RenderNow error = %v", err)
	}

	got := out.String()
	if strings.Contains(got, "top") || !strings.Contains(got, "footer") {
		t.Fatalf("output = %q, want bottom viewport only", got)
	}
}

func TestEngineScrollsOverflowingViewport(t *testing.T) {
	var out bytes.Buffer
	term := terminal.New(nil, &out, 20, 3)
	ui := New(term, Options{Title: "Nu"})
	ui.AddChild(text.New("top", text.Options{}))
	ui.AddChild(text.New("middle", text.Options{}))
	ui.AddChild(text.New("bottom", text.Options{}))
	ui.AddChild(text.New("footer", text.Options{}))

	if err := ui.Start(); err != nil {
		t.Fatalf("Start error = %v", err)
	}
	if changed := ui.ScrollBy(1); !changed {
		t.Fatalf("ScrollBy did not report viewport change")
	}
	if err := ui.RenderNow(); err != nil {
		t.Fatalf("RenderNow error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "top") || !strings.Contains(got, "middle") || strings.Contains(got, "footer") {
		t.Fatalf("output = %q, want viewport shifted up one row", got)
	}
}

func TestEngineDiffUsesAbsoluteRowsNearBottom(t *testing.T) {
	var out bytes.Buffer
	term := terminal.New(nil, &out, 20, 3)
	ui := New(term, Options{Title: "Nu"})
	label := text.New("a", text.Options{})
	ui.AddChild(text.New("top", text.Options{}))
	ui.AddChild(text.New("middle", text.Options{}))
	ui.AddChild(label)

	if err := ui.Start(); err != nil {
		t.Fatalf("Start error = %v", err)
	}
	if err := ui.RenderNow(); err != nil {
		t.Fatalf("RenderNow first error = %v", err)
	}
	out.Reset()
	label.SetText("bottom" + core.CursorMarker)
	if err := ui.RenderNow(); err != nil {
		t.Fatalf("RenderNow second error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "\x1b[3;1H\x1b[2Kbottom") {
		t.Fatalf("output = %q, want absolute bottom-row redraw", got)
	}
	if strings.Contains(got, "\r\n\x1b[2K") {
		t.Fatalf("output = %q, want no newline-based bottom redraw", got)
	}
}
