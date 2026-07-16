package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dm-vev/nu/agent/config"
)

// ExecuteTaskFromConfig executes a task using its YAML configuration.
func (a *Agent) ExecuteTaskFromConfig(ctx context.Context, taskName string, taskConfigs config.TaskConfigs, variables map[string]string) (string, error) {
	taskConfig, exists := taskConfigs[taskName]
	if !exists {
		return "", fmt.Errorf("task configuration for %s not found", taskName)
	}

	description := taskConfig.Description
	for key, value := range variables {
		description = strings.ReplaceAll(description, fmt.Sprintf("{%s}", key), value)
	}
	result, err := a.Run(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to execute task %s: %w", taskName, err)
	}

	if taskConfig.OutputFile == "" {
		return result, nil
	}
	outputPath := taskConfig.OutputFile
	for key, value := range variables {
		outputPath = strings.ReplaceAll(outputPath, fmt.Sprintf("{%s}", key), value)
	}
	if err := os.WriteFile(outputPath, []byte(result), 0600); err != nil {
		return result, fmt.Errorf("failed to write output to file %s: %w", outputPath, err)
	}
	return result, nil
}
