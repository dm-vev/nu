package grep

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNUF074GrepLiteralAndRegex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\nhallo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	literal, err := Run(context.Background(), dir, `{"pattern":"hello","literal":true}`, 1000)
	if err != nil {
		t.Fatalf("Run literal error = %v", err)
	}
	regex, err := Run(context.Background(), dir, `{"pattern":"h.llo"}`, 1000)
	if err != nil {
		t.Fatalf("Run regex error = %v", err)
	}
	if !strings.Contains(literal.Content, "a.txt:1:hello") || !strings.Contains(regex.Content, "a.txt:2:hallo") {
		t.Fatalf("grep results literal=%q regex=%q", literal.Content, regex.Content)
	}
}

func TestNUF074GrepRespectsGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("needle\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Run(context.Background(), dir, `{"pattern":"needle","literal":true}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if len(toStrings(t, got["matches"])) != 0 {
		t.Fatalf("matches = %#v, want none", got["matches"])
	}
}

func TestGrepIgnoreCase(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("HELLO\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	result, err := Run(context.Background(), dir, `{"pattern":"hello","literal":true,"ignore_case":true}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(result.Content, "HELLO") {
		t.Fatalf("grep result = %q, want HELLO", result.Content)
	}
}

func TestGrepTruncatesLongMatchingLine(t *testing.T) {
	dir := t.TempDir()
	longLine := strings.Repeat("x", 70*1024) + "needle\n"
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte(longLine), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Run(context.Background(), dir, `{"pattern":"needle","literal":true}`, 10000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(result.Content, "[truncated]") {
		t.Fatalf("grep result = %q, want truncated marker", result.Content)
	}
}

func TestGrepRejectsInvalidRegex(t *testing.T) {
	_, err := Run(context.Background(), t.TempDir(), `{"pattern":"["}`, 1000)
	if err == nil || !strings.Contains(err.Error(), "compile grep pattern") {
		t.Fatalf("Run error = %v, want regex compile error", err)
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
