package service

import . "nu/internal/task"

func (a *CoreBridgeAdapter) taskUpdateToCoreUpdate(update Update) CoreTaskUpdate {
	// Map task update types to core update fields
	switch update.Type {
	case "add_step":
		return CoreTaskUpdate{
			Field: "add_step",
			Value: map[string]interface{}{
				"name":        "Step",
				"description": update.Description,
				"type":        "task",
			},
		}
	case "update_status":
		return CoreTaskUpdate{
			Field: "status",
			Value: string(a.taskStatusToCoreStatus(Status(update.Status))),
		}
	case "modify_step":
		return CoreTaskUpdate{
			Field: "update_step",
			Value: map[string]interface{}{
				"id":     update.StepID,
				"status": string(a.taskStatusToCoreStatus(Status(update.Status))),
			},
		}
	default:
		// Default simple mapping
		return CoreTaskUpdate{
			Field: update.Type,
			Value: update.Description,
		}
	}
}
