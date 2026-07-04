package core

// Component is the Pi TUI render contract.
type Component interface {
	Render(width int) []string
}

// Filler renders flexible rows assigned by the engine.
type Filler interface {
	Component
	FillLines(width int, rows int) []string
}

// Invalidatable marks components with internal render caches.
type Invalidatable interface {
	Invalidate()
}

// Focusable receives raw input when focused.
type Focusable interface {
	Component
	HandleInput(data string)
	SetFocused(focused bool)
}
