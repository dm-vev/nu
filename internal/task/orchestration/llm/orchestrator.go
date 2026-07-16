package llm

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/task/orchestration"
	"github.com/dm-vev/nu/telemetry"
)

// Orchestrator executes a query across multiple registered agents.
type Orchestrator struct {
	registry *orchestration.AgentRegistry
	planner  contracts.LLM
	logger   telemetry.Logger
}

// NewOrchestrator creates an LLM orchestrator.
func NewOrchestrator(registry *orchestration.AgentRegistry, planner contracts.LLM) *Orchestrator {
	return &Orchestrator{
		registry: registry,
		planner:  planner,
		logger:   telemetry.NewLogger(),
	}
}

// WithLogger sets the logger for the orchestrator
func (o *Orchestrator) WithLogger(logger telemetry.Logger) *Orchestrator {
	o.logger = logger
	return o
}

// Execute executes a query using the orchestrator
func (o *Orchestrator) Execute(ctx context.Context, query string) (string, error) {
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
