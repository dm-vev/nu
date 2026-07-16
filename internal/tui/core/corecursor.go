package core

// CursorMarker is a zero-width marker stripped by the engine before writing.
const CursorMarker = "\x1b_pi:c\x07"

// Cursor stores a logical cursor position.
type Cursor struct {
	Row int
	Col int
}
