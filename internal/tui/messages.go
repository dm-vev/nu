package tui

import (
	"nu/internal/tui/components/assistantmessage"
	"nu/internal/tui/components/spacer"
	"nu/internal/tui/components/usermessage"
	tuimessage "nu/internal/tui/message"
)

func (a *App) rebuildChatLocked() {
	a.chat.Clear()
	for _, item := range a.messages {
		switch item.Role {
		case tuimessage.RoleUser:
			a.chat.AddChild(spacer.New(1))
			a.chat.AddChild(usermessage.New(firstText(item), usermessage.Options{
				PaddingX:      1,
				PaddingY:      1,
				TextStyle:     ansiText,
				StrongStyle:   boldText,
				EmphasisStyle: italicText,
				CodeStyle:     inlineCode,
				Background:    userBackground,
			}))
		case tuimessage.RoleAssistant:
			a.chat.AddChild(assistantmessage.NewMessage(item, assistantmessage.Options{
				PaddingX:       1,
				PaddingY:       0,
				TextStyle:      ansiText,
				HeadingStyle:   greenBold,
				StrongStyle:    boldText,
				EmphasisStyle:  italicText,
				CodeStyle:      inlineCode,
				ThinkingStyle:  thinkingText,
				ThinkingStrong: thinkingStrong,
				ToolPendingBg:  toolPendingBackground,
				ToolSuccessBg:  toolSuccessBackground,
				ToolErrorBg:    toolErrorBackground,
				ToolTitle:      boldText,
				ToolText:       ansiText,
				ToolErrorText:  red,
				ToolAdded:      green,
				ToolRemoved:    red,
				ToolContext:    muted,
			}))
		}
	}
	opts := a.footer.Options()
	opts.Used = estimateContextTokens(a.messages)
	a.footer.SetOptions(opts)
}

func estimateContextTokens(messages []tuimessage.Message) int {
	runes := 0
	for _, message := range messages {
		for _, part := range message.Parts {
			runes += len([]rune(part.Text))
			runes += len([]rune(part.ToolArguments))
			runes += len([]rune(part.ToolResult))
		}
	}
	if runes == 0 {
		return 0
	}
	// ponytail: display-only estimate until provider usage events are wired into TUI state.
	return max(1, (runes+3)/4)
}

func (a *App) appendAssistantDeltaLocked(delta string) {
	if delta == "" {
		return
	}
	last := len(a.messages) - 1
	if last >= 0 && a.messages[last].Role == tuimessage.RoleAssistant {
		a.messages[last].AppendText(delta)
		return
	}
	msg := tuimessage.NewAssistant()
	msg.AppendText(delta)
	a.messages = append(a.messages, msg)
}

func (a *App) appendAssistantThinkingLocked(delta string) {
	if delta == "" {
		return
	}
	last := len(a.messages) - 1
	if last >= 0 && a.messages[last].Role == tuimessage.RoleAssistant {
		a.messages[last].AppendThinking(delta)
		return
	}
	msg := tuimessage.NewAssistant()
	msg.AppendThinking(delta)
	a.messages = append(a.messages, msg)
}

func (a *App) replaceLastAssistantLocked(value string) {
	last := len(a.messages) - 1
	if last >= 0 && a.messages[last].Role == tuimessage.RoleAssistant {
		if hasToolPart(a.messages[last]) {
			return
		}
		a.messages[last].ReplaceText(value)
		return
	}
	a.messages = append(a.messages, tuimessage.NewAssistantText(value))
}

func (a *App) appendToolLocked(id string, name string, arguments string) {
	last := len(a.messages) - 1
	if last < 0 || a.messages[last].Role != tuimessage.RoleAssistant {
		a.messages = append(a.messages, tuimessage.NewAssistant())
		last = len(a.messages) - 1
	}
	a.messages[last].AddTool(id, name, arguments)
}

func (a *App) finishToolLocked(id string, result string, failed bool) {
	last := len(a.messages) - 1
	if last < 0 || a.messages[last].Role != tuimessage.RoleAssistant {
		return
	}
	a.messages[last].FinishTool(id, result, failed)
}

func firstText(value tuimessage.Message) string {
	for _, part := range value.Parts {
		if part.Kind == tuimessage.PartText {
			return part.Text
		}
	}
	return ""
}

func hasToolPart(value tuimessage.Message) bool {
	for _, part := range value.Parts {
		if part.Kind == tuimessage.PartTool {
			return true
		}
	}
	return false
}
