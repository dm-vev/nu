package app

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"nu/internal/provider"
	"nu/internal/testkit"
)

func TestAppRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--help"},
		Stdout: &stdout,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "Usage: nu") {
		t.Fatalf("Run help stdout = %q, want usage", stdout.String())
	}
}

func TestAppRunPrintModeUsesInjectedRuntime(t *testing.T) {
	var stdout bytes.Buffer
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "ok"},
		provider.Event{Type: provider.EventDone},
	)
	code := Run(context.Background(), Options{
		Args:     []string{"--print", "hello"},
		Stdout:   &stdout,
		Provider: fake,
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	requests := fake.Requests()
	if len(requests) != 1 {
		t.Fatalf("Provider requests = %d, want 1", len(requests))
	}
	if requests[0].Messages[0].Content != "hello" {
		t.Fatalf("Provider prompt = %q, want hello", requests[0].Messages[0].Content)
	}
	if stdout.String() != "ok\n" {
		t.Fatalf("Run stdout = %q, want ok", stdout.String())
	}
}

func TestNUF170JSONModeStdoutIsOnlyJSONL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fake := testkit.NewScriptedProvider(
		provider.Event{Type: provider.EventStart},
		provider.Event{Type: provider.EventText, Delta: "ok"},
		provider.Event{Type: provider.EventDone},
	)
	code := Run(context.Background(), Options{
		Args:      []string{"--mode", "json", "hello"},
		CWD:       "/tmp/nu-test",
		Stdout:    &stdout,
		Stderr:    &stderr,
		Provider:  fake,
		SessionID: "s1",
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("Run stderr = %q, want empty", stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("JSON mode lines = %d, want header and events; stdout=%q", len(lines), stdout.String())
	}
	var header map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatalf("Header JSON error = %v", err)
	}
	if header["type"] != "session" || header["id"] != "s1" {
		t.Fatalf("Header = %#v, want session s1", header)
	}
	for _, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("JSONL line %q error = %v", line, err)
		}
	}
	var last map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("Last JSON error = %v", err)
	}
	if last["type"] != "turn_end" {
		t.Fatalf("Last event = %#v, want turn_end", last)
	}
}

func TestAppRunPrintModeWithoutHandlerFails(t *testing.T) {
	var stderr bytes.Buffer
	code := Run(context.Background(), Options{
		Args:   []string{"--print", "hello"},
		Stderr: &stderr,
	})
	if code != exitError {
		t.Fatalf("Run exit code = %d, want %d", code, exitError)
	}
	if !strings.Contains(stderr.String(), "print mode requires agent handler") {
		t.Fatalf("Run stderr = %q, want missing handler error", stderr.String())
	}
}
