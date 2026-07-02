package rpc

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRPCJSONLReadLinesStrictLF(t *testing.T) {
	var lines []string
	input := "{\"text\":\"a\u2028b\"}\r\n{\"ok\":true}"

	if err := ReadLines(strings.NewReader(input), func(line string) error {
		lines = append(lines, line)
		return nil
	}); err != nil {
		t.Fatalf("ReadLines error = %v", err)
	}

	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2: %#v", len(lines), lines)
	}
	if lines[0] != "{\"text\":\"a\u2028b\"}" {
		t.Fatalf("first line = %q, want Unicode separator preserved", lines[0])
	}
	if lines[1] != "{\"ok\":true}" {
		t.Fatalf("second line = %q, want final unterminated line", lines[1])
	}
}

func TestRPCJSONLReadLinesReturnsCallbackError(t *testing.T) {
	want := errors.New("stop")

	err := ReadLines(strings.NewReader("one\ntwo\n"), func(string) error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("ReadLines error = %v, want callback error", err)
	}
}

func TestRPCJSONLWriteLine(t *testing.T) {
	var out bytes.Buffer

	if err := WriteLine(&out, map[string]any{"type": "response", "ok": true}); err != nil {
		t.Fatalf("WriteLine error = %v", err)
	}

	if got := out.String(); got != "{\"ok\":true,\"type\":\"response\"}\n" {
		t.Fatalf("WriteLine output = %q", got)
	}
}
