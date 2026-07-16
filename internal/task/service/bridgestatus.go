package service

import . "nu/internal/task"

func (a *CoreBridgeAdapter) coreStatusToTaskStatus(coreStatus CoreStatus) Status {
	switch coreStatus {
	case CoreStatusPending:
		return StatusPending
	case CoreStatusPlanning:
		return StatusPlanning
	case CoreStatusAwaitingApproval:
		return StatusApproval
	case CoreStatusExecuting:
		return StatusExecuting
	case CoreStatusCompleted:
		return StatusCompleted
	case CoreStatusFailed:
		return StatusFailed
	default:
		return StatusPending
	}
}

func (a *CoreBridgeAdapter) taskStatusToCoreStatus(taskStatus Status) CoreStatus {
	switch taskStatus {
	case StatusPending:
		return CoreStatusPending
	case StatusPlanning:
		return CoreStatusPlanning
	case StatusApproval:
		return CoreStatusAwaitingApproval
	case StatusExecuting:
		return CoreStatusExecuting
	case StatusCompleted:
		return CoreStatusCompleted
	case StatusFailed:
		return CoreStatusFailed
	default:
		return CoreStatusPending
	}
}

func (a *CoreBridgeAdapter) convertStatusFilter(taskStatuses []Status) CoreStatus {
	// CoreTaskFilter accepts one status, so use the first legacy filter value.
	if len(taskStatuses) > 0 {
		return a.taskStatusToCoreStatus(taskStatuses[0])
	}
	return ""
}
