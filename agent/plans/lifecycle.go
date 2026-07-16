package plans

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
)

// RunWithExecutionPlan creates and stores a plan for user approval.
func (s *Service) RunWithExecutionPlan(ctx context.Context, input string) (string, error) {
	plan, err := s.generator.GenerateExecutionPlan(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to generate execution plan: %w", err)
	}
	s.store.StorePlan(plan)
	formattedPlan := FormatExecutionPlan(plan)
	if s.memory != nil {
		if err := s.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: "I've created an execution plan for your request:\n\n" + formattedPlan + "\nDo you approve the plan? You can modify it if needed.",
		}); err != nil {
			return "", fmt.Errorf("failed to add plan to memory: %w", err)
		}
	}
	return "I've created an execution plan for your request:\n\n" + formattedPlan + "\nDo you approve the plan? You can modify it if needed.", nil
}

// ApproveExecutionPlan approves an execution plan for execution.
func (s *Service) ApproveExecutionPlan(ctx context.Context, plan *ExecutionPlan) (string, error) {
	return s.approvePlan(ctx, plan)
}

// ModifyExecutionPlan modifies an execution plan based on user input.
func (s *Service) ModifyExecutionPlan(ctx context.Context, plan *ExecutionPlan, modifications string) (*ExecutionPlan, error) {
	return s.generator.ModifyExecutionPlan(ctx, plan, modifications)
}

// GenerateExecutionPlan generates an execution plan.
func (s *Service) GenerateExecutionPlan(ctx context.Context, input string) (*ExecutionPlan, error) {
	return s.generator.GenerateExecutionPlan(ctx, input)
}

// GetTaskByID returns a task by its ID.
func (s *Service) GetTaskByID(taskID string) (*ExecutionPlan, bool) {
	return s.store.GetPlanByTaskID(taskID)
}

// ListTasks returns all execution
func (s *Service) ListTasks() []*ExecutionPlan { return s.store.ListPlans() }
