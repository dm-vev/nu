package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/internal/transport/grpc/pb"
)

// RunStream executes the remote agent with streaming response
func (r *Client) RunStream(ctx context.Context, input string) (<-chan contracts.AgentStreamEvent, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	// Create request
	req := &pb.RunRequest{
		Input:   input,
		Context: make(map[string]string),
	}

	// Add org_id from context if available
	if orgID, _ := multitenancy.GetOrgID(ctx); orgID != "" {
		req.OrgId = orgID
	}

	// Add conversation_id from context if available
	if conversationID, ok := conversation.GetConversationID(ctx); ok && conversationID != "" {
		req.ConversationId = conversationID
	}

	// Add timeout to context
	ctx, cancel := r.withTimeoutIfSet(ctx)

	// Execute streaming call
	stream, err := r.client.RunStream(ctx, req)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	// Create event channel
	eventChan := make(chan contracts.AgentStreamEvent, 100)

	// Start goroutine to handle streaming response
	go func() {
		defer cancel()
		defer close(eventChan)

		// Recover from any panics in the streaming goroutine
		defer func() {
			if r := recover(); r != nil {
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("stream panic recovered: %v", r),
					Timestamp: time.Now(),
				}
			}
		}()

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Stream completed normally
					eventChan <- contracts.AgentStreamEvent{
						Type:      contracts.AgentEventComplete,
						Timestamp: time.Now(),
					}
					return
				}
				// Stream error
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("stream error: %w", err),
					Timestamp: time.Now(),
				}
				return
			}

			// Check for nil response to prevent panic
			if resp == nil {
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("received nil response from stream"),
					Timestamp: time.Now(),
				}
				return
			}

			if resp.Error != "" {
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("remote agent error: %s", resp.Error),
					Timestamp: time.Now(),
				}
				return
			}

			// Convert gRPC response to AgentStreamEvent
			event := convertPbToStreamEvent(resp)
			eventChan <- event
		}
	}()

	return eventChan, nil
}

// RunStreamWithAuth executes the remote agent with streaming response and explicit auth token
func (r *Client) RunStreamWithAuth(ctx context.Context, input string, authToken string) (<-chan contracts.AgentStreamEvent, error) {
	if err := r.ensureConnected(); err != nil {
		return nil, err
	}

	// Create request
	req := &pb.RunRequest{
		Input:   input,
		Context: make(map[string]string),
	}

	// Add org_id from context if available
	if orgID, _ := multitenancy.GetOrgID(ctx); orgID != "" {
		req.OrgId = orgID
	}

	// Add conversation_id from context if available
	if conversationID, ok := conversation.GetConversationID(ctx); ok && conversationID != "" {
		req.ConversationId = conversationID
	}

	// Add explicit auth token to gRPC metadata
	if authToken != "" {
		md := metadata.Pairs("authorization", "Bearer "+authToken)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Add timeout to context
	ctx, cancel := r.withTimeoutIfSet(ctx)

	// Execute streaming call
	stream, err := r.client.RunStream(ctx, req)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	// Create event channel
	eventChan := make(chan contracts.AgentStreamEvent, 100)

	// Start goroutine to handle streaming response
	go func() {
		defer cancel()
		defer close(eventChan)

		// Recover from any panics in the streaming goroutine
		defer func() {
			if r := recover(); r != nil {
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("stream panic recovered: %v", r),
					Timestamp: time.Now(),
				}
			}
		}()

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Stream completed normally
					eventChan <- contracts.AgentStreamEvent{
						Type:      contracts.AgentEventComplete,
						Timestamp: time.Now(),
					}
					return
				}
				// Stream error
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("stream error: %w", err),
					Timestamp: time.Now(),
				}
				return
			}

			// Check for nil response to prevent panic
			if resp == nil {
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("received nil response from stream"),
					Timestamp: time.Now(),
				}
				return
			}

			if resp.Error != "" {
				eventChan <- contracts.AgentStreamEvent{
					Type:      contracts.AgentEventError,
					Error:     fmt.Errorf("remote agent error: %s", resp.Error),
					Timestamp: time.Now(),
				}
				return
			}

			// Convert gRPC response to AgentStreamEvent
			event := convertPbToStreamEvent(resp)
			eventChan <- event
		}
	}()

	return eventChan, nil
}
