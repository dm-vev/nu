package agent

import (
	"context"

	"github.com/dm-vev/nu/agent/plans"
)

func (a *Agent) extractPlanAction(input string) (string, string, string) {
	return plans.ExtractPlanAction(input)
}

func (a *Agent) handlePlanAction(ctx context.Context, taskID, action, input string) (string, error) {
	return a.planService.HandlePlanAction(ctx, taskID, action, input)
}

func (a *Agent) runWithExecutionPlan(ctx context.Context, input string) (string, error) {
	return a.planService.RunWithExecutionPlan(ctx, input)
}

// ApproveExecutionPlan approves an execution plan for execution.
func (a *Agent) ApproveExecutionPlan(ctx context.Context, plan *plans.ExecutionPlan) (string, error) {
	return a.planService.ApproveExecutionPlan(ctx, plan)
}

// ModifyExecutionPlan modifies an execution plan based on user input.
func (a *Agent) ModifyExecutionPlan(ctx context.Context, plan *plans.ExecutionPlan, modifications string) (*plans.ExecutionPlan, error) {
	return a.planService.ModifyExecutionPlan(ctx, plan, modifications)
}

// GenerateExecutionPlan generates an execution plan.
func (a *Agent) GenerateExecutionPlan(ctx context.Context, input string) (*plans.ExecutionPlan, error) {
	return a.planService.GenerateExecutionPlan(ctx, input)
}

// GetTaskByID returns a task by its ID.
func (a *Agent) GetTaskByID(taskID string) (*plans.ExecutionPlan, bool) {
	return a.planService.GetTaskByID(taskID)
}

// ListTasks returns all execution plans.
func (a *Agent) ListTasks() []*plans.ExecutionPlan {
	return a.planService.ListTasks()
}
