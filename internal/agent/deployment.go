package agent

import (
	"context"
	"fmt"
	"time"

	agentconfig "nu/internal/agent/config"
)

// LoadAgentFromRemoteDeploymentConfig creates an agent from remote deployment config.
func LoadAgentFromRemoteDeploymentConfig(ctx context.Context, agentName, environment string, options ...Option) (*Agent, error) {
	config, err := agentconfig.LoadDeploymentAgentConfig(ctx, agentName, environment,
		agentconfig.WithDeploymentConfigRemoteOnly(),
		agentconfig.WithDeploymentConfigCache(5*time.Minute),
		agentconfig.WithDeploymentConfigEnvOverrides(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load remote config: %w", err)
	}

	return NewAgentFromConfigObject(ctx, config, nil, options...)
}

// LoadAgentFromLocalDeploymentConfig creates an agent from local deployment config.
func LoadAgentFromLocalDeploymentConfig(ctx context.Context, agentName, environment string, options ...Option) (*Agent, error) {
	config, err := agentconfig.LoadDeploymentAgentConfig(ctx, agentName, environment,
		agentconfig.WithDeploymentConfigLocalOnly(),
		agentconfig.WithDeploymentConfigEnvOverrides(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	return NewAgentFromConfigObject(ctx, config, nil, options...)
}

// LoadAgentFromDeploymentConfig tries remote config first, then local config.
func LoadAgentFromDeploymentConfig(ctx context.Context, agentName, environment string, options ...Option) (*Agent, error) {
	config, err := agentconfig.LoadDeploymentAgentConfig(ctx, agentName, environment,
		agentconfig.WithDeploymentConfigLocalFallback(""), // Auto-detect local file
		agentconfig.WithDeploymentConfigCache(5*time.Minute),
		agentconfig.WithDeploymentConfigEnvOverrides(),
		agentconfig.WithDeploymentConfigVerbose(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from any source: %w", err)
	}

	return NewAgentFromConfigObject(ctx, config, nil, options...)
}

// LoadAgentFromDeploymentConfigWithOptions provides full control over loading.
func LoadAgentFromDeploymentConfigWithOptions(ctx context.Context, agentName, environment string, loadOptions []agentconfig.DeploymentConfigLoadOption, agentOptions ...Option) (*Agent, error) {
	config, err := agentconfig.LoadDeploymentAgentConfig(ctx, agentName, environment, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent config: %w", err)
	}

	return NewAgentFromConfigObject(ctx, config, nil, agentOptions...)
}

// LoadAgentFromDeploymentConfigWithVariables applies variable substitutions.
func LoadAgentFromDeploymentConfigWithVariables(ctx context.Context, agentName, environment string, variables map[string]string, options ...Option) (*Agent, error) {
	config, err := agentconfig.LoadDeploymentAgentConfig(ctx, agentName, environment,
		agentconfig.WithDeploymentConfigLocalFallback(""),
		agentconfig.WithDeploymentConfigCache(5*time.Minute),
		agentconfig.WithDeploymentConfigEnvOverrides(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return NewAgentFromConfigObject(ctx, config, variables, options...)
}
