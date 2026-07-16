package message

// AppendText appends visible assistant text, coalescing adjacent text parts.
func (m *Message) AppendText(delta string) {
	if delta == "" {
		return
	}
	if len(m.Parts) > 0 && m.Parts[len(m.Parts)-1].Kind == PartText {
		m.Parts[len(m.Parts)-1].Text += delta
		return
	}
	m.Parts = append(m.Parts, Part{Kind: PartText, Text: delta})
}

// AppendThinking appends model reasoning, coalescing adjacent thinking parts.
func (m *Message) AppendThinking(delta string) {
	if delta == "" {
		return
	}
	if len(m.Parts) > 0 && m.Parts[len(m.Parts)-1].Kind == PartThinking {
		m.Parts[len(m.Parts)-1].Text += delta
		return
	}
	m.Parts = append(m.Parts, Part{Kind: PartThinking, Text: delta})
}

// ReplaceText replaces the visible assistant text without deleting tool/thinking parts.
func (m *Message) ReplaceText(value string) {
	for i := len(m.Parts) - 1; i >= 0; i-- {
		if m.Parts[i].Kind != PartText {
			continue
		}
		m.Parts[i].Text = value
		return
	}
	m.Parts = append(m.Parts, Part{Kind: PartText, Text: value})
}

// AddTool appends a pending tool execution block.
func (m *Message) AddTool(id string, name string, arguments string) {
	m.Parts = append(m.Parts, Part{
		Kind:          PartTool,
		ToolID:        id,
		ToolName:      name,
		ToolArguments: arguments,
		ToolState:     ToolPending,
	})
}

// FinishTool marks a tool block complete and stores its output.
func (m *Message) FinishTool(id string, result string, failed bool) {
	for i := len(m.Parts) - 1; i >= 0; i-- {
		if m.Parts[i].Kind != PartTool {
			continue
		}
		if id != "" && m.Parts[i].ToolID != "" && m.Parts[i].ToolID != id {
			continue
		}
		m.Parts[i].ToolResult = result
		if failed {
			m.Parts[i].ToolState = ToolError
		} else {
			m.Parts[i].ToolState = ToolSuccess
		}
		return
	}
}
