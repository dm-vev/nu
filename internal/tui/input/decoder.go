package input

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Decoder turns a raw byte stream into terminal key/input events.
type Decoder struct {
	reader *bufio.Reader
}

// New creates a decoder over reader.
func New(reader io.Reader) *Decoder {
	return &Decoder{reader: bufio.NewReader(reader)}
}

// Read returns the next decoded event.
func (d *Decoder) Read() (Event, error) {
	first, err := d.reader.ReadByte()
	if err != nil {
		return Event{}, err
	}
	if first == 0x1b {
		return d.readEscape()
	}
	if first < utf8.RuneSelf {
		return Event{Data: string(first)}, nil
	}
	return d.readUTF8(first)
}

func (d *Decoder) readUTF8(first byte) (Event, error) {
	buffer := []byte{first}
	for !utf8.FullRune(buffer) {
		next, err := d.reader.ReadByte()
		if err != nil {
			return Event{}, fmt.Errorf("read utf8 input: %w", err)
		}
		buffer = append(buffer, next)
	}
	return Event{Data: string(buffer)}, nil
}

func (d *Decoder) readEscape() (Event, error) {
	var builder strings.Builder
	builder.WriteByte(0x1b)

	// Escape sequences usually arrive in one terminal packet; consume the buffered suffix.
	for d.reader.Buffered() > 0 {
		next, err := d.reader.ReadByte()
		if err != nil {
			return Event{}, fmt.Errorf("read escape input: %w", err)
		}
		builder.WriteByte(next)
		if isKnownEscapeEnd(builder.String()) {
			break
		}
	}
	value := builder.String()
	if value == bracketedPasteStart {
		paste, err := d.readPaste()
		if err != nil {
			return Event{}, err
		}
		return Event{Data: bracketedPasteStart + paste + bracketedPasteEnd}, nil
	}
	return Event{Data: value}, nil
}

func isKnownEscapeEnd(value string) bool {
	if strings.HasSuffix(value, "~") || strings.HasSuffix(value, "u") {
		return true
	}
	if len(value) >= 3 {
		last := value[len(value)-1]
		return last >= '@' && last <= '~'
	}
	return false
}
