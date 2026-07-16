package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	. "github.com/dm-vev/nu/internal/task"
	"github.com/dm-vev/nu/internal/task/service/bridge"
	"github.com/dm-vev/nu/telemetry"
)

// AgentTaskService provides a complete implementation for the AgentTaskServiceInterface
// Agents can use this directly without additional wrapper.
type AgentTaskService struct {
	service        AgentAdapterService
	currentTask    *Task
	currentTaskMux sync.RWMutex
	logger         telemetry.Logger
}

// NewAgentTaskService creates a new TaskService for agents
func NewAgentTaskService(logger telemetry.Logger) (*AgentTaskService, error) {
	// Since we've removed the InMemoryTaskService, we'll use our CoreBridgeAdapter instead
	// Create a core planner
	corePlanner := &SimplePlannerCore{}

	// Create a core memory service
	coreService := &SimpleCoreMemoryTaskService{
		tasks:   make(map[string]*CoreTask),
		logs:    make(map[string][]*CoreLog),
		mutex:   sync.RWMutex{},
		logger:  logger,
		planner: corePlanner,
	}

	// Create a bridge adapter
	bridgeAdapter := bridge.NewSimple(coreService, logger)

	return &AgentTaskService{
		service:     bridgeAdapter,
		currentTask: nil,
		logger:      logger,
	}, nil
}

// NewAgentTaskServiceWithAdapter creates a new TaskService for agents using a custom service adapter
func NewAgentTaskServiceWithAdapter(logger telemetry.Logger, service AgentAdapterService) *AgentTaskService {
	return &AgentTaskService{
		service:     service,
		currentTask: nil,
		logger:      logger,
	}
}

// CreateTask creates a new task with the given parameters
func (s *AgentTaskService) CreateTask(ctx context.Context, title, desc string, userID string, metadata map[string]interface{}) (*Task, error) {
	req := CreateTaskRequest{
		Title:       title,
		Description: desc,
		UserID:      userID,
		Metadata:    metadata,
	}

	task, err := s.service.CreateTask(ctx, req)
	if err != nil {
		return nil, err
	}

	s.currentTaskMux.Lock()
	s.currentTask = task
	s.currentTaskMux.Unlock()

	return task, nil
}

// CurrentTask returns the current task being worked on, if any
func (s *AgentTaskService) CurrentTask() *Task {
	s.currentTaskMux.RLock()
	defer s.currentTaskMux.RUnlock()
	return s.currentTask
}

// GetTask gets a task by ID
func (s *AgentTaskService) GetTask(ctx context.Context, taskID string) (*Task, error) {
	return s.service.GetTask(ctx, taskID)
}

// ListTasks returns all tasks for the current user
func (s *AgentTaskService) ListTasks(ctx context.Context, userID string) ([]*Task, error) {
	filter := Filter{
		UserID: userID,
	}
	return s.service.ListTasks(ctx, filter)
}

// StartTask sets the current task to a specific task ID
func (s *AgentTaskService) StartTask(ctx context.Context, taskID string) (*Task, error) {
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	s.currentTaskMux.Lock()
	s.currentTask = task
	s.currentTaskMux.Unlock()

	return task, nil
}

// ResetCurrentTask clears the current task
func (s *AgentTaskService) ResetCurrentTask() {
	s.currentTaskMux.Lock()
	s.currentTask = nil
	s.currentTaskMux.Unlock()
}

// AddTaskStep adds a step to the current task
func (s *AgentTaskService) AddTaskStep(ctx context.Context, description string) error {
	s.currentTaskMux.RLock()
	currentTask := s.currentTask
	s.currentTaskMux.RUnlock()

	if currentTask == nil {
		return errors.New("no current task")
	}

	updates := []Update{
		{
			Type:        "add_step",
			Description: description,
		},
	}

	task, err := s.service.UpdateTask(ctx, currentTask.ID, updates)
	if err != nil {
		return err
	}

	s.currentTaskMux.Lock()
	s.currentTask = task
	s.currentTaskMux.Unlock()

	return nil
}

// UpdateTaskStep updates a step in the current task
func (s *AgentTaskService) UpdateTaskStep(ctx context.Context, stepID string, status Status, output string, err error) error {
	s.currentTaskMux.RLock()
	currentTask := s.currentTask
	s.currentTaskMux.RUnlock()

	if currentTask == nil {
		return errors.New("no current task")
	}

	update := Update{
		Type:   "modify_step",
		StepID: stepID,
		Status: string(status),
	}

	// For error and output information, we need to use separate updates
	// or modify the step directly through the API
	updates := []Update{update}

	// Add error information if present
	if err != nil {
		errorUpdate := Update{
			Type:        "add_log",
			StepID:      stepID,
			Description: "Error: " + err.Error(),
		}
		updates = append(updates, errorUpdate)
	}

	// Add output information if present
	if output != "" {
		// Use a log entry for the output
		outputUpdate := Update{
			Type:        "add_log",
			StepID:      stepID,
			Description: "Output: " + output,
		}
		updates = append(updates, outputUpdate)
	}

	task, updateErr := s.service.UpdateTask(ctx, currentTask.ID, updates)
	if updateErr != nil {
		return updateErr
	}

	s.currentTaskMux.Lock()
	s.currentTask = task
	s.currentTaskMux.Unlock()

	return nil
}

// UpdateTaskStatus updates the status of the current task
func (s *AgentTaskService) UpdateTaskStatus(ctx context.Context, status Status) error {
	s.currentTaskMux.RLock()
	currentTask := s.currentTask
	s.currentTaskMux.RUnlock()

	if currentTask == nil {
		return errors.New("no current task")
	}

	update := Update{
		Type:   "update_status",
		Status: string(status),
	}

	task, err := s.service.UpdateTask(ctx, currentTask.ID, []Update{update})
	if err != nil {
		return err
	}

	s.currentTaskMux.Lock()
	s.currentTask = task
	s.currentTaskMux.Unlock()

	return nil
}

// LogTaskInfo adds an info log entry to the current task
func (s *AgentTaskService) LogTaskInfo(ctx context.Context, message string) error {
	s.currentTaskMux.RLock()
	currentTask := s.currentTask
	s.currentTaskMux.RUnlock()

	if currentTask == nil {
		return errors.New("no current task")
	}

	return s.service.AddTaskLog(ctx, currentTask.ID, message, "info")
}

// LogTaskError adds an error log entry to the current task
func (s *AgentTaskService) LogTaskError(ctx context.Context, message string) error {
	s.currentTaskMux.RLock()
	currentTask := s.currentTask
	s.currentTaskMux.RUnlock()

	if currentTask == nil {
		return errors.New("no current task")
	}

	return s.service.AddTaskLog(ctx, currentTask.ID, message, "error")
}

// LogTaskDebug adds a debug log entry to the current task
func (s *AgentTaskService) LogTaskDebug(ctx context.Context, message string) error {
	s.currentTaskMux.RLock()
	currentTask := s.currentTask
	s.currentTaskMux.RUnlock()

	if currentTask == nil {
		return errors.New("no current task")
	}

	return s.service.AddTaskLog(ctx, currentTask.ID, message, "debug")
}

// FormatTaskProgress formats a string with the task progress
func (s *AgentTaskService) FormatTaskProgress() string {
	s.currentTaskMux.RLock()
	defer s.currentTaskMux.RUnlock()

	if s.currentTask == nil {
		return "No active task"
	}

	task := s.currentTask
	result := fmt.Sprintf("Task: %s (Status: %s)\n", task.Title, task.Status)

	if task.Plan != nil && len(task.Steps) > 0 {
		result += "Progress:\n"
		for i, step := range task.Steps {
			statusEmoji := "⏱️"
			switch step.Status {
			case StatusCompleted:
				statusEmoji = "✅"
			case StatusFailed:
				statusEmoji = "❌"
			case StatusExecuting:
				statusEmoji = "⚙️"
			}
			result += fmt.Sprintf("  %d. %s %s\n", i+1, statusEmoji, step.Description)
		}
	}

	return result
}
