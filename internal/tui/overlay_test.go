package tui

import "testing"

func TestNUF100OverlayFocusRestoresPrevious(t *testing.T) {
	stack := NewOverlayStack()
	first := stack.Push("first")
	second := stack.Push("second")

	focused, ok := stack.Focused()
	if !ok || focused.ID != second.ID {
		t.Fatalf("focused = %#v/%v, want second", focused, ok)
	}
	if !stack.Close(second) {
		t.Fatalf("Close(second) = false, want true")
	}
	focused, ok = stack.Focused()
	if !ok || focused.ID != first.ID {
		t.Fatalf("focused after close = %#v/%v, want first", focused, ok)
	}
	if stack.Close(second) {
		t.Fatalf("Close(disposed second) = true, want false")
	}
}
