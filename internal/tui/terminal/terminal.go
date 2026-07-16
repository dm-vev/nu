package terminal

import "io"

// Terminal owns raw terminal IO and dimensions.
type Terminal struct {
	stdin  io.Reader
	stdout io.Writer
	width  int
	height int
}

// New creates a terminal wrapper.
func New(stdin io.Reader, stdout io.Writer, width int, height int) *Terminal {
	if stdin == nil {
		stdin = io.Reader(nil)
	}
	if stdout == nil {
		stdout = io.Discard
	}
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return &Terminal{stdin: stdin, stdout: stdout, width: width, height: height}
}

// Stdin returns the input reader.
func (t *Terminal) Stdin() io.Reader {
	return t.stdin
}

// Size returns current terminal dimensions.
func (t *Terminal) Size() (int, int) {
	width, height := terminalQuerySize(t.stdout, t.width, t.height)
	t.width, t.height = width, height
	return width, height
}
