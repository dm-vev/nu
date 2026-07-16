package bridge

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/task"
	"github.com/dm-vev/nu/telemetry"
)

// Service defines the interface for task management from the task package
// This is defined here to avoid import cycles
type Service interface {
	// CreateTask creates a new task
	CreateTask(ctx context.Context, req task.CreateTaskRequest) (*task.Task, error)
	// GetTask gets a task by ID
	GetTask(ctx context.Context, taskID string) (*task.Task, error)
	// ListTasks returns tasks filtered by the provided criteria
	ListTasks(ctx context.Context, filter task.Filter) ([]*task.Task, error)
	// ApproveTaskPlan approves or rejects a task plan
	ApproveTaskPlan(ctx context.Context, taskID string, req task.ApproveTaskPlanRequest) (*task.Task, error)
	// UpdateTask updates an existing task with new steps or modifications
	UpdateTask(ctx context.Context, taskID string, updates []task.Update) (*task.Task, error)
	// AddTaskLog adds a log entry to a task
	AddTaskLog(ctx context.Context, taskID string, message string, level string) error
}

// Core provides a bridge between the old task.Service interface and the new contracts.TaskService
// This allows migrating to the new core interfaces without breaking existing code
type Core struct {
	coreService contracts.TaskService
	logger      telemetry.Logger
}

// NewCore creates a new bridge adapter
func NewCore(coreService contracts.TaskService, logger telemetry.Logger) Service {
	return &Core{
		coreService: coreService,
		logger:      logger,
	}
}

// CreateTask creates a new task
func (a *Core) CreateTask(ctx context.Context, req task.CreateTaskRequest) (*task.Task, error) {
	// Convert the request to core format
	coreReq := task.CoreCreateTaskRequest{
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
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// GetTask gets a task by ID
func (a *Core) GetTask(ctx context.Context, taskID string) (*task.Task, error) {
	// Call the core service
	coreTaskObj, err := a.coreService.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// ListTasks returns tasks filtered by the provided criteria
func (a *Core) ListTasks(ctx context.Context, filter task.Filter) ([]*task.Task, error) {
	// Convert the filter to core format
	coreFilter := task.CoreTaskFilter{
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
	tasks := make([]*task.Task, 0, len(coreTaskObjList))
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

// ApproveTaskPlan approves or rejects a task plan
func (a *Core) ApproveTaskPlan(ctx context.Context, taskID string, req task.ApproveTaskPlanRequest) (*task.Task, error) {
	// Convert the request to core format
	coreReq := task.CoreApproveTaskPlanRequest{
		Approved: req.Approved,
		Feedback: req.Feedback,
	}

	// Call the core service
	coreTaskObj, err := a.coreService.ApproveTaskPlan(ctx, taskID, coreReq)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// UpdateTask updates an existing task with new steps or modifications
func (a *Core) UpdateTask(ctx context.Context, taskID string, updates []task.Update) (*task.Task, error) {
	// Convert the updates to core format
	coreUpdates := make([]task.CoreTaskUpdate, len(updates))
	for i, update := range updates {
		coreUpdates[i] = a.taskUpdateToCoreUpdate(update)
	}

	// Call the core service
	coreTaskObj, err := a.coreService.UpdateTask(ctx, taskID, coreUpdates)
	if err != nil {
		return nil, err
	}

	// Type assertion
	coreTask, ok := coreTaskObj.(*task.CoreTask)
	if !ok {
		return nil, fmt.Errorf("unexpected type returned from core service")
	}

	// Convert the response back to task format
	return a.coreTaskToTask(coreTask), nil
}

// AddTaskLog adds a log entry to a task
func (a *Core) AddTaskLog(ctx context.Context, taskID string, message string, level string) error {
	// Call the core service
	return a.coreService.AddTaskLog(ctx, taskID, message, level)
}
