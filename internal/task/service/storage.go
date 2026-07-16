package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	. "nu/internal/task"
)

// CreateTask creates a new task
func (s *InMemoryTaskService) CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error) {
	taskID := uuid.New().String()
	s.logger.Info(ctx, "Creating new task", map[string]interface{}{
		"task_id": taskID,
	})

	newTask := &Task{
		ID:          taskID,
		Description: req.Description,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UserID:      req.UserID,
		Logs:        []LogEntry{},
		Metadata:    req.Metadata,
	}

	// Add initial log entry
	logEntry := LogEntry{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Message:   "Task created",
		Level:     "info",
		Timestamp: time.Now(),
	}
	newTask.Logs = append(newTask.Logs, logEntry)

	// Store the task
	s.mutex.Lock()
	s.tasks[taskID] = newTask
	s.taskHistories[taskID] = []string{req.Description}
	s.mutex.Unlock()

	s.logger.Info(ctx, "Task created successfully", map[string]interface{}{
		"task_id": taskID,
	})

	// Start planning in a goroutine if planner is available
	if s.planner != nil {
		go s.planTask(context.Background(), newTask) // #nosec G118 - background context is intentional for async planning
	}

	return newTask, nil
}

// GetTask gets a task by ID
func (s *InMemoryTaskService) GetTask(ctx context.Context, taskID string) (*Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Info(ctx, "Getting task", map[string]interface{}{
		"task_id": taskID,
	})

	t, ok := s.tasks[taskID]
	if !ok {
		s.logger.Error(ctx, "Task not found", map[string]interface{}{
			"task_id": taskID,
		})
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return t, nil
}

// ListTasks returns tasks filtered by the provided criteria
func (s *InMemoryTaskService) ListTasks(ctx context.Context, filter Filter) ([]*Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tasks := make([]*Task, 0)
	for _, t := range s.tasks {
		// Apply filters
		if filter.UserID != "" && t.UserID != filter.UserID {
			continue
		}

		if len(filter.Status) > 0 {
			statusMatch := false
			for _, status := range filter.Status {
				if t.Status == status {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				continue
			}
		}

		if filter.TaskKind != "" && t.TaskKind != filter.TaskKind {
			continue
		}

		if filter.CreatedAfter != nil && !t.CreatedAt.After(*filter.CreatedAfter) {
			continue
		}

		if filter.CreatedBefore != nil && !t.CreatedAt.Before(*filter.CreatedBefore) {
			continue
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// CreateTask creates a new task
func (s *CoreMemoryTaskService) CreateTask(ctx context.Context, reqObj interface{}) (interface{}, error) {
	// Try to convert the request
	var req CoreCreateTaskRequest
	var ok bool

	if req, ok = reqObj.(CoreCreateTaskRequest); !ok {
		// Try to convert from map
		if reqMap, ok := reqObj.(map[string]interface{}); ok {
			req = CoreCreateTaskRequest{
				Name:        reqMap["name"].(string),
				Description: reqMap["description"].(string),
				UserID:      reqMap["user_id"].(string),
			}

			// Handle optional fields
			if conv, ok := reqMap["conversation_id"].(string); ok {
				req.ConversationID = conv
			}
			if meta, ok := reqMap["metadata"].(map[string]interface{}); ok {
				req.Metadata = meta
			}
			if input, ok := reqMap["input"].(map[string]interface{}); ok {
				req.Input = input
			}
		} else {
			return nil, fmt.Errorf("invalid request type")
		}
	}

	taskID := uuid.New().String()
	now := time.Now()

	task := &CoreTask{
		ID:             taskID,
		Name:           req.Name,
		Description:    req.Description,
		Status:         CoreStatusPending,
		UserID:         req.UserID,
		ConversationID: req.ConversationID,
		Steps:          []*CoreStep{},
		CreatedAt:      now,
		UpdatedAt:      now,
		Input:          req.Input,
		Metadata:       req.Metadata,
	}

	s.mutex.Lock()
	s.tasks[taskID] = task
	s.logs[taskID] = []*CoreLog{}
	s.mutex.Unlock()

	s.logger.Info(ctx, "Created new core task", map[string]interface{}{
		"task_id": taskID,
	})

	return task, nil
}

// GetTask gets a task by ID
func (s *CoreMemoryTaskService) GetTask(ctx context.Context, taskID string) (interface{}, error) {
	s.mutex.RLock()
	task, exists := s.tasks[taskID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns tasks based on the filter
func (s *CoreMemoryTaskService) ListTasks(ctx context.Context, filterObj interface{}) ([]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Try to convert the filter
	var filter CoreTaskFilter
	if f, ok := filterObj.(CoreTaskFilter); ok {
		filter = f
	} else if fMap, ok := filterObj.(map[string]interface{}); ok {
		// Try to extract filter fields from map
		if userID, ok := fMap["user_id"].(string); ok {
			filter.UserID = userID
		}
		if convID, ok := fMap["conversation_id"].(string); ok {
			filter.ConversationID = convID
		}
		if status, ok := fMap["status"].(string); ok {
			filter.Status = CoreStatus(status)
		}
	}

	var results []interface{}

	for _, task := range s.tasks {
		// Apply filters
		if filter.UserID != "" && task.UserID != filter.UserID {
			continue
		}
		if filter.ConversationID != "" && task.ConversationID != filter.ConversationID {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.FromDate != nil && task.CreatedAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && task.CreatedAt.After(*filter.ToDate) {
			continue
		}

		results = append(results, task)
	}

	// Apply limits and offset
	if filter.Limit > 0 && len(results) > filter.Limit {
		end := filter.Offset + filter.Limit
		if end > len(results) {
			end = len(results)
		}
		if filter.Offset < len(results) {
			results = results[filter.Offset:end]
		} else {
			results = []interface{}{}
		}
	}

	return results, nil
}
