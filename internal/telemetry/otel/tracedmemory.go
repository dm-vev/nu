package otel

import (
	"context"

	"nu/internal/contracts"
)

// TracedMemory implements middleware for memory operations with unified tracing
type TracedMemory struct {
	memory contracts.Memory
	tracer contracts.Tracer
}

// NewTracedMemory creates a new memory middleware with unified tracing
func NewTracedMemory(memory contracts.Memory, tracer contracts.Tracer) *TracedMemory {
	return &TracedMemory{
		memory: memory,
		tracer: tracer,
	}
}

// AddMessage adds a message to memory with tracing
func (m *TracedMemory) AddMessage(ctx context.Context, message contracts.Message) error {
	// Start span
	ctx, span := m.tracer.StartSpan(ctx, "memory.add_message")
	defer span.End()

	// Add attributes
	span.SetAttribute("message.role", string(message.Role))
	span.SetAttribute("message.content_length", len(message.Content))
	span.SetAttribute("message.content_hash", hashString(message.Content))
	if len(message.ToolCalls) > 0 {
		span.SetAttribute("message.tool_calls_count", len(message.ToolCalls))
	}

	// Call the underlying memory
	err := m.memory.AddMessage(ctx, message)
	if err != nil {
		span.RecordError(err)
	}

	return err
}

// GetMessages gets messages from memory with tracing
func (m *TracedMemory) GetMessages(ctx context.Context, options ...contracts.GetMessagesOption) ([]contracts.Message, error) {
	// Start span
	ctx, span := m.tracer.StartSpan(ctx, "memory.get_messages")
	defer span.End()

	// Call the underlying memory
	messages, err := m.memory.GetMessages(ctx, options...)
	if err != nil {
		span.RecordError(err)
	} else {
		span.SetAttribute("messages.count", len(messages))
	}

	return messages, err
}

// Clear clears memory with tracing
func (m *TracedMemory) Clear(ctx context.Context) error {
	// Start span
	ctx, span := m.tracer.StartSpan(ctx, "memory.clear")
	defer span.End()

	// Call the underlying memory
	err := m.memory.Clear(ctx)
	if err != nil {
		span.RecordError(err)
	}

	return err
}
