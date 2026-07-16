package tui

import agent "nu/internal/agentui"

// Emit updates UI state from one agent event.
func (a *App) Emit(ev agent.Event) {
	a.mu.Lock()
	switch ev.Type {
	case "turn_start":
		a.setFooterNoticeLocked("")
		a.status.SetText("running")
	case "message_update":
		a.setFooterNoticeLocked("")
		a.status.SetText("bubbling")
		if thinking := eventText(ev.Data, "thinking_delta"); thinking != "" {
			a.appendAssistantThinkingLocked(thinking)
		} else if eventText(ev.Data, "kind") == "thinking" {
			a.appendAssistantThinkingLocked(eventText(ev.Data, "delta"))
		} else {
			a.appendAssistantDeltaLocked(eventText(ev.Data, "delta"))
		}
	case "turn_end":
		a.setFooterNoticeLocked("")
		a.status.SetText("")
		if value := eventText(ev.Data, "text"); value != "" {
			a.replaceLastAssistantLocked(value)
		}
	case "tool_start":
		a.status.SetText("running tool")
		a.appendToolLocked(
			eventText(ev.Data, "id"),
			eventText(ev.Data, "name"),
			eventText(ev.Data, "arguments"),
		)
	case "tool_end":
		a.status.SetText("running")
		a.finishToolLocked(
			eventText(ev.Data, "id"),
			eventText(ev.Data, "result"),
			eventBool(ev.Data, "error"),
		)
	case "rate_limit":
		a.setFooterNoticeLocked("Rate limit")
		a.status.SetAlertText("Retrying")
	}
	a.rebuildChatLocked()
	a.mu.Unlock()
	a.render()
}

func (a *App) setFooterNoticeLocked(value string) {
	opts := a.footer.Options()
	opts.Notice = value
	a.footer.SetOptions(opts)
}

func eventBool(data any, key string) bool {
	values, ok := data.(map[string]string)
	if ok {
		return values[key] == "true"
	}
	generic, ok := data.(map[string]any)
	if ok {
		value, _ := generic[key].(bool)
		return value
	}
	return false
}

func eventText(data any, key string) string {
	values, ok := data.(map[string]string)
	if ok {
		return values[key]
	}
	generic, ok := data.(map[string]any)
	if ok {
		text, _ := generic[key].(string)
		return text
	}
	return ""
}
