package task

import (
	"context"
	"time"
)

// SimpleExecutor implements task execution with minimal functionality.
type SimpleExecutor struct{}

// ExecuteStep marks a step as completed.
func (e *SimpleExecutor) ExecuteStep(ctx context.Context, task *Task, step *Step) error {
	step.Status = StepStatusCompleted
	now := time.Now()
	step.CompletedAt = &now
	return nil
}

// ExecuteTask marks every task step as completed.
func (e *SimpleExecutor) ExecuteTask(ctx context.Context, task *Task) error {
	for i := range task.Steps {
		task.Steps[i].Status = StepStatusCompleted
		now := time.Now()
		task.Steps[i].CompletedAt = &now
	}
	return nil
}
