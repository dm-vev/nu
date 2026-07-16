package agent_test

import (
	"testing"
	"time"

	"nu/internal/agent"
	memory "nu/internal/memory/conversation"
	toolregistry "nu/internal/tools/registry"
)

func TestAgentContext(t *testing.T) {
	// Create a new context
	ctx := agent.NewAgentContext()

	// Test organization ID
	orgID := "test-org"
	ctx = ctx.WithOrganizationID(orgID)
	retrievedOrgID, ok := ctx.OrganizationID()
	if !ok {
		t.Error("Expected organization ID to be set")
	}
	if retrievedOrgID != orgID {
		t.Errorf("Expected organization ID to be %s, got %s", orgID, retrievedOrgID)
	}

	// Test conversation ID
	conversationID := "test-conversation"
	ctx = ctx.WithConversationID(conversationID)
	retrievedConversationID, ok := ctx.ConversationID()
	if !ok {
		t.Error("Expected conversation ID to be set")
	}
	if retrievedConversationID != conversationID {
		t.Errorf("Expected conversation ID to be %s, got %s", conversationID, retrievedConversationID)
	}

	// Test user ID
	userID := "test-user"
	ctx = ctx.WithUserID(userID)
	retrievedUserID, ok := ctx.UserID()
	if !ok {
		t.Error("Expected user ID to be set")
	}
	if retrievedUserID != userID {
		t.Errorf("Expected user ID to be %s, got %s", userID, retrievedUserID)
	}

	// Test request ID
	requestID := "test-request"
	ctx = ctx.WithRequestID(requestID)
	retrievedRequestID, ok := ctx.RequestID()
	if !ok {
		t.Error("Expected request ID to be set")
	}
	if retrievedRequestID != requestID {
		t.Errorf("Expected request ID to be %s, got %s", requestID, retrievedRequestID)
	}

	// Test memory
	memory := memory.NewConversationBuffer()
	ctx = ctx.WithMemory(memory)
	retrievedMemory, ok := ctx.Memory()
	if !ok {
		t.Error("Expected memory to be set")
	}
	if retrievedMemory != memory {
		t.Error("Expected memory to be the same instance")
	}

	// Test tools
	toolRegistry := toolregistry.NewRegistry()
	ctx = ctx.WithTools(toolRegistry)
	retrievedTools, ok := ctx.Tools()
	if !ok {
		t.Error("Expected tools to be set")
	}
	if retrievedTools != toolRegistry {
		t.Error("Expected tools to be the same instance")
	}

	// Test environment
	ctx = ctx.WithEnvironment("test-key", "test-value")
	retrievedValue, ok := ctx.Environment("test-key")
	if !ok {
		t.Error("Expected environment variable to be set")
	}
	if retrievedValue != "test-value" {
		t.Errorf("Expected environment variable to be %s, got %v", "test-value", retrievedValue)
	}

	// Test timeout
	timeout := 5 * time.Second
	ctxWithTimeout, cancel := ctx.WithTimeout(timeout)
	defer cancel()
	deadline, ok := ctxWithTimeout.Deadline()
	if !ok {
		t.Error("Expected deadline to be set")
	}
	if time.Until(deadline) > timeout {
		t.Error("Expected deadline to be within timeout")
	}

	// Test cancel
	ctxWithCancel, cancel := ctx.WithCancel()
	cancel()
	select {
	case <-ctxWithCancel.Done():
		// Expected
	default:
		t.Error("Expected context to be canceled")
	}
}
