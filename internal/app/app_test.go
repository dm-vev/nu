package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"nu/internal/cli"
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
	called := false
	code := Run(context.Background(), Options{
		Args:   []string{"--print", "hello"},
		Stdout: &stdout,
		Print: func(rt *Runtime, req cli.Request) error {
			called = true
			if len(req.Prompt) != 1 || req.Prompt[0] != "hello" {
				t.Fatalf("Request prompt = %v, want hello", req.Prompt)
			}
			_, err := rt.Options.Stdout.Write([]byte("ok\n"))
			return err
		},
	})
	if code != exitOK {
		t.Fatalf("Run exit code = %d, want %d", code, exitOK)
	}
	if !called {
		t.Fatal("Print handler was not called")
	}
	if stdout.String() != "ok\n" {
		t.Fatalf("Run stdout = %q, want ok", stdout.String())
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
