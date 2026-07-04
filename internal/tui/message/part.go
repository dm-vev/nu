package message

// PartKind identifies how one message part must be rendered.
type PartKind string

const (
	// PartText is Markdown-capable visible message text.
	PartText PartKind = "text"

	// PartThinking is model reasoning shown dim and italic.
	PartThinking PartKind = "thinking"

	// PartTool is a command, patch, or tool execution block.
	PartTool PartKind = "tool"
)

// ToolState describes the current execution state of a tool block.
type ToolState string

const (
	// ToolPending is used while the command is still running.
	ToolPending ToolState = "pending"

	// ToolSuccess is used for completed tool calls.
	ToolSuccess ToolState = "success"

	// ToolError is used for failed tool calls or failed command exit status.
	ToolError ToolState = "error"
)

// Part is one renderable unit inside a chat message.
type Part struct {
	Kind PartKind

	Text string

	ToolID        string
	ToolName      string
	ToolArguments string
	ToolResult    string
	ToolState     ToolState
}
