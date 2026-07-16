package service

import (
	"context"
	"fmt"
	"time"

	. "github.com/dm-vev/nu/internal/task"
	"github.com/google/uuid"
)

// AddTaskLog adds a log entry to a task
func (s *InMemoryTaskService) AddTaskLog(ctx context.Context, taskID string, message string, level string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info(ctx, "Adding log entry to task", map[string]interface{}{
		"task_id": taskID,
	})

	t, ok := s.tasks[taskID]
	if !ok {
		s.logger.Error(ctx, "Task not found", map[string]interface{}{
			"task_id": taskID,
		})
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Add log entry
	logEntry := LogEntry{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
	}
	t.Logs = append(t.Logs, logEntry)

	return nil
}

// AddTaskLog adds a log entry to a task
func (s *CoreMemoryTaskService) AddTaskLog(ctx context.Context, taskID string, message string, level string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	logEntry := &CoreLog{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Message:   message,
		Level:     level,
		CreatedAt: time.Now(),
	}

	s.logs[taskID] = append(s.logs[taskID], logEntry)

	s.logger.Info(ctx, "Added log to task", map[string]interface{}{
		"task_id": taskID,
		"level":   level,
	})

	return nil
}

// GetTaskLogs returns all logs for a task
func (s *CoreMemoryTaskService) GetTaskLogs(ctx context.Context, taskID string) ([]*CoreLog, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	logs, exists := s.logs[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return logs, nil
}
