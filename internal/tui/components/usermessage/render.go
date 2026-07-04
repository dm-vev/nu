package usermessage

import (
	"nu/internal/tui/components/box"
	"nu/internal/tui/components/markdown"
)

const (
	osc133ZoneStart = "\x1b]133;A\x07"
	osc133ZoneEnd   = "\x1b]133;B\x07"
	osc133ZoneFinal = "\x1b]133;C\x07"
)

// Render returns a padded, colored user message block.
func (m *Message) Render(width int) []string {
	content := box.New(box.Options{PaddingX: m.opts.PaddingX, PaddingY: m.opts.PaddingY, Bg: m.opts.Background})
	content.AddChild(markdown.New(m.text, markdown.Options{
		TextStyle:     m.opts.TextStyle,
		HeadingStyle:  m.opts.StrongStyle,
		StrongStyle:   m.opts.StrongStyle,
		EmphasisStyle: m.opts.EmphasisStyle,
		CodeStyle:     m.opts.CodeStyle,
		QuoteStyle:    m.opts.TextStyle,
		BulletStyle:   m.opts.TextStyle,
	}))
	lines := content.Render(width)
	if len(lines) == 0 {
		return lines
	}
	// OSC 133 zones let capable terminals treat a user turn as one prompt region.
	lines[0] = osc133ZoneStart + lines[0]
	lines[len(lines)-1] = osc133ZoneEnd + osc133ZoneFinal + lines[len(lines)-1]
	return lines
}
