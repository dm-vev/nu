package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// loadFromLocal loads configuration from local YAML file
func loadDeploymentConfigFromLocal(agentName, environment string, opts *DeploymentConfigLoadOptions) (*AgentConfig, error) {
	// Determine file path
	localPath := opts.LocalPath
	if localPath == "" {
		// Try common locations
		possiblePaths := []string{
			fmt.Sprintf("./configs/%s.yaml", agentName),
			fmt.Sprintf("./configs/%s-%s.yaml", agentName, environment),
			fmt.Sprintf("./agents/%s.yaml", agentName),
			fmt.Sprintf("./%s.yaml", agentName),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				localPath = path
				break
			}
		}

		if localPath == "" {
			return nil, fmt.Errorf("no local configuration file found for agent %s", agentName)
		}
	}

	// Use existing LoadAgentConfigsFromFile to load the file
	configs, err := LoadAgentConfigsFromFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	// Get the specific agent config
	config, exists := configs[agentName]
	if !exists {
		// Try loading as single agent config
		// #nosec G304 - localPath is controlled by application logic, not user input
		data, err := os.ReadFile(localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	}

	// Set source metadata
	absPath, _ := filepath.Abs(localPath)
	config.ConfigSource = &ConfigSourceMetadata{
		Type:   "local",
		Source: absPath,
	}

	return &config, nil
}
