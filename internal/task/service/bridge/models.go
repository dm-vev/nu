package bridge

import "github.com/dm-vev/nu/internal/task"

func (a *Core) coreTaskToTask(coreTask *task.CoreTask) *task.Task {
	if coreTask == nil {
		return nil
	}

	// Convert the canonical model to the legacy agent-task model.
	result := &task.Task{
		ID:             coreTask.ID,
		Description:    coreTask.Description,
		Status:         a.coreStatusToTaskStatus(coreTask.Status),
		Title:          coreTask.Name,
		ConversationID: coreTask.ConversationID,
		CreatedAt:      coreTask.CreatedAt,
		UpdatedAt:      coreTask.UpdatedAt,
		CompletedAt:    coreTask.CompletedAt,
		UserID:         coreTask.UserID,
		Metadata:       coreTask.Metadata,
	}

	// Convert steps
	if len(coreTask.Steps) > 0 {
		steps := make([]task.Step, len(coreTask.Steps))
		for i, coreStep := range coreTask.Steps {
			steps[i] = a.coreStepToTaskStep(coreStep)
		}
		result.Steps = steps
	}

	// Create a simple plan if plan string is available
	if coreTask.Plan != "" {
		result.Plan = &task.Plan{
			ID:         coreTask.ID + "_plan",
			TaskID:     coreTask.ID,
			CreatedAt:  coreTask.CreatedAt,
			IsApproved: coreTask.Status == task.CoreStatusExecuting || coreTask.Status == task.CoreStatusCompleted,
			Steps:      result.Steps,
		}
	}

	return result
}

func (a *Core) coreStepToTaskStep(coreStep *task.CoreStep) task.Step {
	var output string
	if coreStep.Output != nil {
		// Convert map to string representation
		if result, ok := coreStep.Output["result"]; ok {
			if str, ok := result.(string); ok {
				output = str
			}
		}
	}

	return task.Step{
		ID:          coreStep.ID,
		PlanID:      coreStep.ID + "_plan", // Placeholder
		Description: coreStep.Description,
		Status:      a.coreStatusToTaskStatus(coreStep.Status),
		Order:       coreStep.OrderIndex,
		StartedAt:   nil, // Not available directly
		CompletedAt: coreStep.CompletedAt,
		Error:       coreStep.Error,
		Output:      output,
	}
}
