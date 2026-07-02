package tui

import (
	"bytes"
	"unicode/utf8"
)

const (
	pasteStart = "\x1b[200~"
	pasteEnd   = "\x1b[201~"
)

// EventKind identifies one decoded input event.
type EventKind string

const (
	EventText    EventKind = "text"
	EventKey     EventKind = "key"
	EventPaste   EventKind = "paste"
	EventUnknown EventKind = "unknown"
)

// InputEvent is one decoded terminal input event.
type InputEvent struct {
	Kind EventKind
	Key  string
	Text string
	Raw  string
}

// Decoder decodes byte-stream terminal input.
type Decoder struct {
	pending []byte
}

// NewDecoder creates an empty input decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Write decodes all complete events in chunk.
func (d *Decoder) Write(chunk []byte) []InputEvent {
	d.pending = append(d.pending, chunk...)
	var events []InputEvent
	for len(d.pending) > 0 {
		if bytes.HasPrefix(d.pending, []byte(pasteStart)) {
			end := bytes.Index(d.pending[len(pasteStart):], []byte(pasteEnd))
			if end < 0 {
				// Bracketed paste can be arbitrarily split; hold all bytes until the terminator.
				return events
			}
			start := len(pasteStart)
			stop := start + end
			events = append(events, InputEvent{Kind: EventPaste, Text: string(d.pending[start:stop])})
			d.pending = d.pending[stop+len(pasteEnd):]
			continue
		}
		if d.pending[0] == 0x1b {
			event, ok, needMore := d.readEscape()
			if needMore {
				return events
			}
			if ok {
				events = append(events, event)
				continue
			}
		}
		if event, ok := d.readControl(); ok {
			events = append(events, event)
			continue
		}
		event, ok := d.readText()
		if !ok {
			return events
		}
		events = append(events, event)
	}
	return events
}

// Flush emits any pending incomplete bytes.
func (d *Decoder) Flush() []InputEvent {
	if len(d.pending) == 0 {
		return nil
	}
	raw := string(d.pending)
	d.pending = nil
	if utf8.ValidString(raw) {
		return []InputEvent{{Kind: EventText, Text: raw, Raw: raw}}
	}
	return []InputEvent{{Kind: EventUnknown, Raw: raw}}
}

func (d *Decoder) readEscape() (InputEvent, bool, bool) {
	known := map[string]string{
		"\x1b[A": "up",
		"\x1b[B": "down",
		"\x1b[C": "right",
		"\x1b[D": "left",
	}
	for seq, key := range known {
		if bytes.HasPrefix(d.pending, []byte(seq)) {
			d.pending = d.pending[len(seq):]
			return InputEvent{Kind: EventKey, Key: key, Raw: seq}, true, false
		}
	}
	if len(d.pending) == 1 || bytes.Equal(d.pending, []byte("\x1b[")) {
		return InputEvent{}, false, true
	}
	if bytes.HasPrefix(d.pending, []byte("\x1b[")) && len(d.pending) < 3 {
		return InputEvent{}, false, true
	}
	rawLen := 1
	if len(d.pending) > 1 {
		rawLen = 2
	}
	raw := string(d.pending[:rawLen])
	d.pending = d.pending[rawLen:]
	return InputEvent{Kind: EventUnknown, Raw: raw}, true, false
}

func (d *Decoder) readControl() (InputEvent, bool) {
	switch d.pending[0] {
	case '\r', '\n':
		d.pending = d.pending[1:]
		return InputEvent{Kind: EventKey, Key: "enter", Raw: "\n"}, true
	case 0x7f, 0x08:
		d.pending = d.pending[1:]
		return InputEvent{Kind: EventKey, Key: "backspace", Raw: "\b"}, true
	case 0x04:
		d.pending = d.pending[1:]
		return InputEvent{Kind: EventKey, Key: "ctrl+d", Raw: "\x04"}, true
	case 0x03:
		d.pending = d.pending[1:]
		return InputEvent{Kind: EventKey, Key: "ctrl+c", Raw: "\x03"}, true
	default:
		return InputEvent{}, false
	}
}

func (d *Decoder) readText() (InputEvent, bool) {
	if !utf8.FullRune(d.pending) {
		return InputEvent{}, false
	}
	var size int
	for size < len(d.pending) {
		if d.pending[size] == 0x1b || d.pending[size] < 0x20 || d.pending[size] == 0x7f {
			break
		}
		_, runeSize := utf8.DecodeRune(d.pending[size:])
		if runeSize == 0 || !utf8.FullRune(d.pending[size:]) {
			break
		}
		size += runeSize
	}
	if size == 0 {
		raw := string(d.pending[:1])
		d.pending = d.pending[1:]
		return InputEvent{Kind: EventUnknown, Raw: raw}, true
	}
	text := string(d.pending[:size])
	d.pending = d.pending[size:]
	return InputEvent{Kind: EventText, Text: text, Raw: text}, true
}
