package service

import (
	"time"

	"github.com/google/uuid"
	. "nu/internal/task"
	"nu/internal/telemetry"
)

// CoreAdapter is responsible for adapting between the SDK task model and the core model
type CoreAdapter struct {
	logger telemetry.Logger
}

// NewCoreAdapter creates a new adapter for the core model
func NewCoreAdapter(logger telemetry.Logger) *CoreAdapter {
	return &CoreAdapter{
		logger: logger,
	}
}

// CreateTaskRequestToCore converts a request to CoreCreateTaskRequest.
func (a *CoreAdapter) CreateTaskRequestToCore(req interface{}) CoreCreateTaskRequest {
	// Check if we have the expected type
	if coreReq, ok := req.(CoreCreateTaskRequest); ok {
		return coreReq
	}

	// Otherwise, convert from a map
	if m, ok := req.(map[string]interface{}); ok {
		var metadata map[string]interface{}
		if m["metadata"] != nil {
			metadata = m["metadata"].(map[string]interface{})
		}

		var input map[string]interface{}
		if m["input"] != nil {
			input = m["input"].(map[string]interface{})
		}

		return CoreCreateTaskRequest{
			Name:           m["name"].(string),
			Description:    m["description"].(string),
			UserID:         m["user_id"].(string),
			ConversationID: m["conversation_id"].(string),
			Input:          input,
			Metadata:       metadata,
		}
	}

	// Empty fallback
	return CoreCreateTaskRequest{}
}

// CoreTaskToTask copies a CoreTask.
func (a *CoreAdapter) CoreTaskToTask(coreTask *CoreTask) interface{} {
	if coreTask == nil {
		return nil
	}

	// Convert steps
	steps := make([]*CoreStep, len(coreTask.Steps))
	for i, step := range coreTask.Steps {
		steps[i] = &CoreStep{
			ID:          step.ID,
			Name:        step.Name,
			Description: step.Description,
			Type:        step.Type,
			Status:      step.Status,
			OrderIndex:  step.OrderIndex,
			CreatedAt:   step.CreatedAt,
			UpdatedAt:   step.UpdatedAt,
			CompletedAt: step.CompletedAt,
			FailedAt:    step.FailedAt,
			Error:       step.Error,
			Output:      step.Output,
			Context:     step.Context,
		}
	}

	// Return the core task directly as it's our unified model now
	return &CoreTask{
		ID:             coreTask.ID,
		Name:           coreTask.Name,
		Description:    coreTask.Description,
		Status:         coreTask.Status,
		UserID:         coreTask.UserID,
		ConversationID: coreTask.ConversationID,
		Plan:           coreTask.Plan,
		Steps:          steps,
		CreatedAt:      coreTask.CreatedAt,
		UpdatedAt:      coreTask.UpdatedAt,
		CompletedAt:    coreTask.CompletedAt,
		FailedAt:       coreTask.FailedAt,
		Input:          coreTask.Input,
		Output:         coreTask.Output,
		Metadata:       coreTask.Metadata,
	}
}

// CreateTaskFromDetails creates a new task with provided details
// This is a utility function for creating tasks directly
func (a *CoreAdapter) CreateTaskFromDetails(name, description, userID string, metadata map[string]interface{}) *CoreTask {
	taskID := uuid.New().String()
	now := time.Now()

	return &CoreTask{
		ID:          taskID,
		Name:        name,
		Description: description,
		Status:      CoreStatusPending,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    metadata,
		Steps:       []*CoreStep{},
	}
}

// CreateStep creates a new step for a task
func (a *CoreAdapter) CreateStep(taskID, description, stepType string, orderIndex int) *CoreStep {
	now := time.Now()
	return &CoreStep{
		ID:          uuid.New().String(),
		Name:        "Step " + description,
		Description: description,
		Type:        stepType,
		Status:      CoreStatusPending,
		OrderIndex:  orderIndex,
		CreatedAt:   now,
		UpdatedAt:   now,
		Output:      make(map[string]interface{}),
	}
}

// CreateLog creates a new log entry for a task
func (a *CoreAdapter) CreateLog(taskID, message, level string) *CoreLog {
	return &CoreLog{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Message:   message,
		Level:     level,
		CreatedAt: time.Now(),
	}
}

// UpdateTaskStatusFromStep updates a task's status based on its steps
func (a *CoreAdapter) UpdateTaskStatusFromStep(task *CoreTask) {
	// If the task is not running yet, don't change the status
	if task.Status != CoreStatusExecuting {
		return
	}

	// Count statuses
	allCompleted := true
	hasFailed := false

	for _, step := range task.Steps {
		if step.Status == CoreStatusFailed {
			hasFailed = true
		}
		if step.Status != CoreStatusCompleted {
			allCompleted = false
		}
	}

	// Update task status based on step statuses
	if hasFailed {
		task.Status = CoreStatusFailed
		now := time.Now()
		task.FailedAt = &now
	} else if allCompleted && len(task.Steps) > 0 {
		task.Status = CoreStatusCompleted
		now := time.Now()
		task.CompletedAt = &now
	}

	// Always update the updated at time
	task.UpdatedAt = time.Now()
}
