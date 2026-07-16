package bridge

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/task"
	"github.com/dm-vev/nu/telemetry"
)

// Simple implements AgentAdapterService
type Simple struct {
	coreService contracts.TaskService
	logger      telemetry.Logger
}

// NewSimple creates a compatibility bridge over a canonical task service.
func NewSimple(coreService contracts.TaskService, logger telemetry.Logger) *Simple {
	return &Simple{coreService: coreService, logger: logger}
}

// CreateTask creates a task using the core service
func (a *Simple) CreateTask(ctx context.Context, req task.CreateTaskRequest) (*task.Task, error) {
	coreReq := task.CoreCreateTaskRequest{
		Name:        req.Title,
		Description: req.Description,
		UserID:      req.UserID,
		Metadata:    req.Metadata,
	}

	coreTaskObj, err := a.coreService.CreateTask(ctx, coreReq)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// GetTask gets a task by ID
func (a *Simple) GetTask(ctx context.Context, taskID string) (*task.Task, error) {
	coreTaskObj, err := a.coreService.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// ListTasks returns tasks based on filter
func (a *Simple) ListTasks(ctx context.Context, filter task.Filter) ([]*task.Task, error) {
	coreFilter := task.CoreTaskFilter{
		UserID: filter.UserID,
	}

	coreTaskObjList, err := a.coreService.ListTasks(ctx, coreFilter)
	if err != nil {
		return nil, err
	}

	var tasks []*task.Task
	for _, coreTaskObj := range coreTaskObjList {
		// Type assertion
		coreTask, ok := coreTaskObj.(*task.CoreTask)
		if !ok {
			continue // Skip items that don't match the expected type
		}
		tasks = append(tasks, a.coreTaskToTask(coreTask))
	}

	return tasks, nil
}

// ApproveTaskPlan approves or rejects a plan
func (a *Simple) ApproveTaskPlan(ctx context.Context, taskID string, req task.ApproveTaskPlanRequest) (*task.Task, error) {
	coreReq := task.CoreApproveTaskPlanRequest{
		Approved: req.Approved,
		Feedback: req.Feedback,
	}

	coreTaskObj, err := a.coreService.ApproveTaskPlan(ctx, taskID, coreReq)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// UpdateTask updates a task
func (a *Simple) UpdateTask(ctx context.Context, taskID string, updates []task.Update) (*task.Task, error) {
	var coreUpdates []task.CoreTaskUpdate

	for _, update := range updates {
		switch update.Type {
		case "add_step":
			coreUpdates = append(coreUpdates, task.CoreTaskUpdate{
				Field: "add_step",
				Value: map[string]interface{}{
					"description": update.Description,
				},
			})
		case "update_status":
			coreUpdates = append(coreUpdates, task.CoreTaskUpdate{
				Field: "status",
				Value: update.Status,
			})
		}
	}

	coreTaskObj, err := a.coreService.UpdateTask(ctx, taskID, coreUpdates)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// AddTaskLog adds a log to a task
func (a *Simple) AddTaskLog(ctx context.Context, taskID string, message string, level string) error {
	return a.coreService.AddTaskLog(ctx, taskID, message, level)
}

// coreTaskToTask converts a task.CoreTask to a legacy task.Task.
func (a *Simple) coreTaskToTask(coreTask *task.CoreTask) *task.Task {
	result := &task.Task{
		ID:          coreTask.ID,
		Title:       coreTask.Name,
		Description: coreTask.Description,
		Status:      task.Status(coreTask.Status),
		UserID:      coreTask.UserID,
		CreatedAt:   coreTask.CreatedAt,
		UpdatedAt:   coreTask.UpdatedAt,
		CompletedAt: coreTask.CompletedAt,
		Metadata:    coreTask.Metadata,
	}

	// Convert steps
	for _, coreStep := range coreTask.Steps {
		step := task.Step{
			ID:          coreStep.ID,
			Description: coreStep.Description,
			Status:      task.Status(coreStep.Status),
			Order:       coreStep.OrderIndex,
			CompletedAt: coreStep.CompletedAt,
		}
		result.Steps = append(result.Steps, step)
	}

	return result
}
