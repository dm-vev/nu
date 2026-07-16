package components

const (
	userMessageOsc133ZoneStart = "\x1b]133;A\x07"
	userMessageOsc133ZoneEnd   = "\x1b]133;B\x07"
	userMessageOsc133ZoneFinal = "\x1b]133;C\x07"
)

// Render returns a padded, colored user message block.
func (m *UserMessage) Render(width int) []string {
	content := NewBox(BoxOptions{PaddingX: m.opts.PaddingX, PaddingY: m.opts.PaddingY, Bg: m.opts.Background})
	content.AddChild(NewMarkdown(m.text, MarkdownOptions{
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
	lines[0] = userMessageOsc133ZoneStart + lines[0]
	lines[len(lines)-1] = userMessageOsc133ZoneEnd + userMessageOsc133ZoneFinal + lines[len(lines)-1]
	return lines
}
