package toolblock

import "nu/internal/tui/message"

// Block renders one command, patch, or generic tool execution.
type Block struct {
	toolName  string
	toolID    string
	arguments string
	result    string
	state     message.ToolState
	opts      Options
}

// New creates a tool block.
func New(toolName string, toolID string, arguments string, result string, state message.ToolState, opts Options) *Block {
	return &Block{
		toolName:  toolName,
		toolID:    toolID,
		arguments: arguments,
		result:    result,
		state:     state,
		opts:      normalizeOptions(opts),
	}
}

// Invalidate exists for the component interface.
func (b *Block) Invalidate() {}
