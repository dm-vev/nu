package coding

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNUF072EditSingleReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello old\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := RunEdit(context.Background(), dir, `{"path":"a.txt","replacements":[{"old":"old","new":"new"}]}`)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if string(mustRead(t, path)) != "hello new\n" {
		t.Fatalf("edited content mismatch")
	}
	if !strings.Contains(result.Content, "-old") || !strings.Contains(result.Content, "+new") {
		t.Fatalf("patch result = %q, want old/new patch", result.Content)
	}
}

func TestNUF072EditRejectsAmbiguousOldText(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x x"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	_, err := RunEdit(context.Background(), dir, `{"path":"a.txt","replacements":[{"old":"x","new":"y"}]}`)
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("Run error = %v, want ambiguous", err)
	}
}

func TestNUF072EditPreservesCRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a\r\nold\r\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	if _, err := RunEdit(context.Background(), dir, `{"path":"a.txt","replacements":[{"old":"old","new":"new"}]}`); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if string(mustRead(t, path)) != "a\r\nnew\r\n" {
		t.Fatalf("CRLF was not preserved")
	}
}

func TestEditAppliesMultipleReplacementsAgainstOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a b"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	_, err := RunEdit(
		context.Background(),
		dir,
		`{"path":"a.txt","replacements":[{"old":"a","new":"b"},{"old":"b","new":"c"}]}`,
	)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if string(mustRead(t, path)) != "b c" {
		t.Fatalf("edited content = %q, want b c", string(mustRead(t, path)))
	}
}

func TestEditRejectsMissingOldText(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("abc"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	_, err := RunEdit(context.Background(), dir, `{"path":"a.txt","replacements":[{"old":"missing","new":"x"}]}`)
	if err == nil || !strings.Contains(err.Error(), "missing old text") {
		t.Fatalf("Run error = %v, want missing old text", err)
	}
}
