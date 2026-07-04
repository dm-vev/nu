package terminal

import (
	"fmt"
	"io"
)

// Write writes raw terminal bytes.
func (t *Terminal) Write(data string) error {
	if _, err := io.WriteString(t.stdout, data); err != nil {
		return fmt.Errorf("write terminal: %w", err)
	}
	return nil
}

// HideCursor hides the hardware cursor.
func (t *Terminal) HideCursor() error {
	return t.Write(HideCursor)
}

// ShowCursor shows the hardware cursor.
func (t *Terminal) ShowCursor() error {
	return t.Write(ShowCursor)
}

// MoveBy moves cursor vertically.
func (t *Terminal) MoveBy(lines int) error {
	if lines > 0 {
		return t.Write(fmt.Sprintf("\x1b[%dB", lines))
	}
	if lines < 0 {
		return t.Write(fmt.Sprintf("\x1b[%dA", -lines))
	}
	return nil
}

// SetTitle sets terminal title.
func (t *Terminal) SetTitle(title string) error {
	return t.Write("\x1b]0;" + title + "\x07")
}
