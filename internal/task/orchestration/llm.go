package orchestration

import (
	"context"
	"fmt"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// LLMOrchestrator orchestrates the execution of a query using multiple agents
type OrchestratorLLM struct {
	registry *OrchestratorAgentRegistry
	planner  contracts.LLM
	logger   telemetry.Logger
}

// NewLLMOrchestrator creates a new LLM orchestrator
func NewOrchestratorLLM(registry *OrchestratorAgentRegistry, planner contracts.LLM) *OrchestratorLLM {
	return &OrchestratorLLM{
		registry: registry,
		planner:  planner,
		logger:   telemetry.NewLogger(),
	}
}

// WithLogger sets the logger for the orchestrator
func (o *OrchestratorLLM) WithLogger(logger telemetry.Logger) *OrchestratorLLM {
	o.logger = logger
	return o
}

// Execute executes a query using the orchestrator
func (o *OrchestratorLLM) Execute(ctx context.Context, query string) (string, error) {
	o.logger.Info(ctx, "Starting execution for query", map[string]interface{}{"query": query})

	// Create a plan
	plan, err := o.createPlan(ctx, query)
	if err != nil {
		o.logger.Error(ctx, "Failed to create plan", map[string]interface{}{"error": err.Error()})
		return "", fmt.Errorf("failed to create plan: %w", err)
	}

	// Execute the plan
	result, err := o.executePlan(ctx, plan)
	if err != nil {
		o.logger.Error(ctx, "Failed to execute plan", map[string]interface{}{"error": err.Error()})
		return "", fmt.Errorf("failed to execute plan: %w", err)
	}

	// Generate final response
	finalResponse, err := o.generateFinalResponse(ctx, plan, result)
	if err != nil {
		o.logger.Error(ctx, "Failed to generate final response", map[string]interface{}{"error": err.Error()})
		return "", fmt.Errorf("failed to generate final response: %w", err)
	}

	o.logger.Info(ctx, "Execution completed successfully", nil)
	return finalResponse, nil
}
