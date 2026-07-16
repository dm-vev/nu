package agent

import (
	"context"
	"fmt"

	"nu/internal/agent/plans"
	"nu/internal/contracts"
)

// extractPlanAction attempts to extract a plan action from the user input
// Returns taskID, action, and remaining input
func (a *Agent) extractPlanAction(input string) (string, string, string) {
	// This is a placeholder implementation
	// In a real implementation, you would use NLP or pattern matching to extract plan actions
	return "", "", input
}

// handlePlanAction handles actions related to an existing plan
func (a *Agent) handlePlanAction(ctx context.Context, taskID, action, input string) (string, error) {
	plan, exists := a.planStore.GetPlanByTaskID(taskID)
	if !exists {
		return "", fmt.Errorf("plan with task ID %s not found", taskID)
	}

	switch action {
	case "approve":
		return a.approvePlan(ctx, plan)
	case "modify":
		return a.modifyPlan(ctx, plan, input)
	case "cancel":
		return a.cancelPlan(plan)
	case "status":
		return a.getPlanStatus(plan)
	default:
		return "", fmt.Errorf("unknown plan action: %s", action)
	}
}

// approvePlan approves and executes a plan
func (a *Agent) approvePlan(ctx context.Context, plan *plans.ExecutionPlan) (string, error) {
	plan.UserApproved = true
	plan.Status = plans.ExecutionPlanStatusApproved

	// Add the approval to memory
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleUser,
			Content: "I approve the plan. Please proceed with execution.",
		}); err != nil {
			return "", fmt.Errorf("failed to add approval to memory: %w", err)
		}
	}

	// Execute the plan
	result, err := a.planExecutor.ExecutePlan(ctx, plan)
	if err != nil {
		return "", fmt.Errorf("failed to execute plan: %w", err)
	}

	// Add the execution result to memory
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: result,
		}); err != nil {
			return "", fmt.Errorf("failed to add execution result to memory: %w", err)
		}
	}

	return result, nil
}

// modifyPlan modifies a plan based on user input
func (a *Agent) modifyPlan(ctx context.Context, plan *plans.ExecutionPlan, input string) (string, error) {
	// Add the modification request to memory
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleUser,
			Content: "I'd like to modify the plan: " + input,
		}); err != nil {
			return "", fmt.Errorf("failed to add modification request to memory: %w", err)
		}
	}

	// Modify the plan
	modifiedPlan, err := a.planGenerator.ModifyExecutionPlan(ctx, plan, input)
	if err != nil {
		return "", fmt.Errorf("failed to modify plan: %w", err)
	}

	// Update the plan in the store
	a.planStore.StorePlan(modifiedPlan)

	// Format the modified plan
	formattedPlan := plans.FormatExecutionPlan(modifiedPlan)

	// Add the modified plan to memory
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: "I've updated the execution plan based on your feedback:\n\n" + formattedPlan + "\nDo you approve this plan? You can modify it further if needed.",
		}); err != nil {
			return "", fmt.Errorf("failed to add modified plan to memory: %w", err)
		}
	}

	return "I've updated the execution plan based on your feedback:\n\n" + formattedPlan + "\nDo you approve this plan? You can modify it further if needed.", nil
}

// cancelPlan cancels a plan
func (a *Agent) cancelPlan(plan *plans.ExecutionPlan) (string, error) {
	a.planExecutor.CancelPlan(plan)

	return "Plan cancelled. What would you like to do instead?", nil
}

// getPlanStatus returns the status of a plan
func (a *Agent) getPlanStatus(plan *plans.ExecutionPlan) (string, error) {
	status := a.planExecutor.GetPlanStatus(plan)
	formattedPlan := plans.FormatExecutionPlan(plan)

	return fmt.Sprintf("Current plan status: %s\n\n%s", status, formattedPlan), nil
}

// runWithExecutionPlan runs the agent with an execution plan
func (a *Agent) runWithExecutionPlan(ctx context.Context, input string) (string, error) {
	// Generate an execution plan
	plan, err := a.planGenerator.GenerateExecutionPlan(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to generate execution plan: %w", err)
	}

	// Store the plan
	a.planStore.StorePlan(plan)

	// Format the plan for display
	formattedPlan := plans.FormatExecutionPlan(plan)

	// Add the plan to memory
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: "I've created an execution plan for your request:\n\n" + formattedPlan + "\nDo you approve this plan? You can modify it if needed.",
		}); err != nil {
			return "", fmt.Errorf("failed to add plan to memory: %w", err)
		}
	}

	// Return the plan for user approval
	return "I've created an execution plan for your request:\n\n" + formattedPlan + "\nDo you approve this plan? You can modify it if needed.", nil
}

// ApproveExecutionPlan approves an execution plan for execution
func (a *Agent) ApproveExecutionPlan(ctx context.Context, plan *plans.ExecutionPlan) (string, error) {
	return a.approvePlan(ctx, plan)
}

// ModifyExecutionPlan modifies an execution plan based on user input
func (a *Agent) ModifyExecutionPlan(ctx context.Context, plan *plans.ExecutionPlan, modifications string) (*plans.ExecutionPlan, error) {
	return a.planGenerator.ModifyExecutionPlan(ctx, plan, modifications)
}

// GenerateExecutionPlan generates an execution plan
func (a *Agent) GenerateExecutionPlan(ctx context.Context, input string) (*plans.ExecutionPlan, error) {
	return a.planGenerator.GenerateExecutionPlan(ctx, input)
}

// GetTaskByID returns a task by its ID
func (a *Agent) GetTaskByID(taskID string) (*plans.ExecutionPlan, bool) {
	return a.planStore.GetPlanByTaskID(taskID)
}

// ListTasks returns a list of all tasks
func (a *Agent) ListTasks() []*plans.ExecutionPlan {
	return a.planStore.ListPlans()
}
