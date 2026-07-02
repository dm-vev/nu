package find

import (
	"context"
	"encoding/json"
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

	result, err := Run(context.Background(), dir, `{"glob":"*.go"}`, 1000)
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

	result, err := Run(context.Background(), dir, `{"glob":"*.txt"}`, 1000)
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
	result, err := Run(context.Background(), dir, `{"glob":"*.txt","limit":1}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if len(toStrings(t, got["paths"])) != 1 {
		t.Fatalf("paths = %#v, want one result", got["paths"])
	}
}

func TestFindRejectsPathEscape(t *testing.T) {
	_, err := Run(context.Background(), t.TempDir(), `{"root":".."}`, 1000)
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
