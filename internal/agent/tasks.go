package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	agentconfig "nu/internal/agent/config"
)

// ExecuteTaskFromConfig executes a task using its YAML configuration
func (a *Agent) ExecuteTaskFromConfig(ctx context.Context, taskName string, taskConfigs agentconfig.TaskConfigs, variables map[string]string) (string, error) {
	taskConfig, exists := taskConfigs[taskName]
	if !exists {
		return "", fmt.Errorf("task configuration for %s not found", taskName)
	}

	// Replace variables in the task description
	description := taskConfig.Description
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		description = strings.ReplaceAll(description, placeholder, value)
	}

	// Run the agent with the task description
	result, err := a.Run(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to execute task %s: %w", taskName, err)
	}

	// If an output file is specified, write the result to the file
	if taskConfig.OutputFile != "" {
		outputPath := taskConfig.OutputFile
		for key, value := range variables {
			placeholder := fmt.Sprintf("{%s}", key)
			outputPath = strings.ReplaceAll(outputPath, placeholder, value)
		}

		err := os.WriteFile(outputPath, []byte(result), 0600)
		if err != nil {
			return result, fmt.Errorf("failed to write output to file %s: %w", outputPath, err)
		}
	}

	return result, nil
}
