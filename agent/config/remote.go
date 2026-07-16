package config

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

// loadFromRemoteByID loads configuration from starops-config-service using agent_id
func loadDeploymentConfigFromRemote(ctx context.Context, agentID, environment string, opts *DeploymentConfigLoadOptions) (*AgentConfig, error) {
	// Create client
	client, err := NewDeploymentConfigClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	// Fetch from remote service using agent_id
	response, err := client.FetchAgentConfig(ctx, agentID, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote config: %w", err)
	}

	// DEBUG: Log what the config server returned
	fmt.Printf("[DEBUG] Config server response - ResolvedVariables count: %d\n", len(response.ResolvedVariables))
	for key, value := range response.ResolvedVariables {
		displayValue := value
		if len(value) > 10 {
			displayValue = value[:10] + "..."
		}
		fmt.Printf("[DEBUG] ResolvedVariable: %s = '%s'\n", key, displayValue)
	}

	// Parse the resolved YAML - it has the agent name as top-level key
	// Format: agent_name: { role: "...", goal: "...", ... }
	var wrappedConfig map[string]AgentConfig
	if err := yaml.Unmarshal([]byte(response.ResolvedYAML), &wrappedConfig); err != nil {
		return nil, fmt.Errorf("failed to parse remote YAML: %w", err)
	}

	// Extract the first (and only) agent config from the map
	if len(wrappedConfig) == 0 {
		return nil, fmt.Errorf("no agent configuration found in remote YAML")
	}

	var config AgentConfig
	var actualAgentName string
	for name, cfg := range wrappedConfig {
		actualAgentName = name
		config = cfg
		fmt.Printf("[DEBUG] loadFromRemoteByID - Loaded config for agent: %s\n", actualAgentName)
		fmt.Printf("[DEBUG] loadFromRemoteByID - Role: %s\n", cfg.Role)
		fmt.Printf("[DEBUG] loadFromRemoteByID - Goal: %s\n", cfg.Goal)
		fmt.Printf("[DEBUG] loadFromRemoteByID - Backstory: %s\n", cfg.Backstory)
		break
	}

	// Set source metadata with the actual agent name from YAML
	config.ConfigSource = &ConfigSourceMetadata{
		Type:        "remote",
		Source:      fmt.Sprintf("starops-config-service://agent_id=%s/%s", agentID, environment),
		AgentID:     agentID,
		AgentName:   actualAgentName, // Use the actual agent name from YAML
		Environment: environment,
		Variables:   response.ResolvedVariables,
	}

	return &config, nil
}
