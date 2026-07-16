package orchestration

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nu/internal/telemetry"
)

// Orchestrator orchestrates handoffs between agents
type HandoffOrchestrator struct {
	registry *OrchestratorAgentRegistry
	router   HandoffRouter
	logger   telemetry.Logger
}

// NewOrchestrator creates a new orchestrator
func NewHandoffOrchestrator(registry *OrchestratorAgentRegistry, router HandoffRouter) *HandoffOrchestrator {
	return &HandoffOrchestrator{
		registry: registry,
		router:   router,
		logger:   telemetry.NewLogger(), // Default logger
	}
}

// WithLogger sets the logger for the orchestrator
func (o *HandoffOrchestrator) WithLogger(logger telemetry.Logger) *HandoffOrchestrator {
	o.logger = logger
	return o
}

// HandleRequest handles a request, potentially routing it through multiple agents
func (o *HandoffOrchestrator) HandleRequest(ctx context.Context, query string, initialContext map[string]interface{}) (*HandoffResult, error) {
	// Determine which agent should handle the request
	agentID, err := o.router.Route(ctx, query, initialContext)
	if err != nil {
		return nil, fmt.Errorf("failed to route request: %w", err)
	}

	o.logger.Info(ctx, "Initial routing decision", map[string]interface{}{
		"agent_id": agentID,
		"query":    query,
	})

	// Create initial handoff request
	handoffReq := &HandoffRequest{
		TargetAgentID:  agentID,
		Query:          query,
		Context:        initialContext,
		PreserveMemory: true,
	}

	// Process handoffs until completion or max iterations
	maxIterations := 5
	for i := 0; i < maxIterations; i++ {
		// Check if context is done
		select {
		case <-ctx.Done():
			o.logger.Warn(ctx, "Context deadline exceeded during handoff", map[string]interface{}{
				"iteration": i,
				"agent_id":  handoffReq.TargetAgentID,
			})
			return nil, ctx.Err()
		default:
			// Continue processing
		}

		// Process handoff
		result, err := o.processHandoff(ctx, handoffReq)
		if err != nil {
			o.logger.Error(ctx, "Failed to process handoff", map[string]interface{}{
				"error":    err.Error(),
				"agent_id": handoffReq.TargetAgentID,
				"query":    handoffReq.Query,
			})
			return nil, fmt.Errorf("failed to process handoff: %w", err)
		}

		// Check if completed or no next handoff
		if result.Completed || result.NextHandoff == nil {
			o.logger.Info(ctx, "Request completed", map[string]interface{}{
				"agent_id":  result.AgentID,
				"completed": result.Completed,
			})
			return result, nil
		}

		// Log handoff
		o.logger.Info(ctx, "Handoff detected", map[string]interface{}{
			"from_agent":   result.AgentID,
			"to_agent":     result.NextHandoff.TargetAgentID,
			"reason":       result.NextHandoff.Reason,
			"preserve_mem": result.NextHandoff.PreserveMemory,
		})

		// Prepare for next handoff
		handoffReq = result.NextHandoff
	}

	o.logger.Warn(ctx, "Exceeded maximum number of handoffs", map[string]interface{}{
		"max_iterations": maxIterations,
	})
	return nil, fmt.Errorf("exceeded maximum number of handoffs")
}

// processHandoff processes a single handoff
func (o *HandoffOrchestrator) processHandoff(ctx context.Context, req *HandoffRequest) (*HandoffResult, error) {
	// Get the target agent
	targetAgent, ok := o.registry.Get(req.TargetAgentID)
	if !ok {
		o.logger.Error(ctx, "Agent not found", map[string]interface{}{
			"agent_id": req.TargetAgentID,
		})
		return nil, fmt.Errorf("agent not found: %s", req.TargetAgentID)
	}

	o.logger.Info(ctx, "Processing request with agent", map[string]interface{}{
		"agent_id": req.TargetAgentID,
		"query":    req.Query,
	})

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Run the agent
	response, err := targetAgent.Run(ctx, req.Query)
	if err != nil {
		o.logger.Error(ctx, "Agent execution failed", map[string]interface{}{
			"agent_id": req.TargetAgentID,
			"error":    err.Error(),
		})
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Check for handoff request in the response
	nextHandoff := o.parseHandoffRequest(response)
	if nextHandoff != nil {
		o.logger.Debug(ctx, "Handoff request parsed from response", map[string]interface{}{
			"from_agent": req.TargetAgentID,
			"to_agent":   nextHandoff.TargetAgentID,
			"reason":     nextHandoff.Reason,
		})
	}

	// Create result
	result := &HandoffResult{
		AgentID:     req.TargetAgentID,
		Response:    response,
		Completed:   nextHandoff == nil,
		NextHandoff: nextHandoff,
	}

	return result, nil
}

// parseHandoffRequest parses a handoff request from an agent's response
func (o *HandoffOrchestrator) parseHandoffRequest(response string) *HandoffRequest {
	// Look for a handoff marker in the response
	// Format: [HANDOFF:agent_id:reason]
	re := regexp.MustCompile(`\[HANDOFF:([a-zA-Z0-9_-]+):([^\]]+)\]`)
	matches := re.FindStringSubmatch(response)
	if len(matches) < 3 {
		return nil
	}

	// Extract handoff information
	agentID := matches[1]
	reason := matches[2]

	// Extract the query (everything after the handoff marker)
	query := response[len(matches[0]):]
	query = strings.TrimSpace(query)

	// Create handoff request
	return &HandoffRequest{
		TargetAgentID:  agentID,
		Reason:         reason,
		Query:          query,
		PreserveMemory: true,
		Context:        make(map[string]interface{}),
	}
}
