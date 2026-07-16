package orchestration

import (
	"context"
	"fmt"
	"strings"
)

// generateFinalResponse generates the final response
func (o *OrchestratorLLM) generateFinalResponse(ctx context.Context, plan *OrchestratorPlan, results map[string]string) (string, error) {
	o.logger.Info(ctx, "Generating final response using agent", map[string]interface{}{"agent": plan.FinalAgentID})

	// Get the final agent
	finalAgent, ok := o.registry.Get(plan.FinalAgentID)
	if !ok {
		// If the specified final agent is not available, try to use a fallback
		o.logger.Info(ctx, "Final agent not found, trying to use a fallback", map[string]interface{}{"agent": plan.FinalAgentID})

		// Try to use summary agent as fallback
		if summaryAgent, ok := o.registry.Get("summary"); ok {
			finalAgent = summaryAgent
			o.logger.Info(ctx, "Using summary agent as fallback for final response", nil)
		} else if creativeAgent, ok := o.registry.Get("creative"); ok {
			// Try creative agent as second fallback
			finalAgent = creativeAgent
			o.logger.Info(ctx, "Using creative agent as fallback for final response", nil)
		} else {
			// No suitable fallback found
			return "", fmt.Errorf("no suitable agent found for generating final response")
		}
	}

	// Create the final prompt
	var finalPrompt strings.Builder
	finalPrompt.WriteString("Based on the following information, provide a comprehensive response:\n\n")

	// Add the results from each step
	completedSteps := 0
	for i, step := range plan.Steps {
		stepID := fmt.Sprintf("step_%d", i)
		if result, ok := results[stepID]; ok {
			fmt.Fprintf(&finalPrompt, "--- %s (%s) ---\n%s\n\n", step.Description, step.AgentID, result)
			completedSteps++
		}
	}

	o.logger.Info(ctx, "Completed steps before generating final response", map[string]interface{}{"completed": completedSteps, "total": len(plan.Steps)})

	// Generate the final response
	finalResponse, err := finalAgent.Run(ctx, finalPrompt.String())
	if err != nil {
		o.logger.Error(ctx, "Failed to generate final response", map[string]interface{}{"error": err.Error()})
		return "", fmt.Errorf("failed to generate final response: %w", err)
	}

	o.logger.Info(ctx, "Final response generated successfully", nil)
	return finalResponse, nil
}
