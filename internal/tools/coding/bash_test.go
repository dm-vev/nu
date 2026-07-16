package coding

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNUF073BashCapturesStdoutAndStderr(t *testing.T) {
	result, err := RunBash(context.Background(), t.TempDir(), `{"command":"printf out; printf err >&2; exit 7"}`, 100)
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
	result, err := RunBash(context.Background(), t.TempDir(), `{"command":"sleep 2","timeout_ms":50}`, 100)
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
	result, err := RunBash(context.Background(), t.TempDir(), `{"command":"printf abcdef"}`, 3)
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
	_, err := RunBash(context.Background(), t.TempDir(), `{"command":"   "}`, 100)
	if err == nil || !strings.Contains(err.Error(), "missing command") {
		t.Fatalf("Run error = %v, want missing command", err)
	}
}

func TestBashRejectsInteractiveSudo(t *testing.T) {
	result, err := RunBash(context.Background(), t.TempDir(), `{"command":"sudo true"}`, 1000)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["exit_code"] != float64(1) || !strings.Contains(got["stderr"].(string), "sudo is disabled") {
		t.Fatalf("bash result = %#v, want non-interactive sudo failure", got)
	}
}

func TestBashAllowsNonInteractiveSudoForms(t *testing.T) {
	for _, command := range []string{"sudo -n true", "sudo -S true", "sudo --non-interactive true"} {
		if usesInteractiveSudo(command) {
			t.Fatalf("usesInteractiveSudo(%q) = true, want false", command)
		}
	}
}

func TestBashKeepsUnrelatedSudoFlagsInteractive(t *testing.T) {
	if !usesInteractiveSudo("sudo --preserve-env true") {
		t.Fatalf("usesInteractiveSudo(--preserve-env) = false, want true")
	}
}
