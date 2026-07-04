package engine

// Options configures TUI rendering.
type Options struct {
	Title            string
	ShowCursor       bool
	ClearOnShrink    bool
	MinRenderRows    int
	InitialClear     bool
	SynchronizedDraw bool
}
