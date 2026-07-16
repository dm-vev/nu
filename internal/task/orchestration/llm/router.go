package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

// HandoffRouter uses an LLM to choose the agent for a request.
type HandoffRouter struct {
	llm    contracts.LLM
	logger telemetry.Logger
}

// NewHandoffRouter creates an LLM-backed handoff router.
func NewHandoffRouter(llm contracts.LLM) *HandoffRouter {
	return &HandoffRouter{
		llm:    llm,
		logger: telemetry.NewLogger(), // Default logger
	}
}

// WithLogger sets the logger for the router
func (r *HandoffRouter) WithLogger(logger telemetry.Logger) *HandoffRouter {
	r.logger = logger
	return r
}

// Route determines which agent should handle a request
func (r *HandoffRouter) Route(ctx context.Context, query string, context map[string]interface{}) (string, error) {
	r.logger.Debug(ctx, "Routing query", map[string]interface{}{
		"query": query,
	})

	// Create a prompt for the LLM
	prompt := fmt.Sprintf(`You are a router that determines which specialized agent should handle a user query.
Available agents:
%s

User query: %s

Respond with only the ID of the agent that should handle this query.`, formatAgents(context["agents"].(map[string]string)), query)

	r.logger.Debug(ctx, "Generated routing prompt", map[string]interface{}{
		"prompt": prompt,
	})

	// Generate a response
	response, err := r.llm.Generate(ctx, prompt)
	if err != nil {
		r.logger.Error(ctx, "Failed to generate routing response", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	// Clean up the response
	response = strings.TrimSpace(response)
	r.logger.Debug(ctx, "Received routing response", map[string]interface{}{
		"raw_response": response,
	})

	// Validate the response
	if _, ok := context["agents"].(map[string]string)[response]; !ok {
		r.logger.Error(ctx, "Invalid agent ID returned by router", map[string]interface{}{
			"agent_id": response,
		})
		return "", fmt.Errorf("invalid agent ID: %s", response)
	}

	r.logger.Info(ctx, "Query routed to agent", map[string]interface{}{
		"agent_id": response,
		"query":    query,
	})

	return response, nil
}

// formatAgents formats a map of agent IDs to descriptions
func formatAgents(agents map[string]string) string {
	var result strings.Builder
	for id, desc := range agents {
		fmt.Fprintf(&result, "- %s: %s\n", id, desc)
	}
	return result.String()
}
