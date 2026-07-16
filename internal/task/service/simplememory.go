package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nu/internal/contracts"
	. "nu/internal/task"
	"nu/internal/telemetry"
)

// SimpleMemoryService implements contracts.TaskService
type SimpleCoreMemoryTaskService struct {
	tasks   map[string]*CoreTask
	logs    map[string][]*CoreLog
	mutex   sync.RWMutex
	logger  telemetry.Logger
	planner contracts.TaskPlanner
}

// CreateTask creates a new task
func (s *SimpleCoreMemoryTaskService) CreateTask(ctx context.Context, req interface{}) (interface{}, error) {
	coreReq, ok := req.(CoreCreateTaskRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	task := &CoreTask{
		ID:          "task-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:        coreReq.Name,
		Description: coreReq.Description,
		Status:      CoreStatusPending,
		UserID:      coreReq.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Steps:       []*CoreStep{},
	}

	s.mutex.Lock()
	s.tasks[task.ID] = task
	s.mutex.Unlock()

	return task, nil
}

// GetTask gets a task by ID
func (s *SimpleCoreMemoryTaskService) GetTask(ctx context.Context, taskID string) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns tasks filtered by criteria
func (s *SimpleCoreMemoryTaskService) ListTasks(ctx context.Context, filter interface{}) ([]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var tasks []interface{}

	// Apply canonical filters when provided.
	coreFilter, ok := filter.(CoreTaskFilter)
	if ok {
		for _, task := range s.tasks {
			if coreFilter.UserID != "" && task.UserID != coreFilter.UserID {
				continue
			}
			tasks = append(tasks, task)
		}
	} else {
		// Otherwise, just return all tasks
		for _, task := range s.tasks {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// ApproveTaskPlan approves or rejects a plan
func (s *SimpleCoreMemoryTaskService) ApproveTaskPlan(ctx context.Context, taskID string, req interface{}) (interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Try to convert the request
	coreReq, ok := req.(CoreApproveTaskPlanRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	if coreReq.Approved {
		task.Status = CoreStatusExecuting
	}

	return task, nil
}

// UpdateTask updates a task
func (s *SimpleCoreMemoryTaskService) UpdateTask(ctx context.Context, taskID string, updates interface{}) (interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Try to convert the updates
	coreUpdates, ok := updates.([]CoreTaskUpdate)
	if !ok {
		return nil, fmt.Errorf("invalid updates type")
	}

	for _, update := range coreUpdates {
		switch update.Field {
		case "add_step":
			if stepData, ok := update.Value.(map[string]interface{}); ok {
				step := &CoreStep{
					ID:          "step-" + fmt.Sprintf("%d", time.Now().UnixNano()),
					Name:        "Step",
					Description: stepData["description"].(string),
					Status:      CoreStatusPending,
					OrderIndex:  len(task.Steps),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				task.Steps = append(task.Steps, step)
			}
		case "status":
			if status, ok := update.Value.(string); ok {
				task.Status = CoreStatus(status)
			}
		}
	}

	task.UpdatedAt = time.Now()
	return task, nil
}

// AddTaskLog adds a log to a task
func (s *SimpleCoreMemoryTaskService) AddTaskLog(ctx context.Context, taskID string, message string, level string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	log := &CoreLog{
		ID:        "log-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		TaskID:    taskID,
		Message:   message,
		Level:     level,
		CreatedAt: time.Now(),
	}

	s.logs[taskID] = append(s.logs[taskID], log)
	return nil
}
