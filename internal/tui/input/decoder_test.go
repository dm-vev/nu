package input

import (
	"io"
	"strings"
	"testing"
)

func TestInputDecoderReadsPrintableUTF8AndEscapeSequence(t *testing.T) {
	decoder := New(strings.NewReader("я\x1b[D"))

	first, err := decoder.Read()
	if err != nil {
		t.Fatalf("read utf8: %v", err)
	}
	if first.Data != "я" {
		t.Fatalf("first = %q, want utf8 rune", first.Data)
	}

	second, err := decoder.Read()
	if err != nil {
		t.Fatalf("read escape: %v", err)
	}
	if second.Data != "\x1b[D" {
		t.Fatalf("second = %q, want left arrow", second.Data)
	}

	if _, err := decoder.Read(); err != io.EOF {
		t.Fatalf("final err = %v, want EOF", err)
	}
}

func TestInputDecoderRewrapsBracketedPaste(t *testing.T) {
	decoder := New(strings.NewReader(bracketedPasteStart + "hello\nworld" + bracketedPasteEnd))

	event, err := decoder.Read()
	if err != nil {
		t.Fatalf("read paste: %v", err)
	}
	want := bracketedPasteStart + "hello\nworld" + bracketedPasteEnd
	if event.Data != want {
		t.Fatalf("event = %q, want paste event", event.Data)
	}
}
