package terminal

import (
	"bytes"
	"strings"
	"testing"
)

func TestTerminalTerminalWriteWrapsErrorsByWritingBytes(t *testing.T) {
	var out bytes.Buffer
	term := New(nil, &out, 80, 24)

	if err := term.Write(SyncStart + "x" + SyncEnd); err != nil {
		t.Fatalf("Write error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, "x") {
		t.Fatalf("output = %q", got)
	}
}
