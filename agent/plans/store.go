package plans

import (
	"sync"
)

// ExecutionPlanStore handles storage and retrieval of execution plans.
type ExecutionPlanStore struct {
	plans      map[string]*ExecutionPlan
	plansMutex sync.RWMutex
}

// NewExecutionPlanStore creates an execution plan store.
func NewExecutionPlanStore() *ExecutionPlanStore {
	return &ExecutionPlanStore{
		plans: make(map[string]*ExecutionPlan),
	}
}

// StorePlan stores an execution plan
func (s *ExecutionPlanStore) StorePlan(plan *ExecutionPlan) {
	s.plansMutex.Lock()
	defer s.plansMutex.Unlock()
	s.plans[plan.TaskID] = plan
}

// GetPlanByTaskID retrieves an execution plan by its task ID
func (s *ExecutionPlanStore) GetPlanByTaskID(taskID string) (*ExecutionPlan, bool) {
	s.plansMutex.RLock()
	defer s.plansMutex.RUnlock()
	plan, exists := s.plans[taskID]
	return plan, exists
}

// ListPlans returns a list of all plans
func (s *ExecutionPlanStore) ListPlans() []*ExecutionPlan {
	s.plansMutex.RLock()
	defer s.plansMutex.RUnlock()

	plans := make([]*ExecutionPlan, 0, len(s.plans))
	for _, plan := range s.plans {
		plans = append(plans, plan)
	}
	return plans
}

// DeletePlan deletes a plan by its task ID
func (s *ExecutionPlanStore) DeletePlan(taskID string) bool {
	s.plansMutex.Lock()
	defer s.plansMutex.Unlock()

	_, exists := s.plans[taskID]
	if exists {
		delete(s.plans, taskID)
	}
	return exists
}
