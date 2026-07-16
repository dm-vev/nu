package agent

import (
	"context"

	"github.com/dm-vev/nu/agent/config"
)

// LoadAgentFromRemoteDeploymentConfig creates an agent from remote deployment config.
func LoadAgentFromRemoteDeploymentConfig(ctx context.Context, agentName, environment string, options ...Option) (*Agent, error) {
	config, err := config.LoadRemote(ctx, agentName, environment)
	if err != nil {
		return nil, err
	}
	return NewAgentFromConfigObject(ctx, config, nil, options...)
}

// LoadAgentFromLocalDeploymentConfig creates an agent from local deployment config.
func LoadAgentFromLocalDeploymentConfig(ctx context.Context, agentName, environment string, options ...Option) (*Agent, error) {
	config, err := config.LoadLocal(ctx, agentName, environment)
	if err != nil {
		return nil, err
	}
	return NewAgentFromConfigObject(ctx, config, nil, options...)
}

// LoadAgentFromDeploymentConfig tries remote config first, then local config.
func LoadAgentFromDeploymentConfig(ctx context.Context, agentName, environment string, options ...Option) (*Agent, error) {
	config, err := config.Load(ctx, agentName, environment)
	if err != nil {
		return nil, err
	}
	return NewAgentFromConfigObject(ctx, config, nil, options...)
}

// LoadAgentFromDeploymentConfigWithOptions provides full control over loading.
func LoadAgentFromDeploymentConfigWithOptions(ctx context.Context, agentName, environment string, loadOptions []config.DeploymentConfigLoadOption, agentOptions ...Option) (*Agent, error) {
	config, err := config.LoadWithOptions(ctx, agentName, environment, loadOptions)
	if err != nil {
		return nil, err
	}
	return NewAgentFromConfigObject(ctx, config, nil, agentOptions...)
}

// LoadAgentFromDeploymentConfigWithVariables applies variable substitutions.
func LoadAgentFromDeploymentConfigWithVariables(ctx context.Context, agentName, environment string, variables map[string]string, options ...Option) (*Agent, error) {
	config, err := config.LoadWithVariables(ctx, agentName, environment, variables)
	if err != nil {
		return nil, err
	}
	return NewAgentFromConfigObject(ctx, config, variables, options...)
}
