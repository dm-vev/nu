package service

import (
	"sync"

	"nu/internal/contracts"
	. "nu/internal/task"
	"nu/internal/telemetry"
)

// InMemoryTaskService implements the Service interface with an in-memory storage
type InMemoryTaskService struct {
	tasks         map[string]*Task
	mutex         sync.RWMutex
	logger        telemetry.Logger
	taskHistories map[string][]string
	planner       contracts.TaskPlanner
	executor      contracts.TaskExecutor
}

// NewInMemoryTaskService creates a new in-memory task service
func NewInMemoryTaskService(logger telemetry.Logger, planner contracts.TaskPlanner, executor contracts.TaskExecutor) *InMemoryTaskService {
	return &InMemoryTaskService{
		tasks:         make(map[string]*Task),
		taskHistories: make(map[string][]string),
		mutex:         sync.RWMutex{},
		logger:        logger,
		planner:       planner,
		executor:      executor,
	}
}
