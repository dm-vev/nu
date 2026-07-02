package ls

import (
	"context"
	"encoding/json"
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

	result, err := Run(context.Background(), dir, `{}`, 1000)
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

	_, err := Run(context.Background(), dir, `{"path":"file"}`, 1000)
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("Run error = %v, want not a directory", err)
	}
}

func TestLsRejectsPathEscape(t *testing.T) {
	_, err := Run(context.Background(), t.TempDir(), `{"path":".."}`, 1000)
	if err == nil || !strings.Contains(err.Error(), "escapes cwd") {
		t.Fatalf("Run error = %v, want cwd escape", err)
	}
}

func decodeResult(t *testing.T, raw string) map[string]any {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("result JSON error = %v; raw=%q", err, raw)
	}
	return got
}

func toStrings(t *testing.T, raw any) []string {
	t.Helper()
	values, ok := raw.([]any)
	if !ok {
		t.Fatalf("value = %#v, want []any", raw)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("item = %#v, want string", value)
		}
		out = append(out, text)
	}
	return out
}
