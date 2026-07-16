package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadAgentConfigsFromFile loads agent configurations from a YAML file
func LoadAgentConfigsFromFile(filePath string) (AgentConfigs, error) {
	// Validate file path
	if !IsValidFilePath(filePath) {
		return nil, fmt.Errorf("invalid file path")
	}

	// Read file safely
	data, err := os.ReadFile(filePath) // #nosec G304 - Path is validated with isValidFilePath() before use
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config file: %w", err)
	}

	var configs AgentConfigs
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent configs: %w", err)
	}

	return configs, nil
}

// isValidFilePath checks if a file path is valid and safe
func IsValidFilePath(filePath string) bool {
	// Check for empty path
	if filePath == "" {
		return false
	}

	// Clean and normalize the path
	cleanPath := filepath.Clean(filePath)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return false
	}

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false
	}

	// On Unix systems, check if the path is absolute and doesn't start with /proc, /sys, etc.
	// which could lead to sensitive information disclosure
	if strings.HasPrefix(absPath, "/proc") ||
		strings.HasPrefix(absPath, "/sys") ||
		strings.HasPrefix(absPath, "/dev") {
		return false
	}

	// Ensure the file exists
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return false
	}

	// Ensure it's a regular file, not a directory or symlink
	return fileInfo.Mode().IsRegular()
}

// ValidateConfigPath validates that a configuration file path is safe.
func ValidateConfigPath(path string) error {
	if path == "" {
		return fmt.Errorf("config path cannot be empty")
	}
	if !IsValidFilePath(path) {
		return fmt.Errorf("invalid or unsafe config path: %s", path)
	}
	return nil
}

// LoadAgentConfigsFromDir loads all agent configurations from YAML files in a directory
func LoadAgentConfigsFromDir(dirPath string) (AgentConfigs, error) {
	// Validate directory path
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config directory: %w", err)
	}

	configs := make(AgentConfigs)
	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())

		// Validate the file path before loading
		if !IsValidFilePath(filePath) {
			continue // Skip invalid files but don't fail completely
		}

		fileConfigs, err := LoadAgentConfigsFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load agent configs from %s: %w", filePath, err)
		}

		// Merge configs
		for name, config := range fileConfigs {
			configs[name] = config
		}
	}

	return configs, nil
}

// LoadTaskConfigsFromFile loads task configurations from a YAML file
func LoadTaskConfigsFromFile(filePath string) (TaskConfigs, error) {
	// Validate file path
	if !IsValidFilePath(filePath) {
		return nil, fmt.Errorf("invalid file path")
	}

	// Read file safely
	data, err := os.ReadFile(filePath) // #nosec G304 - Path is validated with isValidFilePath() before use
	if err != nil {
		return nil, fmt.Errorf("failed to read task config file: %w", err)
	}

	var configs TaskConfigs
	if err := yaml.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task configs: %w", err)
	}

	return configs, nil
}

// LoadTaskConfigsFromDir loads all task configurations from YAML files in a directory
func LoadTaskConfigsFromDir(dirPath string) (TaskConfigs, error) {
	// Validate directory path
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task config directory: %w", err)
	}

	configs := make(TaskConfigs)
	for _, file := range files {
		if file.IsDir() || (!strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml")) {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())

		// Validate the file path before loading
		if !IsValidFilePath(filePath) {
			continue // Skip invalid files but don't fail completely
		}

		fileConfigs, err := LoadTaskConfigsFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load task configs from %s: %w", filePath, err)
		}

		// Merge configs
		for name, config := range fileConfigs {
			configs[name] = config
		}
	}

	return configs, nil
}

// FormatSystemPromptFromConfig formats a system prompt based on the agent configuration
func FormatSystemPromptFromConfig(config AgentConfig, variables map[string]string) string {
	role := config.Role
	goal := config.Goal
	backstory := config.Backstory

	// Replace variables in the configuration
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		role = strings.ReplaceAll(role, placeholder, value)
		goal = strings.ReplaceAll(goal, placeholder, value)
		backstory = strings.ReplaceAll(backstory, placeholder, value)
	}

	return fmt.Sprintf("# Role\n%s\n\n# Goal\n%s\n\n# Backstory\n%s", role, goal, backstory)
}

// GetAgentForTask returns the agent name for a given task
func GetAgentForTask(taskConfigs TaskConfigs, taskName string) (string, error) {
	taskConfig, exists := taskConfigs[taskName]
	if !exists {
		return "", fmt.Errorf("task %s not found in configuration", taskName)
	}
	return taskConfig.Agent, nil
}
