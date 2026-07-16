package bridge

import "github.com/dm-vev/nu/internal/task"

func (a *Core) taskUpdateToCoreUpdate(update task.Update) task.CoreTaskUpdate {
	// Map task update types to core update fields
	switch update.Type {
	case "add_step":
		return task.CoreTaskUpdate{
			Field: "add_step",
			Value: map[string]interface{}{
				"name":        "task.Step",
				"description": update.Description,
				"type":        "task",
			},
		}
	case "update_status":
		return task.CoreTaskUpdate{
			Field: "status",
			Value: string(a.taskStatusToCoreStatus(task.Status(update.Status))),
		}
	case "modify_step":
		return task.CoreTaskUpdate{
			Field: "update_step",
			Value: map[string]interface{}{
				"id":     update.StepID,
				"status": string(a.taskStatusToCoreStatus(task.Status(update.Status))),
			},
		}
	default:
		// Default simple mapping
		return task.CoreTaskUpdate{
			Field: update.Type,
			Value: update.Description,
		}
	}
}
