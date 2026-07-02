package tui

import "fmt"

// OverlayHandle identifies an overlay owned by the caller.
type OverlayHandle struct {
	ID    string
	Title string
}

// OverlayStack tracks focus order for modal overlays.
type OverlayStack struct {
	nextID   int
	stack    []OverlayHandle
	disposed map[string]bool
}

// NewOverlayStack creates an empty overlay stack.
func NewOverlayStack() *OverlayStack {
	return &OverlayStack{disposed: map[string]bool{}}
}

// Push adds an overlay and focuses it.
func (s *OverlayStack) Push(title string) OverlayHandle {
	s.nextID++
	handle := OverlayHandle{ID: fmt.Sprintf("overlay-%d", s.nextID), Title: title}
	s.stack = append(s.stack, handle)
	return handle
}

// Close removes an overlay and restores focus to the previous active one.
func (s *OverlayStack) Close(handle OverlayHandle) bool {
	if s.disposed[handle.ID] {
		return false
	}
	for i := len(s.stack) - 1; i >= 0; i-- {
		if s.stack[i].ID != handle.ID {
			continue
		}
		// Removing in place preserves the previous focus order for all older overlays.
		s.stack = append(s.stack[:i], s.stack[i+1:]...)
		s.disposed[handle.ID] = true
		return true
	}
	return false
}

// Focused returns the currently focused overlay.
func (s *OverlayStack) Focused() (OverlayHandle, bool) {
	if len(s.stack) == 0 {
		return OverlayHandle{}, false
	}
	return s.stack[len(s.stack)-1], true
}
