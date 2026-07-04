package tui

import "testing"

func TestLimitedCharsetDetectsOptionEnvAndTerm(t *testing.T) {
	if !limitedCharset(AppOptions{ASCII: true}) {
		t.Fatalf("ASCII option should force limited charset")
	}

	t.Setenv("NU_TUI_ASCII", "1")
	if !limitedCharset(AppOptions{}) {
		t.Fatalf("NU_TUI_ASCII=1 should force limited charset")
	}

	t.Setenv("NU_TUI_ASCII", "")
	t.Setenv("TERM", "linux")
	if !limitedCharset(AppOptions{}) {
		t.Fatalf("TERM=linux should force limited charset")
	}

	t.Setenv("TERM", "xterm-256color")
	if limitedCharset(AppOptions{}) {
		t.Fatalf("xterm-256color should keep full charset by default")
	}
}
