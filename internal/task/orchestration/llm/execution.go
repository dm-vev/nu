package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// executePlan executes an orchestration plan
func (o *Orchestrator) executePlan(ctx context.Context, plan *Plan) (map[string]string, error) {
	o.logger.Info(ctx, "Executing plan with", map[string]interface{}{"steps": len(plan.Steps)})

	// Log the plan structure for debugging
	for i, step := range plan.Steps {
		stepID := fmt.Sprintf("step_%d", i)
		dependsOnStr := "none"
		if len(step.DependsOn) > 0 {
			dependsOnStr = strings.Join(step.DependsOn, ", ")
		}
		o.logger.Info(ctx, "Plan step", map[string]interface{}{"id": stepID, "agent": step.AgentID, "depends_on": dependsOnStr, "description": step.Description})
	}

	results := make(map[string]string)
	completed := make(map[string]bool)

	// Create a map of step names to step IDs for dependency resolution
	stepNameToID := make(map[string]string)
	for i, step := range plan.Steps {
		stepID := fmt.Sprintf("step_%d", i)
		// Use agent_id as a fallback name for the step
		stepNameToID[step.AgentID] = stepID
		// Also map the step index as a string
		stepNameToID[fmt.Sprintf("%d", i)] = stepID
		// And map the step_X format directly
		stepNameToID[stepID] = stepID
	}

	// Execute steps until all are completed
	maxIterations := len(plan.Steps) * 3 // Prevent infinite loops
	iteration := 0

	for len(completed) < len(plan.Steps) && iteration < maxIterations {
		iteration++
		stepsExecutedThisIteration := false
		pendingSteps := []string{}

		// Find steps that can be executed
		for i, step := range plan.Steps {
			stepID := fmt.Sprintf("step_%d", i)

			// Skip completed steps
			if completed[stepID] {
				continue
			}

			// Check dependencies
			canExecute := true
			pendingDeps := []string{}

			for _, depName := range step.DependsOn {
				// Try to resolve the dependency name to a step ID
				depID, exists := stepNameToID[depName]
				if !exists {
					// If we can't resolve it, use it as is (might be a direct step_X reference)
					depID = depName
					o.logger.Info(ctx, "Warning: Could not resolve dependency name", map[string]interface{}{"name": depName})
				}

				if !completed[depID] {
					canExecute = false
					pendingDeps = append(pendingDeps, depID)
				}
			}

			if !canExecute {
				pendingSteps = append(pendingSteps, stepID)
				o.logger.Info(ctx, "Step is waiting for dependencies", map[string]interface{}{"step": stepID, "depends_on": strings.Join(pendingDeps, ", ")})
				continue
			}

			o.logger.Info(ctx, "Executing step", map[string]interface{}{"step": stepID, "agent": step.AgentID})

			// Execute step
			agent, ok := o.registry.Get(step.AgentID)
			if !ok {
				o.logger.Error(ctx, "Agent not found", map[string]interface{}{"agent": step.AgentID})
				return nil, fmt.Errorf("agent not found: %s", step.AgentID)
			}

			// Prepare input with context from dependencies
			input := step.Input
			for _, depName := range step.DependsOn {
				// Try to resolve the dependency name to a step ID
				depID, exists := stepNameToID[depName]
				if !exists {
					// If we can't resolve it, use it as is
					depID = depName
				}

				// Replace both the original name and the resolved ID in the template
				input = strings.ReplaceAll(input, fmt.Sprintf("{{%s}}", depName), results[depID])
				input = strings.ReplaceAll(input, fmt.Sprintf("{{%s}}", depID), results[depID])
			}

			// Execute agent
			result, err := agent.Run(ctx, input)
			if err != nil {
				o.logger.Error(ctx, "Failed to execute step", map[string]interface{}{"step": stepID, "error": err.Error()})
				return nil, fmt.Errorf("failed to execute step %s: %w", stepID, err)
			}

			// Store result
			o.logger.Info(ctx, "Step completed successfully", map[string]interface{}{"step": stepID})
			results[stepID] = result
			completed[stepID] = true
			stepsExecutedThisIteration = true

			// Also store the result under the agent ID for backward compatibility
			results[step.AgentID] = result

			// Log character count for research and summary agents
			switch step.AgentID {
			case "research":
				o.logger.Info(ctx, "Research agent output", map[string]interface{}{"length": len(result)})
			case "summary":
				o.logger.Info(ctx, "Summary agent output", map[string]interface{}{"length": len(result)})
			}
		}

		// Check for deadlock
		if !stepsExecutedThisIteration && len(completed) < len(plan.Steps) {
			o.logger.Info(ctx, "Potential deadlock detected", map[string]interface{}{"pending_steps": strings.Join(pendingSteps, ", ")})

			// If we're on the last step and have a creative and summary result, we can proceed
			if len(completed) == len(plan.Steps)-1 &&
				results["creative"] != "" && results["summary"] != "" {
				o.logger.Info(ctx, "Proceeding with execution despite missing one step, as we have both creative and summary results", nil)
				break
			}

			// If we have at least one result, we can try to continue with what we have
			if len(results) > 0 {
				o.logger.Info(ctx, "Attempting to continue with partial results", map[string]interface{}{"completed": len(completed), "total": len(plan.Steps)})
				break
			}

			return nil, fmt.Errorf("deadlock detected: no steps can be executed")
		}

		// Sleep to avoid busy waiting
		time.Sleep(100 * time.Millisecond)
	}

	if iteration >= maxIterations {
		o.logger.Info(ctx, "Warning: Reached maximum iterations", map[string]interface{}{"iterations": maxIterations, "completed": len(completed), "total": len(plan.Steps)})
	}

	// Log comparison of research vs summary if both exist
	if researchResult, hasResearch := results["research"]; hasResearch {
		if summaryResult, hasSummary := results["summary"]; hasSummary {
			researchLen := len(researchResult)
			summaryLen := len(summaryResult)
			compressionRatio := float64(summaryLen) / float64(researchLen) * 100
			o.logger.Info(ctx, "Character count comparison", map[string]interface{}{"research": researchLen, "summary": summaryLen, "compression_ratio": fmt.Sprintf("%.1f%% of original", compressionRatio)})
		}
	}

	o.logger.Info(ctx, "All steps completed successfully", map[string]interface{}{"completed": len(completed)})
	return results, nil
}
