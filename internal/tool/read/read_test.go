package read

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNUF070ReadTextWithOffsetLimit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("0123456789"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Run(context.Background(), dir, `{"path":"a.txt","offset":2,"limit":4}`, 100)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["content"] != "2345" {
		t.Fatalf("content = %#v, want 2345", got["content"])
	}
}

func TestNUF070ReadTruncatesLargeFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("abcdef"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Run(context.Background(), dir, `{"path":"a.txt"}`, 3)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["content"] != "abc" || got["truncated"] != true {
		t.Fatalf("read result = %#v, want truncated abc", got)
	}
}

func TestNUF070ReadImageAttachment(t *testing.T) {
	dir := t.TempDir()
	data := []byte{0x89, 'P', 'N', 'G'}
	if err := os.WriteFile(filepath.Join(dir, "a.png"), data, 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Run(context.Background(), dir, `{"path":"a.png"}`, 100)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["mime_type"] != "image/png" || got["data"] != base64.StdEncoding.EncodeToString(data) {
		t.Fatalf("image result = %#v, want png attachment", got)
	}
}

func TestReadRejectsPathEscape(t *testing.T) {
	_, err := Run(context.Background(), t.TempDir(), `{"path":"../x"}`, 100)
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
