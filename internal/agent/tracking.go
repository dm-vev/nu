package agent

import (
	"context"

	"nu/internal/contracts"
)

// trackingTool wraps a Tool and records each invocation with the usage tracker
// before delegating. Records the call when the LLM client actually invokes
// Execute or Run, so the execution summary reflects tools the model chose to
// call rather than every tool that was made available.
type trackingTool struct {
	inner   contracts.Tool
	tracker *usageTracker
}

func (t *trackingTool) Name() string                                   { return t.inner.Name() }
func (t *trackingTool) Description() string                            { return t.inner.Description() }
func (t *trackingTool) Parameters() map[string]contracts.ParameterSpec { return t.inner.Parameters() }

func (t *trackingTool) Run(ctx context.Context, input string) (string, error) {
	if t.tracker != nil {
		t.tracker.addToolCall(t.inner.Name())
	}
	return t.inner.Run(ctx, input)
}

func (t *trackingTool) Execute(ctx context.Context, args string) (string, error) {
	if t.tracker != nil {
		t.tracker.addToolCall(t.inner.Name())
	}
	return t.inner.Execute(ctx, args)
}

// DisplayName forwards to the inner tool when it implements ToolWithDisplayName.
func (t *trackingTool) DisplayName() string {
	if d, ok := t.inner.(contracts.ToolWithDisplayName); ok {
		return d.DisplayName()
	}
	return t.inner.Name()
}

// Internal forwards to the inner tool when it implements InternalTool.
func (t *trackingTool) Internal() bool {
	if i, ok := t.inner.(contracts.InternalTool); ok {
		return i.Internal()
	}
	return false
}

// wrapToolsWithTracker wraps each tool so its invocation is recorded with the
// tracker. Returns the original slice unchanged when tracker is nil.
func wrapToolsWithTracker(tools []contracts.Tool, tracker *usageTracker) []contracts.Tool {
	if tracker == nil || len(tools) == 0 {
		return tools
	}
	wrapped := make([]contracts.Tool, len(tools))
	for i, t := range tools {
		wrapped[i] = &trackingTool{inner: t, tracker: tracker}
	}
	return wrapped
}
