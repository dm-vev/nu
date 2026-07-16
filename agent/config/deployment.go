package config

import (
	"context"
	"fmt"
	"os"
	"time"
)

// LoadDeploymentAgentConfig loads an AgentConfig from remote and/or local sources.
// It uses AGENT_DEPLOYMENT_ID to load configuration from remote, then falls back to local if configured
func LoadDeploymentAgentConfig(ctx context.Context, agentName, environment string, options ...DeploymentConfigLoadOption) (*AgentConfig, error) {
	// Apply options
	opts := DefaultDeploymentConfigLoadOptions()
	for _, option := range options {
		option(opts)
	}

	agentID := os.Getenv("AGENT_DEPLOYMENT_ID")
	if agentID == "" && (opts.PreferRemote || opts.MergeStrategy != DeploymentConfigMergeNone) {
		return nil, fmt.Errorf("AGENT_DEPLOYMENT_ID environment variable is required")
	}

	if opts.Verbose {
		fmt.Printf("Loading agent config: agent_id=%s (env: %s)\n", agentID, environment)
	}
	cacheID := agentID
	if cacheID == "" {
		cacheID = agentName
	}

	// Try cache first if enabled
	if opts.EnableCache {
		cacheKey := fmt.Sprintf("%s:%s", cacheID, environment)
		if cached := getFromCache(cacheKey); cached != nil {
			if opts.Verbose {
				fmt.Printf("Loaded from cache: %s\n", cacheKey)
			}
			return cached, nil
		}
	}

	var config *AgentConfig
	var remoteConfig *AgentConfig
	var localConfig *AgentConfig
	var source DeploymentConfigSource
	var err error
	var remoteErr, localErr error

	// If merging is enabled, load both configs
	if opts.MergeStrategy != DeploymentConfigMergeNone {
		if opts.Verbose {
			fmt.Printf("Merge strategy enabled: %s\n", opts.MergeStrategy)
		}

		// Load remote config
		remoteConfig, remoteErr = loadDeploymentConfigFromRemote(ctx, agentID, environment, opts)
		if remoteErr != nil && opts.Verbose {
			fmt.Printf("Remote loading failed (will merge with local if available): %v\n", remoteErr)
		}

		// Load local config
		localConfig, localErr = loadDeploymentConfigFromLocal(agentName, environment, opts)
		if localErr != nil && opts.Verbose {
			fmt.Printf("Local loading failed (will merge with remote if available): %v\n", localErr)
		}

		// Perform merge based on strategy
		switch opts.MergeStrategy {
		case DeploymentConfigMergeRemotePriority:
			if remoteConfig != nil && localConfig != nil {
				// Both configs available - merge with remote priority
				config = MergeDeploymentAgentConfig(remoteConfig, localConfig, opts.MergeStrategy)
				source = DeploymentConfigSourceMerged
				if opts.Verbose {
					fmt.Printf("Merged remote (priority) + local configs\n")
				}
			} else if remoteConfig != nil {
				// Only remote available
				config = remoteConfig
				source = DeploymentConfigSourceRemote
			} else if localConfig != nil {
				// Only local available
				config = localConfig
				source = DeploymentConfigSourceLocal
			}
		case DeploymentConfigMergeLocalPriority:
			if remoteConfig != nil && localConfig != nil {
				// Both configs available - merge with local priority
				config = MergeDeploymentAgentConfig(localConfig, remoteConfig, opts.MergeStrategy)
				source = DeploymentConfigSourceMerged
				if opts.Verbose {
					fmt.Printf("Merged local (priority) + remote configs\n")
				}
			} else if localConfig != nil {
				// Only local available
				config = localConfig
				source = DeploymentConfigSourceLocal
			} else if remoteConfig != nil {
				// Only remote available
				config = remoteConfig
				source = DeploymentConfigSourceRemote
			}
		}

		// If no config loaded after merge attempt, return error
		if config == nil {
			return nil, fmt.Errorf("failed to load config for merging: remote error: %v, local error: %v", remoteErr, localErr)
		}
	} else {
		// No merging - use original behavior (either/or with fallback)
		// Try remote first if preferred
		if opts.PreferRemote {
			config, err = loadDeploymentConfigFromRemote(ctx, agentID, environment, opts)
			if err == nil {
				source = DeploymentConfigSourceRemote
			} else if opts.Verbose {
				fmt.Printf("Remote loading failed: %v\n", err)
			}
		}

		// Fall back to local if remote failed and fallback is enabled
		if config == nil && (opts.AllowFallback || !opts.PreferRemote) {
			config, err = loadDeploymentConfigFromLocal(agentName, environment, opts)
			if err == nil {
				source = DeploymentConfigSourceLocal
			} else if opts.Verbose {
				fmt.Printf("Local loading failed: %v\n", err)
			}
		}

		if config == nil {
			return nil, fmt.Errorf("failed to load agent config from any source: %w", err)
		}
	}

	// Add source metadata (preserve existing metadata if already set by loader)
	if config.ConfigSource == nil {
		config.ConfigSource = &ConfigSourceMetadata{}
	}
	// Only override if not already set by the loader
	if config.ConfigSource.Type == "" {
		config.ConfigSource.Type = string(source)
	}
	if config.ConfigSource.AgentID == "" {
		config.ConfigSource.AgentID = agentID
	}
	if config.ConfigSource.Environment == "" {
		config.ConfigSource.Environment = environment
	}
	// Keep the actual agent name from remote if available, otherwise use the parameter
	if config.ConfigSource.AgentName == "" {
		config.ConfigSource.AgentName = agentName
	}
	config.ConfigSource.LoadedAt = time.Now()

	// Apply environment overrides if enabled
	if opts.EnableEnvOverrides {
		*config = ExpandAgentConfig(*config)
	}

	// Cache the result if enabled
	if opts.EnableCache {
		cacheKey := fmt.Sprintf("%s:%s", cacheID, environment)
		cacheConfig(cacheKey, config, opts.CacheTimeout)
	}

	if opts.Verbose {
		fmt.Printf("Successfully loaded from %s\n", source)
	}

	return config, nil
}

// PreviewDeploymentAgentConfig resolves deployment config without caching it.
func PreviewDeploymentAgentConfig(ctx context.Context, agentName, environment string) (*AgentConfig, error) {
	return LoadDeploymentAgentConfig(ctx, agentName, environment,
		WithDeploymentConfigLocalFallback(""),
		WithoutDeploymentConfigCache(),
		WithDeploymentConfigEnvOverrides(),
	)
}
