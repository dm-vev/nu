package coding

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNUF075FindGlob(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := RunFind(context.Background(), dir, `{"glob":"*.go"}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if !reflect.DeepEqual(toStrings(t, got["paths"]), []string{"a.go"}) {
		t.Fatalf("paths = %#v, want a.go", got["paths"])
	}
}

func TestNUF075FindRespectsGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := RunFind(context.Background(), dir, `{"glob":"*.txt"}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if len(toStrings(t, got["paths"])) != 0 {
		t.Fatalf("paths = %#v, want none", got["paths"])
	}
}

func TestFindEnforcesLimit(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0o644); err != nil {
			t.Fatalf("WriteFile error = %v", err)
		}
	}
	result, err := RunFind(context.Background(), dir, `{"glob":"*.txt","limit":1}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if len(toStrings(t, got["paths"])) != 1 {
		t.Fatalf("paths = %#v, want one result", got["paths"])
	}
}

func TestFindRejectsPathEscape(t *testing.T) {
	_, err := RunFind(context.Background(), t.TempDir(), `{"root":".."}`, 1000)
	if err == nil || !strings.Contains(err.Error(), "escapes cwd") {
		t.Fatalf("Run error = %v, want cwd escape", err)
	}
}
