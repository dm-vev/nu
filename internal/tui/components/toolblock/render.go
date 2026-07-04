package toolblock

import (
	"nu/internal/tui/components/box"
	"nu/internal/tui/components/text"
	"nu/internal/tui/message"
)

// Render returns a Pi-like tool execution box.
func (b *Block) Render(width int) []string {
	content := b.formatContent()
	if content == "" {
		return nil
	}

	bg := b.opts.SuccessBg
	if b.state == message.ToolPending {
		bg = b.opts.PendingBg
	} else if b.state == message.ToolError || b.resultLooksFailed() {
		bg = b.opts.ErrorBg
	}

	container := box.New(box.Options{
		PaddingX: b.opts.PaddingX,
		PaddingY: b.opts.PaddingY,
		Bg:       bg,
	})
	container.AddChild(text.New(content, text.Options{}))
	return container.Render(width)
}
