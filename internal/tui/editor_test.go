package tui

import "testing"

func TestNUF101EditorInsertDeleteUndo(t *testing.T) {
	editor := NewEditor()

	editor.Insert("привет")
	editor.Move(-2)
	editor.Backspace()
	if got := editor.Snapshot().Text; got != "приет" {
		t.Fatalf("text after backspace = %q, want приет", got)
	}
	if !editor.Undo() {
		t.Fatalf("Undo returned false, want true")
	}
	snapshot := editor.Snapshot()
	if snapshot.Text != "привет" || snapshot.Cursor != 4 {
		t.Fatalf("snapshot after undo = %#v, want text restored with cursor", snapshot)
	}
	if submitted := editor.Submit(); submitted != "привет" {
		t.Fatalf("Submit = %q, want привет", submitted)
	}
	if got := editor.Snapshot().Text; got != "" {
		t.Fatalf("text after submit = %q, want empty", got)
	}
}
