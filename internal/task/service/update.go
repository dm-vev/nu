package service

import (
	"context"
	"fmt"
	"time"

	. "github.com/dm-vev/nu/internal/task"
	"github.com/google/uuid"
)

// UpdateTask updates an existing task with new steps or modifications
func (s *InMemoryTaskService) UpdateTask(ctx context.Context, taskID string, updates []Update) (*Task, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info(ctx, "Updating task", map[string]interface{}{
		"task_id": taskID,
	})

	t, ok := s.tasks[taskID]
	if !ok {
		s.logger.Error(ctx, "Task not found", map[string]interface{}{
			"task_id": taskID,
		})
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Process updates
	for _, update := range updates {
		switch update.Type {
		case "add_step":
			// Create new step
			step := Step{
				ID:          uuid.New().String(),
				Description: update.Description,
				Status:      StepStatusPending,
				Order:       len(t.Steps) + 1,
			}

			// Add to task steps
			t.Steps = append(t.Steps, step)

			// Add log entry
			logEntry := LogEntry{
				ID:        uuid.New().String(),
				TaskID:    taskID,
				Message:   fmt.Sprintf("Added step: %s", update.Description),
				Level:     "info",
				Timestamp: time.Now(),
			}
			t.Logs = append(t.Logs, logEntry)

		case "modify_step":
			// Find step
			var stepFound bool
			for i, step := range t.Steps {
				if step.ID == update.StepID {
					// Update step
					if update.Description != "" {
						t.Steps[i].Description = update.Description
					}
					if update.Status != "" {
						t.Steps[i].Status = Status(update.Status)
					}
					stepFound = true
					break
				}
			}

			if !stepFound {
				s.logger.Error(ctx, "Step not found", map[string]interface{}{
					"task_id": taskID,
					"step_id": update.StepID,
				})
				return nil, fmt.Errorf("step not found: %s", update.StepID)
			}

			// Add log entry
			logEntry := LogEntry{
				ID:        uuid.New().String(),
				TaskID:    taskID,
				StepID:    update.StepID,
				Message:   fmt.Sprintf("Modified step: %s", update.Description),
				Level:     "info",
				Timestamp: time.Now(),
			}
			t.Logs = append(t.Logs, logEntry)

		case "remove_step":
			// Find step
			var stepIndex = -1
			for i, step := range t.Steps {
				if step.ID == update.StepID {
					stepIndex = i
					break
				}
			}

			if stepIndex == -1 {
				s.logger.Error(ctx, "Step not found", map[string]interface{}{
					"task_id": taskID,
					"step_id": update.StepID,
				})
				return nil, fmt.Errorf("step not found: %s", update.StepID)
			}

			// Remove step
			t.Steps = append(t.Steps[:stepIndex], t.Steps[stepIndex+1:]...)

			// Add log entry
			logEntry := LogEntry{
				ID:        uuid.New().String(),
				TaskID:    taskID,
				StepID:    update.StepID,
				Message:   "Removed step",
				Level:     "info",
				Timestamp: time.Now(),
			}
			t.Logs = append(t.Logs, logEntry)

		case "update_status":
			// Update task status
			t.Status = Status(update.Status)

			// Add log entry
			logEntry := LogEntry{
				ID:        uuid.New().String(),
				TaskID:    taskID,
				Message:   fmt.Sprintf("Updated status to: %s", update.Status),
				Level:     "info",
				Timestamp: time.Now(),
			}
			t.Logs = append(t.Logs, logEntry)

		case "add_comment":
			// Add log entry
			logEntry := LogEntry{
				ID:        uuid.New().String(),
				TaskID:    taskID,
				Message:   update.Description,
				Level:     "info",
				Timestamp: time.Now(),
			}
			t.Logs = append(t.Logs, logEntry)
		}
	}

	// Update task
	t.UpdatedAt = time.Now()

	return t, nil
}

// UpdateTask updates a task
func (s *CoreMemoryTaskService) UpdateTask(ctx context.Context, taskID string, updatesObj interface{}) (interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	task.UpdatedAt = time.Now()

	// Try to convert the updates
	var updates []CoreTaskUpdate
	var ok bool

	if updates, ok = updatesObj.([]CoreTaskUpdate); !ok {
		// Try to convert from array of maps
		if updatesArray, ok := updatesObj.([]interface{}); ok {
			for _, updateObj := range updatesArray {
				if updateMap, ok := updateObj.(map[string]interface{}); ok {
					update := CoreTaskUpdate{}
					if field, ok := updateMap["field"].(string); ok {
						update.Field = field
					}
					if value, ok := updateMap["value"]; ok {
						update.Value = value
					}
					updates = append(updates, update)
				}
			}
		} else {
			return nil, fmt.Errorf("invalid updates type")
		}
	}

	for _, update := range updates {
		switch update.Field {
		case "status":
			if statusStr, ok := update.Value.(string); ok {
				task.Status = CoreStatus(statusStr)
				switch task.Status {
				case CoreStatusExecuting:
					now := time.Now()
					task.UpdatedAt = now
				case CoreStatusCompleted, CoreStatusFailed:
					now := time.Now()
					task.CompletedAt = &now
					task.UpdatedAt = now
				}
			}
		case "add_step":
			if stepData, ok := update.Value.(map[string]interface{}); ok {
				step := &CoreStep{
					ID:         uuid.New().String(),
					OrderIndex: len(task.Steps),
					Status:     CoreStatusPending,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
					Output:     make(map[string]interface{}),
				}

				// Get fields from the step data
				if name, ok := stepData["name"].(string); ok {
					step.Name = name
				}
				if desc, ok := stepData["description"].(string); ok {
					step.Description = desc
				}
				if stepType, ok := stepData["type"].(string); ok {
					step.Type = stepType
				}

				task.Steps = append(task.Steps, step)
			}
		case "update_step":
			if stepData, ok := update.Value.(map[string]interface{}); ok {
				stepID, ok := stepData["id"].(string)
				if !ok {
					continue
				}

				for i, step := range task.Steps {
					if step.ID == stepID {
						if status, ok := stepData["status"].(string); ok {
							task.Steps[i].Status = CoreStatus(status)

							switch task.Steps[i].Status {
							case CoreStatusExecuting:
								now := time.Now()
								task.Steps[i].UpdatedAt = now
							case CoreStatusCompleted:
								now := time.Now()
								task.Steps[i].CompletedAt = &now
								task.Steps[i].UpdatedAt = now
							case CoreStatusFailed:
								now := time.Now()
								task.Steps[i].FailedAt = &now
								task.Steps[i].UpdatedAt = now

								if errStr, ok := stepData["error"].(string); ok {
									task.Steps[i].Error = errStr
								}
							}
						}

						if output, ok := stepData["output"].(map[string]interface{}); ok {
							task.Steps[i].Output = output
						}

						break
					}
				}
			}
		case "plan":
			if planStr, ok := update.Value.(string); ok {
				task.Plan = planStr
			}
		}
	}

	s.logger.Info(ctx, "Task updated", map[string]interface{}{
		"task_id":         taskID,
		"update_count":    len(updates),
		"resulting_state": task.Status,
	})

	return task, nil
}
