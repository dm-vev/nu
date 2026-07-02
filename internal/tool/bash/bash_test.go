package bash

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNUF073BashCapturesStdoutAndStderr(t *testing.T) {
	result, err := Run(context.Background(), t.TempDir(), `{"command":"printf out; printf err >&2; exit 7"}`, 100)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["stdout"] != "out" || got["stderr"] != "err" || got["exit_code"] != float64(7) {
		t.Fatalf("bash result = %#v, want stdout/stderr/exit", got)
	}
}

func TestNUF073BashTimeoutKillsProcess(t *testing.T) {
	start := time.Now()
	result, err := Run(context.Background(), t.TempDir(), `{"command":"sleep 2","timeout_ms":50}`, 100)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("Bash timeout took too long")
	}
	got := decodeResult(t, result.Content)
	if got["timed_out"] != true {
		t.Fatalf("bash result = %#v, want timed_out", got)
	}
}

func TestNUF073BashTruncatesAndPersistsFullOutput(t *testing.T) {
	result, err := Run(context.Background(), t.TempDir(), `{"command":"printf abcdef"}`, 3)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["output"] != "abc" || got["truncated"] != true {
		t.Fatalf("bash result = %#v, want truncated abc", got)
	}
	fullPath, ok := got["full_output_path"].(string)
	if !ok || string(mustRead(t, fullPath)) != "abcdef" {
		t.Fatalf("full output path = %#v, want persisted abcdef", got["full_output_path"])
	}
}

func TestBashRejectsEmptyCommand(t *testing.T) {
	_, err := Run(context.Background(), t.TempDir(), `{"command":"   "}`, 100)
	if err == nil || !strings.Contains(err.Error(), "missing command") {
		t.Fatalf("Run error = %v, want missing command", err)
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
