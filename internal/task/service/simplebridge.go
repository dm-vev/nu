package service

import (
	"context"
	"fmt"

	"nu/internal/contracts"
	. "nu/internal/task"
	"nu/internal/telemetry"
)

// SimpleBridgeAdapter implements AgentAdapterService
type SimpleBridgeAdapter struct {
	coreService contracts.TaskService
	logger      telemetry.Logger
}

// CreateTask creates a task using the core service
func (a *SimpleBridgeAdapter) CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error) {
	coreReq := CoreCreateTaskRequest{
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
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// GetTask gets a task by ID
func (a *SimpleBridgeAdapter) GetTask(ctx context.Context, taskID string) (*Task, error) {
	coreTaskObj, err := a.coreService.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// ListTasks returns tasks based on filter
func (a *SimpleBridgeAdapter) ListTasks(ctx context.Context, filter Filter) ([]*Task, error) {
	coreFilter := CoreTaskFilter{
		UserID: filter.UserID,
	}

	coreTaskObjList, err := a.coreService.ListTasks(ctx, coreFilter)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for _, coreTaskObj := range coreTaskObjList {
		// Type assertion
		coreTask, ok := coreTaskObj.(*CoreTask)
		if !ok {
			continue // Skip items that don't match the expected type
		}
		tasks = append(tasks, a.coreTaskToTask(coreTask))
	}

	return tasks, nil
}

// ApproveTaskPlan approves or rejects a plan
func (a *SimpleBridgeAdapter) ApproveTaskPlan(ctx context.Context, taskID string, req ApproveTaskPlanRequest) (*Task, error) {
	coreReq := CoreApproveTaskPlanRequest{
		Approved: req.Approved,
		Feedback: req.Feedback,
	}

	coreTaskObj, err := a.coreService.ApproveTaskPlan(ctx, taskID, coreReq)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// UpdateTask updates a task
func (a *SimpleBridgeAdapter) UpdateTask(ctx context.Context, taskID string, updates []Update) (*Task, error) {
	var coreUpdates []CoreTaskUpdate

	for _, update := range updates {
		switch update.Type {
		case "add_step":
			coreUpdates = append(coreUpdates, CoreTaskUpdate{
				Field: "add_step",
				Value: map[string]interface{}{
					"description": update.Description,
				},
			})
		case "update_status":
			coreUpdates = append(coreUpdates, CoreTaskUpdate{
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
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	return a.coreTaskToTask(coreTask), nil
}

// AddTaskLog adds a log to a task
func (a *SimpleBridgeAdapter) AddTaskLog(ctx context.Context, taskID string, message string, level string) error {
	return a.coreService.AddTaskLog(ctx, taskID, message, level)
}

// coreTaskToTask converts a CoreTask to a legacy Task.
func (a *SimpleBridgeAdapter) coreTaskToTask(coreTask *CoreTask) *Task {
	task := &Task{
		ID:          coreTask.ID,
		Title:       coreTask.Name,
		Description: coreTask.Description,
		Status:      Status(coreTask.Status),
		UserID:      coreTask.UserID,
		CreatedAt:   coreTask.CreatedAt,
		UpdatedAt:   coreTask.UpdatedAt,
		CompletedAt: coreTask.CompletedAt,
		Metadata:    coreTask.Metadata,
	}

	// Convert steps
	for _, coreStep := range coreTask.Steps {
		step := Step{
			ID:          coreStep.ID,
			Description: coreStep.Description,
			Status:      Status(coreStep.Status),
			Order:       coreStep.OrderIndex,
			CompletedAt: coreStep.CompletedAt,
		}
		task.Steps = append(task.Steps, step)
	}

	return task
}
