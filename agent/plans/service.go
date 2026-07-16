package plans

import (
	"github.com/dm-vev/nu/contracts"
)

// Service owns execution-plan lifecycle and plan-backed agent actions.
type Service struct {
	store     *ExecutionPlanStore
	generator *ExecutionPlanGeneratorService
	executor  *ExecutionPlanExecutor
	memory    contracts.Memory
}

// NewService creates an execution-plan service for an agent.
func NewService(llm contracts.LLM, tools []contracts.Tool, systemPrompt string, requireApproval bool, memory contracts.Memory) *Service {
	return &Service{
		store:     NewExecutionPlanStore(),
		generator: NewExecutionPlanGenerator(llm, tools, systemPrompt, requireApproval),
		executor:  NewExecutionPlanExecutor(tools),
		memory:    memory,
	}
}

// ResetTools rebuilds plan components after the available tool set changes.
func (s *Service) ResetTools(llm contracts.LLM, tools []contracts.Tool, systemPrompt string, requireApproval bool) {
	s.generator = NewExecutionPlanGenerator(llm, tools, systemPrompt, requireApproval)
	s.executor = NewExecutionPlanExecutor(tools)
}
