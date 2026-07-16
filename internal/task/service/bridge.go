package service

import (
	"context"
	"fmt"

	"nu/internal/contracts"
	. "nu/internal/task"
	"nu/internal/telemetry"
)

// Service defines the interface for task management from the task package
// This is defined here to avoid import cycles
type Service interface {
	// CreateTask creates a new task
	CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error)
	// GetTask gets a task by ID
	GetTask(ctx context.Context, taskID string) (*Task, error)
	// ListTasks returns tasks filtered by the provided criteria
	ListTasks(ctx context.Context, filter Filter) ([]*Task, error)
	// ApproveTaskPlan approves or rejects a task plan
	ApproveTaskPlan(ctx context.Context, taskID string, req ApproveTaskPlanRequest) (*Task, error)
	// UpdateTask updates an existing task with new steps or modifications
	UpdateTask(ctx context.Context, taskID string, updates []Update) (*Task, error)
	// AddTaskLog adds a log entry to a task
	AddTaskLog(ctx context.Context, taskID string, message string, level string) error
}

// CoreBridgeAdapter provides a bridge between the old task.Service interface and the new contracts.TaskService
// This allows migrating to the new core interfaces without breaking existing code
type CoreBridgeAdapter struct {
	coreService contracts.TaskService
	logger      telemetry.Logger
}

// NewCoreBridgeAdapter creates a new bridge adapter
func NewCoreBridgeAdapter(coreService contracts.TaskService, logger telemetry.Logger) Service {
	return &CoreBridgeAdapter{
		coreService: coreService,
		logger:      logger,
	}
}

// CreateTask creates a new task
func (a *CoreBridgeAdapter) CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error) {
	// Convert the request to core format
	coreReq := CoreCreateTaskRequest{
		Name:           req.Title,
		Description:    req.Description,
		UserID:         req.UserID,
		ConversationID: "",
		Input:          nil,
		Metadata:       req.Metadata,
	}

	// Call the core service
	coreTaskObj, err := a.coreService.CreateTask(ctx, coreReq)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// GetTask gets a task by ID
func (a *CoreBridgeAdapter) GetTask(ctx context.Context, taskID string) (*Task, error) {
	// Call the core service
	coreTaskObj, err := a.coreService.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// ListTasks returns tasks filtered by the provided criteria
func (a *CoreBridgeAdapter) ListTasks(ctx context.Context, filter Filter) ([]*Task, error) {
	// Convert the filter to core format
	coreFilter := CoreTaskFilter{
		UserID:         filter.UserID,
		ConversationID: filter.ConversationID,
		Status:         a.convertStatusFilter(filter.Status),
		FromDate:       filter.CreatedAfter,
		ToDate:         filter.CreatedBefore,
	}

	// Call the core service
	coreTaskObjList, err := a.coreService.ListTasks(ctx, coreFilter)
	if err != nil {
		return nil, err
	}

	// Convert the response back to task format
	tasks := make([]*Task, 0, len(coreTaskObjList))
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

// ApproveTaskPlan approves or rejects a task plan
func (a *CoreBridgeAdapter) ApproveTaskPlan(ctx context.Context, taskID string, req ApproveTaskPlanRequest) (*Task, error) {
	// Convert the request to core format
	coreReq := CoreApproveTaskPlanRequest{
		Approved: req.Approved,
		Feedback: req.Feedback,
	}

	// Call the core service
	coreTaskObj, err := a.coreService.ApproveTaskPlan(ctx, taskID, coreReq)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// UpdateTask updates an existing task with new steps or modifications
func (a *CoreBridgeAdapter) UpdateTask(ctx context.Context, taskID string, updates []Update) (*Task, error) {
	// Convert the updates to core format
	coreUpdates := make([]CoreTaskUpdate, len(updates))
	for i, update := range updates {
		coreUpdates[i] = a.taskUpdateToCoreUpdate(update)
	}

	// Call the core service
	coreTaskObj, err := a.coreService.UpdateTask(ctx, taskID, coreUpdates)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// AddTaskLog adds a log entry to a task
func (a *CoreBridgeAdapter) AddTaskLog(ctx context.Context, taskID string, message string, level string) error {
	// Call the core service
	return a.coreService.AddTaskLog(ctx, taskID, message, level)
}
