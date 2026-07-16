package plans

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/contracts"
)

// ExecutionPlanExecutor handles execution of execution plans.
type ExecutionPlanExecutor struct {
	tools map[string]contracts.Tool
}

// NewExecutionPlanExecutor creates an execution plan executor.
func NewExecutionPlanExecutor(tools []contracts.Tool) *ExecutionPlanExecutor {
	toolMap := make(map[string]contracts.Tool)
	for _, tool := range tools {
		toolMap[tool.Name()] = tool
	}

	return &ExecutionPlanExecutor{
		tools: toolMap,
	}
}

// ExecutePlan executes an approved execution plan
func (e *ExecutionPlanExecutor) ExecutePlan(ctx context.Context, plan *ExecutionPlan) (string, error) {
	if !plan.UserApproved {
		return "", fmt.Errorf("execution plan has not been approved by the user")
	}

	// Update status to executing
	plan.Status = ExecutionPlanStatusExecuting

	// Execute each step in the plan
	results := make([]string, 0, len(plan.Steps))
	for i, step := range plan.Steps {
		// Get the tool
		tool, ok := e.tools[step.ToolName]
		if !ok {
			plan.Status = ExecutionPlanStatusFailed
			return "", fmt.Errorf("unknown tool: %s", step.ToolName)
		}

		// Marshal parameters to JSON for the Execute method
		// This ensures tools receive the expected JSON format
		var inputJSON string
		if len(step.Parameters) > 0 {
			jsonBytes, err := json.Marshal(step.Parameters)
			if err != nil {
				plan.Status = ExecutionPlanStatusFailed
				return "", fmt.Errorf("failed to marshal parameters for step %d: %w", i+1, err)
			}
			inputJSON = string(jsonBytes)
		} else if step.Input != "" {
			// Fallback to step.Input if no parameters are provided
			// This maintains backward compatibility
			inputJSON = step.Input
		} else {
			// If neither parameters nor input is provided, use empty JSON object
			inputJSON = "{}"
		}

		// Execute the tool with JSON input
		result, err := tool.Execute(ctx, inputJSON)
		if err != nil {
			plan.Status = ExecutionPlanStatusFailed
			return "", fmt.Errorf("failed to execute step %d: %w", i+1, err)
		}

		// Add the result to the list of results
		results = append(results, fmt.Sprintf("Step %d (%s): %s", i+1, step.Description, result))
	}

	// Update status to completed
	plan.Status = ExecutionPlanStatusCompleted

	// Format the results
	return fmt.Sprintf("Execution plan completed successfully!\n\n%s", strings.Join(results, "\n\n")), nil
}

// CancelPlan cancels an execution plan
func (e *ExecutionPlanExecutor) CancelPlan(plan *ExecutionPlan) {
	plan.Status = ExecutionPlanStatusCancelled
}

// GetPlanStatus returns the status of an execution plan
func (e *ExecutionPlanExecutor) GetPlanStatus(plan *ExecutionPlan) ExecutionPlanStatus {
	return plan.Status
}
