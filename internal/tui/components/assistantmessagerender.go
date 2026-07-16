package components

import (
	"strings"

	"nu/internal/tui/core"
	"nu/internal/tui/message"
)

const (
	assistantMessageOsc133ZoneStart = "\x1b]133;A\x07"
	assistantMessageOsc133ZoneEnd   = "\x1b]133;B\x07"
	assistantMessageOsc133ZoneFinal = "\x1b]133;C\x07"
)

// Render returns a plain assistant message with Pi-like spacing.
func (m *AssistantMessage) Render(width int) []string {
	container := &core.Container{}
	if m.hasTextOrThinkingContent() {
		container.AddChild(NewSpacer(1))
	}
	for i, part := range m.parts {
		switch part.Kind {
		case message.PartText:
			if strings.TrimSpace(part.Text) == "" {
				continue
			}
			container.AddChild(NewMarkdown(strings.TrimSpace(part.Text), MarkdownOptions{
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
		case message.PartThinking:
			if strings.TrimSpace(part.Text) == "" {
				continue
			}
			container.AddChild(NewThinking(strings.TrimSpace(part.Text), ThinkingOptions{
				PaddingX:      m.opts.PaddingX,
				PaddingY:      m.opts.PaddingY,
				TextStyle:     m.opts.ThinkingStyle,
				StrongStyle:   m.opts.ThinkingStrong,
				EmphasisStyle: m.opts.ThinkingStyle,
				CodeStyle:     m.opts.CodeStyle,
			}))
			if m.hasTextOrThinkingContentAfter(i) {
				container.AddChild(NewSpacer(1))
			}
		case message.PartTool:
			container.AddChild(NewSpacer(1))
			container.AddChild(NewToolBlock(
				part.ToolName,
				part.ToolID,
				part.ToolArguments,
				part.ToolResult,
				part.ToolState,
				ToolBlockOptions{
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
	lines[0] = assistantMessageOsc133ZoneStart + lines[0]
	lines[len(lines)-1] = assistantMessageOsc133ZoneEnd + assistantMessageOsc133ZoneFinal + lines[len(lines)-1]
	return lines
}

func (m *AssistantMessage) hasTextOrThinkingContent() bool {
	for _, part := range m.parts {
		if part.Kind != message.PartTool && strings.TrimSpace(part.Text) != "" {
			return true
		}
	}
	return false
}

func (m *AssistantMessage) hasTextOrThinkingContentAfter(index int) bool {
	for _, part := range m.parts[index+1:] {
		if part.Kind != message.PartTool && strings.TrimSpace(part.Text) != "" {
			return true
		}
	}
	return false
}

func (m *AssistantMessage) hasToolParts() bool {
	for _, part := range m.parts {
		if part.Kind == message.PartTool {
			return true
		}
	}
	return false
}
