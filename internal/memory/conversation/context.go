package conversation

import (
	"context"

	"github.com/dm-vev/nu/telemetry"
)

// ConversationIDKey is the key used to store conversation ID in context
const ConversationIDKey = telemetry.ConversationIDKey

// WithConversationID adds a conversation ID to the context
func WithConversationID(ctx context.Context, conversationID string) context.Context {
	return telemetry.WithConversationID(ctx, conversationID)
}

// GetConversationID retrieves the conversation ID from the context
func GetConversationID(ctx context.Context) (string, bool) {
	return telemetry.GetConversationID(ctx)
}
