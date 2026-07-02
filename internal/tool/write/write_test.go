package write

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestNUF071WriteCreatesFile(t *testing.T) {
	dir := t.TempDir()

	result, err := Run(context.Background(), dir, `{"path":"nested/a.txt","content":"ok"}`)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if string(mustRead(t, filepath.Join(dir, "nested/a.txt"))) != "ok" {
		t.Fatalf("file content mismatch")
	}
	got := decodeResult(t, result.Content)
	if got["bytes"] != float64(2) {
		t.Fatalf("write result = %#v, want 2 bytes", got)
	}
}

func TestNUF071ConcurrentWritesSamePathSerialize(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, content := range []string{"one", "two"} {
		wg.Add(1)
		go func(content string) {
			defer wg.Done()
			_, err := Run(context.Background(), dir, `{"path":"a.txt","content":"`+content+`"}`)
			errs <- err
		}(content)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Run error = %v", err)
		}
	}
	got := string(mustRead(t, filepath.Join(dir, "a.txt")))
	if got != "one" && got != "two" {
		t.Fatalf("final content = %q, want complete write", got)
	}
}

func TestWriteRejectsPathEscape(t *testing.T) {
	_, err := Run(context.Background(), t.TempDir(), `{"path":"../x","content":"bad"}`)
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

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	return data
}
