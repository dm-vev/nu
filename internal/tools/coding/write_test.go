package coding

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestNUF071WriteCreatesFile(t *testing.T) {
	dir := t.TempDir()

	result, err := RunWrite(context.Background(), dir, `{"path":"nested/a.txt","content":"ok"}`)
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
			_, err := RunWrite(context.Background(), dir, `{"path":"a.txt","content":"`+content+`"}`)
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
	_, err := RunWrite(context.Background(), t.TempDir(), `{"path":"../x","content":"bad"}`)
	if err == nil || !strings.Contains(err.Error(), "escapes cwd") {
		t.Fatalf("Run error = %v, want cwd escape", err)
	}
}

func TestWriteRejectsSymlinkParentEscape(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
		t.Skipf("Symlink unsupported: %v", err)
	}

	_, err := RunWrite(context.Background(), dir, `{"path":"link/new.txt","content":"bad"}`)
	if err == nil || !strings.Contains(err.Error(), "escapes cwd") {
		t.Fatalf("Run error = %v, want symlink cwd escape", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new.txt")); err == nil {
		t.Fatalf("outside file was created")
	}
}
