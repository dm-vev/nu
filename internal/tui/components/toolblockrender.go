package components

import "github.com/dm-vev/nu/internal/tui/message"

// Render returns a Pi-like tool execution box.
func (b *ToolBlock) Render(width int) []string {
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

	container := NewBox(BoxOptions{
		PaddingX: b.opts.PaddingX,
		PaddingY: b.opts.PaddingY,
		Bg:       bg,
	})
	container.AddChild(NewText(content, TextOptions{}))
	return container.Render(width)
}
