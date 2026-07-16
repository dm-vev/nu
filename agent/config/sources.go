package config

import (
	"context"
	"fmt"
	"time"
)

// LoadRemote loads an agent configuration from the remote source only.
func LoadRemote(ctx context.Context, agentName, environment string) (*AgentConfig, error) {
	config, err := LoadDeploymentAgentConfig(ctx, agentName, environment,
		WithDeploymentConfigRemoteOnly(),
		WithDeploymentConfigCache(5*time.Minute),
		WithDeploymentConfigEnvOverrides(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load remote config: %w", err)
	}
	return config, nil
}

// LoadLocal loads an agent configuration from the local source only.
func LoadLocal(ctx context.Context, agentName, environment string) (*AgentConfig, error) {
	config, err := LoadDeploymentAgentConfig(ctx, agentName, environment,
		WithDeploymentConfigLocalOnly(),
		WithDeploymentConfigEnvOverrides(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}
	return config, nil
}

// Load resolves an agent configuration with remote and local fallback support.
func Load(ctx context.Context, agentName, environment string) (*AgentConfig, error) {
	config, err := LoadDeploymentAgentConfig(ctx, agentName, environment,
		WithDeploymentConfigLocalFallback(""),
		WithDeploymentConfigCache(5*time.Minute),
		WithDeploymentConfigEnvOverrides(),
		WithDeploymentConfigVerbose(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from any source: %w", err)
	}
	return config, nil
}

// LoadWithOptions loads an agent configuration using explicit loading options.
func LoadWithOptions(ctx context.Context, agentName, environment string, options []DeploymentConfigLoadOption) (*AgentConfig, error) {
	config, err := LoadDeploymentAgentConfig(ctx, agentName, environment, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}
	return config, nil
}

// LoadWithVariables loads an agent configuration with explicit variable substitutions.
func LoadWithVariables(ctx context.Context, agentName, environment string, variables map[string]string) (*AgentConfig, error) {
	config, err := LoadDeploymentAgentConfig(ctx, agentName, environment,
		WithDeploymentConfigLocalFallback(""),
		WithDeploymentConfigCache(5*time.Minute),
		WithDeploymentConfigEnvOverrides(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	// Variables are applied by agent.WithAgentConfig when the Agent is built.
	return config, nil
}
