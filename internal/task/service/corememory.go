package service

import (
	"sync"

	"nu/internal/contracts"
	. "nu/internal/task"
	"nu/internal/telemetry"
)

// CoreMemoryTaskService implements contracts.TaskService with in-memory storage.
type CoreMemoryTaskService struct {
	tasks   map[string]*CoreTask
	logs    map[string][]*CoreLog
	mutex   sync.RWMutex
	logger  telemetry.Logger
	planner contracts.TaskPlanner
}

// NewCoreMemoryTaskService creates an in-memory service for canonical tasks.
func NewCoreMemoryTaskService(logger telemetry.Logger, planner contracts.TaskPlanner) contracts.TaskService {
	return &CoreMemoryTaskService{
		tasks:   make(map[string]*CoreTask),
		logs:    make(map[string][]*CoreLog),
		mutex:   sync.RWMutex{},
		logger:  logger,
		planner: planner,
	}
}

// NewService creates an in-memory task service with the default planner.
func NewService(logger telemetry.Logger) contracts.TaskService {
	return NewCoreMemoryTaskService(logger, NewCorePlanner(logger))
}
