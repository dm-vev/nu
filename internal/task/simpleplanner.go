package task

import (
	"context"

	"nu/internal/telemetry"
)

// SimplePlanner implements TaskPlanner with minimal functionality
type SimplePlanner struct{}

// CreatePlan implements TaskPlanner.CreatePlan
func (p *SimplePlanner) CreatePlan(ctx context.Context, task *Task) (string, error) {
	return "Simple plan for " + task.Title, nil
}

// SimplePlannerCore implements contracts.TaskPlanner
type SimplePlannerCore struct {
	logger telemetry.Logger
}

// CreatePlan creates a simple plan
func (p *SimplePlannerCore) CreatePlan(ctx context.Context, task interface{}) (string, error) {
	if coreTask, ok := task.(*CoreTask); ok {
		return "Simple plan for " + coreTask.Name, nil
	}
	return "Simple plan", nil
}
