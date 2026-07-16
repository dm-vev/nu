package coding

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNUF076LsSortedWithDirs(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "dir"), 0o755); err != nil {
		t.Fatalf("Mkdir error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := RunLS(context.Background(), dir, `{}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	want := []string{".env", "dir/", "file"}
	if !reflect.DeepEqual(toStrings(t, got["entries"]), want) {
		t.Fatalf("entries = %#v, want %#v", got["entries"], want)
	}
}

func TestNUF076LsRejectsNonDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	_, err := RunLS(context.Background(), dir, `{"path":"file"}`, 1000)
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("Run error = %v, want not a directory", err)
	}
}

func TestLsRejectsPathEscape(t *testing.T) {
	_, err := RunLS(context.Background(), t.TempDir(), `{"path":".."}`, 1000)
	if err == nil || !strings.Contains(err.Error(), "escapes cwd") {
		t.Fatalf("Run error = %v, want cwd escape", err)
	}
}
