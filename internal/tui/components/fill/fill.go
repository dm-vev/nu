package fill

// Fill expands to the remaining terminal rows assigned by the engine.
type Fill struct{}

// New creates a flexible blank component.
func New() *Fill {
	return &Fill{}
}

// Render returns no fixed lines; the engine calls FillLines with remaining rows.
func (f *Fill) Render(width int) []string {
	return nil
}
