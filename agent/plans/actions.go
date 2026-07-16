package plans

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
)

// ExtractPlanAction returns a parsed plan action and the remaining input.
func ExtractPlanAction(input string) (string, string, string) {
	return "", "", input
}

// extractPlanAction attempts to extract a plan action from the user input
// Returns taskID, action, and remaining input
func (s *Service) extractPlanAction(input string) (string, string, string) {
	return ExtractPlanAction(input)
}

// handlePlanAction handles actions related to an existing plan
func (s *Service) HandlePlanAction(ctx context.Context, taskID, action, input string) (string, error) {
	plan, exists := s.store.GetPlanByTaskID(taskID)
	if !exists {
		return "", fmt.Errorf("plan with task ID %s not found", taskID)
	}

	switch action {
	case "approve":
		return s.approvePlan(ctx, plan)
	case "modify":
		return s.modifyPlan(ctx, plan, input)
	case "cancel":
		return s.cancelPlan(plan)
	case "status":
		return s.getPlanStatus(plan)
	default:
		return "", fmt.Errorf("unknown plan action: %s", action)
	}
}

// approvePlan approves and executes a plan
func (s *Service) approvePlan(ctx context.Context, plan *ExecutionPlan) (string, error) {
	plan.UserApproved = true
	plan.Status = ExecutionPlanStatusApproved

	// Add the approval to memory
	if s.memory != nil {
		if err := s.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleUser,
			Content: "I approve the plan. Please proceed with execution.",
		}); err != nil {
			return "", fmt.Errorf("failed to add approval to memory: %w", err)
		}
	}

	// Execute the plan
	result, err := s.executor.ExecutePlan(ctx, plan)
	if err != nil {
		return "", fmt.Errorf("failed to execute plan: %w", err)
	}

	// Add the execution result to memory
	if s.memory != nil {
		if err := s.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: result,
		}); err != nil {
			return "", fmt.Errorf("failed to add execution result to memory: %w", err)
		}
	}

	return result, nil
}

// modifyPlan modifies a plan based on user input
func (s *Service) modifyPlan(ctx context.Context, plan *ExecutionPlan, input string) (string, error) {
	// Add the modification request to memory
	if s.memory != nil {
		if err := s.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleUser,
			Content: "I'd like to modify the plan: " + input,
		}); err != nil {
			return "", fmt.Errorf("failed to add modification request to memory: %w", err)
		}
	}

	// Modify the plan
	modifiedPlan, err := s.generator.ModifyExecutionPlan(ctx, plan, input)
	if err != nil {
		return "", fmt.Errorf("failed to modify plan: %w", err)
	}

	// Update the plan in the store
	s.store.StorePlan(modifiedPlan)

	// Format the modified plan
	formattedPlan := FormatExecutionPlan(modifiedPlan)

	// Add the modified plan to memory
	if s.memory != nil {
		if err := s.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: "I've updated the execution plan based on your feedback:\n\n" + formattedPlan + "\nDo you approve this plan? You can modify it further if needed.",
		}); err != nil {
			return "", fmt.Errorf("failed to add modified plan to memory: %w", err)
		}
	}

	return "I've updated the execution plan based on your feedback:\n\n" + formattedPlan + "\nDo you approve this plan? You can modify it further if needed.", nil
}

// cancelPlan cancels a plan
func (s *Service) cancelPlan(plan *ExecutionPlan) (string, error) {
	s.executor.CancelPlan(plan)

	return "Plan cancelled. What would you like to do instead?", nil
}

// getPlanStatus returns the status of a plan
func (s *Service) getPlanStatus(plan *ExecutionPlan) (string, error) {
	status := s.executor.GetPlanStatus(plan)
	formattedPlan := FormatExecutionPlan(plan)

	return fmt.Sprintf("Current plan status: %s\n\n%s", status, formattedPlan), nil
}
