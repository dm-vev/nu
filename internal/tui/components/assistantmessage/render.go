package assistantmessage

import (
	"strings"

	"nu/internal/tui/components/markdown"
	"nu/internal/tui/components/spacer"
	"nu/internal/tui/components/thinking"
	"nu/internal/tui/components/toolblock"
	"nu/internal/tui/core"
	tuimessage "nu/internal/tui/message"
)

const (
	osc133ZoneStart = "\x1b]133;A\x07"
	osc133ZoneEnd   = "\x1b]133;B\x07"
	osc133ZoneFinal = "\x1b]133;C\x07"
)

// Render returns a plain assistant message with Pi-like spacing.
func (m *Message) Render(width int) []string {
	container := &core.Container{}
	if m.hasTextOrThinkingContent() {
		container.AddChild(spacer.New(1))
	}
	for i, part := range m.parts {
		switch part.Kind {
		case tuimessage.PartText:
			if strings.TrimSpace(part.Text) == "" {
				continue
			}
			container.AddChild(markdown.New(strings.TrimSpace(part.Text), markdown.Options{
				PaddingX:      m.opts.PaddingX,
				PaddingY:      m.opts.PaddingY,
				TextStyle:     m.opts.TextStyle,
				HeadingStyle:  m.opts.HeadingStyle,
				StrongStyle:   m.opts.StrongStyle,
				EmphasisStyle: m.opts.EmphasisStyle,
				CodeStyle:     m.opts.CodeStyle,
				QuoteStyle:    m.opts.TextStyle,
				BulletStyle:   m.opts.TextStyle,
			}))
		case tuimessage.PartThinking:
			if strings.TrimSpace(part.Text) == "" {
				continue
			}
			container.AddChild(thinking.New(strings.TrimSpace(part.Text), thinking.Options{
				PaddingX:      m.opts.PaddingX,
				PaddingY:      m.opts.PaddingY,
				TextStyle:     m.opts.ThinkingStyle,
				StrongStyle:   m.opts.ThinkingStrong,
				EmphasisStyle: m.opts.ThinkingStyle,
				CodeStyle:     m.opts.CodeStyle,
			}))
			if m.hasTextOrThinkingContentAfter(i) {
				container.AddChild(spacer.New(1))
			}
		case tuimessage.PartTool:
			container.AddChild(spacer.New(1))
			container.AddChild(toolblock.New(
				part.ToolName,
				part.ToolID,
				part.ToolArguments,
				part.ToolResult,
				part.ToolState,
				toolblock.Options{
					PaddingX:     1,
					PaddingY:     1,
					PendingBg:    m.opts.ToolPendingBg,
					SuccessBg:    m.opts.ToolSuccessBg,
					ErrorBg:      m.opts.ToolErrorBg,
					TitleStyle:   m.opts.ToolTitle,
					TextStyle:    m.opts.ToolText,
					ErrorStyle:   m.opts.ToolErrorText,
					AddedStyle:   m.opts.ToolAdded,
					RemovedStyle: m.opts.ToolRemoved,
					ContextStyle: m.opts.ToolContext,
				},
			))
		}
	}
	lines := container.Render(width)
	if len(lines) == 0 || m.hasToolParts() {
		return lines
	}
	lines[0] = osc133ZoneStart + lines[0]
	lines[len(lines)-1] = osc133ZoneEnd + osc133ZoneFinal + lines[len(lines)-1]
	return lines
}

func (m *Message) hasTextOrThinkingContent() bool {
	for _, part := range m.parts {
		if part.Kind != tuimessage.PartTool && strings.TrimSpace(part.Text) != "" {
			return true
		}
	}
	return false
}

func (m *Message) hasTextOrThinkingContentAfter(index int) bool {
	for _, part := range m.parts[index+1:] {
		if part.Kind != tuimessage.PartTool && strings.TrimSpace(part.Text) != "" {
			return true
		}
	}
	return false
}

func (m *Message) hasToolParts() bool {
	for _, part := range m.parts {
		if part.Kind == tuimessage.PartTool {
			return true
		}
	}
	return false
}
