package tool

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNUF070ReadTextWithOffsetLimit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("0123456789"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Read(context.Background(), dir, `{"path":"a.txt","offset":2,"limit":4}`, 100)
	if err != nil {
		t.Fatalf("Read error = %v", err)
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

	result, err := Read(context.Background(), dir, `{"path":"a.txt"}`, 3)
	if err != nil {
		t.Fatalf("Read error = %v", err)
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

	result, err := Read(context.Background(), dir, `{"path":"a.png"}`, 100)
	if err != nil {
		t.Fatalf("Read error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["mime_type"] != "image/png" || got["data"] != base64.StdEncoding.EncodeToString(data) {
		t.Fatalf("image result = %#v, want png attachment", got)
	}
}

func TestNUF071WriteCreatesFile(t *testing.T) {
	dir := t.TempDir()

	result, err := Write(context.Background(), dir, `{"path":"nested/a.txt","content":"ok"}`)
	if err != nil {
		t.Fatalf("Write error = %v", err)
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
			_, err := Write(context.Background(), dir, `{"path":"a.txt","content":"`+content+`"}`)
			errs <- err
		}(content)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Write error = %v", err)
		}
	}
	got := string(mustRead(t, filepath.Join(dir, "a.txt")))
	if got != "one" && got != "two" {
		t.Fatalf("final content = %q, want complete write", got)
	}
}

func TestNUF072EditSingleReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello old\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Edit(context.Background(), dir, `{"path":"a.txt","replacements":[{"old":"old","new":"new"}]}`)
	if err != nil {
		t.Fatalf("Edit error = %v", err)
	}
	if string(mustRead(t, path)) != "hello new\n" {
		t.Fatalf("edited content mismatch")
	}
	if !strings.Contains(result.Content, "-old") || !strings.Contains(result.Content, "+new") {
		t.Fatalf("patch result = %q, want old/new patch", result.Content)
	}
}

func TestNUF072EditRejectsAmbiguousOldText(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x x"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	_, err := Edit(context.Background(), dir, `{"path":"a.txt","replacements":[{"old":"x","new":"y"}]}`)
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("Edit error = %v, want ambiguous", err)
	}
}

func TestNUF072EditPreservesCRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("a\r\nold\r\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	if _, err := Edit(context.Background(), dir, `{"path":"a.txt","replacements":[{"old":"old","new":"new"}]}`); err != nil {
		t.Fatalf("Edit error = %v", err)
	}
	if string(mustRead(t, path)) != "a\r\nnew\r\n" {
		t.Fatalf("CRLF was not preserved")
	}
}

func TestNUF073BashCapturesStdoutAndStderr(t *testing.T) {
	result, err := Bash(context.Background(), t.TempDir(), `{"command":"printf out; printf err >&2; exit 7"}`, 100)
	if err != nil {
		t.Fatalf("Bash error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if got["stdout"] != "out" || got["stderr"] != "err" || got["exit_code"] != float64(7) {
		t.Fatalf("bash result = %#v, want stdout/stderr/exit", got)
	}
}

func TestNUF073BashTimeoutKillsProcess(t *testing.T) {
	start := time.Now()
	result, err := Bash(context.Background(), t.TempDir(), `{"command":"sleep 2","timeout_ms":50}`, 100)
	if err != nil {
		t.Fatalf("Bash error = %v", err)
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
	result, err := Bash(context.Background(), t.TempDir(), `{"command":"printf abcdef"}`, 3)
	if err != nil {
		t.Fatalf("Bash error = %v", err)
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

func TestNUF074GrepLiteralAndRegex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello\nhallo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	literal, err := Grep(context.Background(), dir, `{"pattern":"hello","literal":true}`, 1000)
	if err != nil {
		t.Fatalf("Grep literal error = %v", err)
	}
	regex, err := Grep(context.Background(), dir, `{"pattern":"h.llo"}`, 1000)
	if err != nil {
		t.Fatalf("Grep regex error = %v", err)
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

	result, err := Grep(context.Background(), dir, `{"pattern":"needle","literal":true}`, 1000)
	if err != nil {
		t.Fatalf("Grep error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if len(toStrings(t, got["matches"])) != 0 {
		t.Fatalf("matches = %#v, want none", got["matches"])
	}
}

func TestNUF075FindGlob(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	result, err := Find(context.Background(), dir, `{"glob":"*.go"}`, 1000)
	if err != nil {
		t.Fatalf("Find error = %v", err)
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

	result, err := Find(context.Background(), dir, `{"glob":"*.txt"}`, 1000)
	if err != nil {
		t.Fatalf("Find error = %v", err)
	}
	got := decodeResult(t, result.Content)
	if len(toStrings(t, got["paths"])) != 0 {
		t.Fatalf("paths = %#v, want none", got["paths"])
	}
}

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

	result, err := Ls(context.Background(), dir, `{}`, 1000)
	if err != nil {
		t.Fatalf("Ls error = %v", err)
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

	_, err := Ls(context.Background(), dir, `{"path":"file"}`, 1000)
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("Ls error = %v, want not a directory", err)
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

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("file %s does not exist", path)
		}
		t.Fatalf("ReadFile error = %v", err)
	}
	return data
}
