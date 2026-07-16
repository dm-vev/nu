package service

import (
	"context"
	"fmt"
	"time"

	. "github.com/dm-vev/nu/internal/task"
	"github.com/google/uuid"
)

// ApproveTaskPlan approves or rejects a task plan
func (s *InMemoryTaskService) ApproveTaskPlan(ctx context.Context, taskID string, req ApproveTaskPlanRequest) (*Task, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info(ctx, "Approving/rejecting task plan", map[string]interface{}{
		"task_id":  taskID,
		"approved": req.Approved,
	})

	t, ok := s.tasks[taskID]
	if !ok {
		s.logger.Error(ctx, "Task not found", map[string]interface{}{
			"task_id": taskID,
		})
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	if t.Plan == nil {
		s.logger.Error(ctx, "Task has no plan", map[string]interface{}{
			"task_id": taskID,
		})
		return nil, fmt.Errorf("task has no plan: %s", taskID)
	}

	if t.Status != StatusApproval {
		s.logger.Error(ctx, "Task is not awaiting approval", map[string]interface{}{
			"task_id": taskID,
			"status":  t.Status,
		})
		return nil, fmt.Errorf("task is not awaiting approval: %s", taskID)
	}

	t.Plan.IsApproved = req.Approved
	approvedTime := time.Now()
	t.Plan.ApprovedAt = &approvedTime
	t.UpdatedAt = time.Now()
	t.Feedback = req.Feedback

	// Add log entry
	logEntry := LogEntry{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Message:   fmt.Sprintf("Plan %s", map[bool]string{true: "approved", false: "rejected"}[req.Approved]),
		Level:     "info",
		Timestamp: time.Now(),
	}
	t.Logs = append(t.Logs, logEntry)

	// If approved, start executing
	if req.Approved {
		t.Status = StatusExecuting
		startTime := time.Now()
		t.StartedAt = &startTime

		// Schedule execution
		if s.executor != nil {
			go func() { // #nosec G118 - background context is intentional for async execution
				if err := s.executor.ExecuteTask(context.Background(), t); err != nil {
					s.logger.Error(context.Background(), "Failed to execute task", map[string]interface{}{
						"task_id": t.ID,
						"error":   err.Error(),
					})

					// Update task status to failed
					s.mutex.Lock()
					t.Status = StatusFailed
					failedTime := time.Now()
					t.CompletedAt = &failedTime
					s.mutex.Unlock()
				}
			}()
		}
	} else {
		// If rejected, replan with feedback
		t.Status = StatusPlanning
		if s.planner != nil {
			go s.replanTask(context.Background(), t, req.Feedback) // #nosec G118 - background context is intentional for async replanning
		}
	}

	return t, nil
}

// planTask handles the planning of a task
func (s *InMemoryTaskService) planTask(ctx context.Context, t *Task) {
	s.mutex.Lock()
	t.Status = StatusPlanning
	s.mutex.Unlock()

	s.logger.Info(ctx, "Starting task planning", map[string]interface{}{
		"task_id": t.ID,
	})

	// Add log entry
	if err := s.AddTaskLog(ctx, t.ID, "Starting task planning", "info"); err != nil {
		s.logger.Error(ctx, "Failed to add task log", map[string]interface{}{
			"task_id": t.ID,
			"error":   err.Error(),
		})
	}

	// Call the planner to create a plan
	_, err := s.planner.CreatePlan(ctx, t)
	if err != nil {
		s.handlePlanningFailure(t, err)
		return
	}

	// Update task status to awaiting approval
	s.mutex.Lock()
	defer s.mutex.Unlock()

	t.Status = StatusApproval
	t.UpdatedAt = time.Now()

	// Add log entry
	logEntry := LogEntry{
		ID:        uuid.New().String(),
		TaskID:    t.ID,
		Message:   "Plan created, awaiting approval",
		Level:     "info",
		Timestamp: time.Now(),
	}
	t.Logs = append(t.Logs, logEntry)
}

// handlePlanningFailure handles a failure during task planning
func (s *InMemoryTaskService) handlePlanningFailure(t *Task, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	t.Status = StatusFailed
	t.UpdatedAt = time.Now()

	// Add log entry
	logEntry := LogEntry{
		ID:        uuid.New().String(),
		TaskID:    t.ID,
		Message:   fmt.Sprintf("Planning failed: %s", err.Error()),
		Level:     "error",
		Timestamp: time.Now(),
	}
	t.Logs = append(t.Logs, logEntry)
}

// replanTask handles the replanning of a task with feedback
func (s *InMemoryTaskService) replanTask(ctx context.Context, t *Task, feedback string) {
	s.mutex.Lock()
	t.Status = StatusPlanning
	s.mutex.Unlock()

	s.logger.Info(ctx, "Replanning task with feedback", map[string]interface{}{
		"task_id":  t.ID,
		"feedback": feedback,
	})

	// Add log entry
	if err := s.AddTaskLog(ctx, t.ID, "Replanning task with feedback", "info"); err != nil {
		s.logger.Error(ctx, "Failed to add task log", map[string]interface{}{
			"task_id": t.ID,
			"error":   err.Error(),
		})
	}

	// Call the planner to create a new plan
	_, err := s.planner.CreatePlan(ctx, t)
	if err != nil {
		s.handlePlanningFailure(t, err)
		return
	}

	// Update task status to awaiting approval
	s.mutex.Lock()
	defer s.mutex.Unlock()

	t.Status = StatusApproval
	t.UpdatedAt = time.Now()

	// Add log entry
	logEntry := LogEntry{
		ID:        uuid.New().String(),
		TaskID:    t.ID,
		Message:   "Task has been replanned with feedback",
		Level:     "info",
		Timestamp: time.Now(),
	}
	t.Logs = append(t.Logs, logEntry)
}

// ApproveTaskPlan approves or rejects a task plan
func (s *CoreMemoryTaskService) ApproveTaskPlan(ctx context.Context, taskID string, reqObj interface{}) (interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Only tasks in planning or awaiting approval can be approved
	if task.Status != CoreStatusPlanning && task.Status != CoreStatusAwaitingApproval {
		return nil, fmt.Errorf("task is not in a state that can be approved: %s", task.Status)
	}

	// Try to convert the request
	var req CoreApproveTaskPlanRequest
	var ok bool

	if req, ok = reqObj.(CoreApproveTaskPlanRequest); !ok {
		// Try to convert from map
		if reqMap, ok := reqObj.(map[string]interface{}); ok {
			if approved, ok := reqMap["approved"].(bool); ok {
				req.Approved = approved
			}
			if feedback, ok := reqMap["feedback"].(string); ok {
				req.Feedback = feedback
			}
		} else {
			return nil, fmt.Errorf("invalid request type")
		}
	}

	// Update task based on approval
	if req.Approved {
		task.Status = CoreStatusExecuting
		now := time.Now()
		task.UpdatedAt = now
	} else {
		task.Status = CoreStatusPlanning // Back to planning
		task.UpdatedAt = time.Now()
	}

	s.logger.Info(ctx, "Task plan approval status updated", map[string]interface{}{
		"task_id":  taskID,
		"approved": req.Approved,
	})

	return task, nil
}
