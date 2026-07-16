package components

import "nu/internal/tui/message"

// ToolBlock renders one command, patch, or generic tool execution.
type ToolBlock struct {
	toolName  string
	toolID    string
	arguments string
	result    string
	state     message.ToolState
	opts      ToolBlockOptions
}

// NewToolBlock creates a tool block.
func NewToolBlock(toolName string, toolID string, arguments string, result string, state message.ToolState, opts ToolBlockOptions) *ToolBlock {
	return &ToolBlock{
		toolName:  toolName,
		toolID:    toolID,
		arguments: arguments,
		result:    result,
		state:     state,
		opts:      toolBlockNormalizeOptions(opts),
	}
}

// Invalidate exists for the component interface.
func (b *ToolBlock) Invalidate() {}
