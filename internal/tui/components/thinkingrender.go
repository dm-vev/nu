package components

// Render returns dim/italic Markdown lines for reasoning content.
func (t *Thinking) Render(width int) []string {
	md := NewMarkdown(t.text, MarkdownOptions{
		PaddingX:      t.opts.PaddingX,
		PaddingY:      t.opts.PaddingY,
		TextStyle:     t.opts.TextStyle,
		HeadingStyle:  t.opts.TextStyle,
		StrongStyle:   t.opts.StrongStyle,
		EmphasisStyle: t.opts.EmphasisStyle,
		CodeStyle:     t.opts.CodeStyle,
		QuoteStyle:    t.opts.TextStyle,
		BulletStyle:   t.opts.TextStyle,
	})
	return md.Render(width)
}
