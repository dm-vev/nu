package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// createPlan creates a plan for executing a query
func (o *OrchestratorLLM) createPlan(ctx context.Context, query string) (*OrchestratorPlan, error) {
	// Get available agents
	agents := o.registry.List()
	agentDescriptions := make(map[string]string)

	for id, agent := range agents {
		// Get agent description from system prompt using reflection
		agentValue := reflect.ValueOf(agent).Elem()
		systemPromptField := agentValue.FieldByName("systemPrompt")

		var description string
		if systemPromptField.IsValid() && systemPromptField.Kind() == reflect.String {
			systemPrompt := systemPromptField.String()
			// Extract first line as description
			description = strings.Split(systemPrompt, "\n")[0]
		} else {
			// Fallback to using the agent ID
			description = id
		}
		agentDescriptions[id] = description
	}

	// Create a prompt for the LLM
	prompt := fmt.Sprintf(`You are an orchestrator that creates plans to solve complex problems using multiple specialized agents.

Available agents:
%s

User query: %s

Create a plan to solve this query using the available agents. The plan should be a JSON object with the following structure:
{
  "steps": [
    {
      "agent_id": "string",
      "input": "string",
      "description": "string",
      "depends_on": ["step_0", "step_1"]
    }
  ],
  "final_agent_id": "string"
}

Each step should specify which agent to use, what input to provide, a description of the step's purpose, and any dependencies on other steps.

IMPORTANT RULES:
1. For dependencies, use the step index in the format "step_0", "step_1", etc. to refer to previous steps. Do not use agent IDs as dependencies.
2. To reference the output of a previous step in the input field, use the format {{step_0}}, {{step_1}}, etc.
3. Make sure all dependencies are valid - a step can only depend on steps with lower indices.
4. The final step should depend on all previous steps that contribute to the final answer.
5. The final_agent_id should specify which agent should provide the final response to the user.
6. Ensure the dependency chain is complete and there are no circular dependencies.

Respond with only the JSON plan.`, formatAgentDescriptions(agentDescriptions), query)

	// Generate a plan
	response, err := o.planner.Generate(ctx, prompt)
	if err != nil {
		o.logger.Error(ctx, "Failed to generate plan", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Extract JSON from response
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		o.logger.Error(ctx, "Failed to extract JSON from response", nil)
		return nil, fmt.Errorf("failed to extract JSON from response: %s", response)
	}

	// Parse the plan
	var plan OrchestratorPlan
	err = json.Unmarshal([]byte(jsonStr), &plan)
	if err != nil {
		o.logger.Error(ctx, "Failed to parse plan", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	// Validate and fix dependencies if needed
	for i := range plan.Steps {
		for j, dep := range plan.Steps[i].DependsOn {
			// If the dependency doesn't start with "step_", prepend it
			if !strings.HasPrefix(dep, "step_") {
				// Check if it's a numeric index
				if _, err := strconv.Atoi(dep); err == nil {
					plan.Steps[i].DependsOn[j] = "step_" + dep
					o.logger.Info(ctx, "Fixed dependency", map[string]interface{}{"from": dep, "to": plan.Steps[i].DependsOn[j]})
				}
			}
		}
	}

	// Ensure the final step depends on all previous steps
	if len(plan.Steps) > 0 {
		finalStepIndex := len(plan.Steps) - 1
		finalStep := &plan.Steps[finalStepIndex]

		// Create a map of existing dependencies for quick lookup
		existingDeps := make(map[string]bool)
		for _, dep := range finalStep.DependsOn {
			existingDeps[dep] = true
		}

		// Add dependencies for all previous steps if not already present
		for i := 0; i < finalStepIndex; i++ {
			depID := fmt.Sprintf("step_%d", i)
			if !existingDeps[depID] {
				finalStep.DependsOn = append(finalStep.DependsOn, depID)
				o.logger.Info(ctx, "Added missing dependency", map[string]interface{}{"dep": depID, "to": finalStep.DependsOn})
			}
		}
	}

	o.logger.Info(ctx, "Plan created with", map[string]interface{}{"steps": len(plan.Steps)})
	return &plan, nil
}

// formatAgentDescriptions formats agent descriptions for the prompt
func formatAgentDescriptions(descriptions map[string]string) string {
	var result strings.Builder
	for id, desc := range descriptions {
		fmt.Fprintf(&result, "- %s: %s\n", id, desc)
	}
	return result.String()
}
