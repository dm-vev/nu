package bridge

import "github.com/dm-vev/nu/internal/task"

func (a *Core) coreStatusToTaskStatus(coreStatus task.CoreStatus) task.Status {
	switch coreStatus {
	case task.CoreStatusPending:
		return task.StatusPending
	case task.CoreStatusPlanning:
		return task.StatusPlanning
	case task.CoreStatusAwaitingApproval:
		return task.StatusApproval
	case task.CoreStatusExecuting:
		return task.StatusExecuting
	case task.CoreStatusCompleted:
		return task.StatusCompleted
	case task.CoreStatusFailed:
		return task.StatusFailed
	default:
		return task.StatusPending
	}
}

func (a *Core) taskStatusToCoreStatus(taskStatus task.Status) task.CoreStatus {
	switch taskStatus {
	case task.StatusPending:
		return task.CoreStatusPending
	case task.StatusPlanning:
		return task.CoreStatusPlanning
	case task.StatusApproval:
		return task.CoreStatusAwaitingApproval
	case task.StatusExecuting:
		return task.CoreStatusExecuting
	case task.StatusCompleted:
		return task.CoreStatusCompleted
	case task.StatusFailed:
		return task.CoreStatusFailed
	default:
		return task.CoreStatusPending
	}
}

func (a *Core) convertStatusFilter(taskStatuses []task.Status) task.CoreStatus {
	// task.CoreTaskFilter accepts one status, so use the first legacy filter value.
	if len(taskStatuses) > 0 {
		return a.taskStatusToCoreStatus(taskStatuses[0])
	}
	return ""
}
