package config

import "time"

// DeploymentConfigSource indicates where deployment configuration came from.
type DeploymentConfigSource string

const (
	DeploymentConfigSourceRemote DeploymentConfigSource = "remote"
	DeploymentConfigSourceLocal  DeploymentConfigSource = "local"
	DeploymentConfigSourceCache  DeploymentConfigSource = "cache"
	DeploymentConfigSourceMerged DeploymentConfigSource = "merged"
)

// DeploymentConfigMergeStrategy determines how remote and local configs are merged.
type DeploymentConfigMergeStrategy string

const (
	// MergeStrategyNone - No merging, use only one source (default behavior)
	DeploymentConfigMergeNone DeploymentConfigMergeStrategy = "none"

	// MergeStrategyRemotePriority - Remote config is primary, local fills gaps (recommended)
	// Use case: Config server has authority, local provides defaults
	DeploymentConfigMergeRemotePriority DeploymentConfigMergeStrategy = "remote_priority"

	// MergeStrategyLocalPriority - Local config is primary, remote fills gaps
	// Use case: Local development with remote fallbacks
	DeploymentConfigMergeLocalPriority DeploymentConfigMergeStrategy = "local_priority"
)

// DeploymentConfigLoadOptions configures deployment configuration loading.
type DeploymentConfigLoadOptions struct {
	// Source preferences
	PreferRemote  bool   // Try remote first
	AllowFallback bool   // Fall back to local if remote fails
	LocalPath     string // Specific local file path

	// Merging
	MergeStrategy DeploymentConfigMergeStrategy // How to merge remote and local configs

	// Caching
	EnableCache  bool
	CacheTimeout time.Duration

	// Behavior
	EnableEnvOverrides bool
	Verbose            bool // Log loading steps
}

// DefaultDeploymentConfigLoadOptions returns sensible defaults.
func DefaultDeploymentConfigLoadOptions() *DeploymentConfigLoadOptions {
	return &DeploymentConfigLoadOptions{
		PreferRemote:       true, // Try remote first
		AllowFallback:      true, // Fall back to local if remote fails
		MergeStrategy:      DeploymentConfigMergeNone,
		EnableCache:        true,
		CacheTimeout:       5 * time.Minute,
		EnableEnvOverrides: true,
		Verbose:            false,
	}
}

// DeploymentConfigLoadOption configures deployment configuration loading.
type DeploymentConfigLoadOption func(*DeploymentConfigLoadOptions)

// WithLocalFallback enables fallback to local file
func WithDeploymentConfigLocalFallback(path string) DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.AllowFallback = true
		opts.LocalPath = path
	}
}

// WithCache enables caching with specified timeout
func WithDeploymentConfigCache(timeout time.Duration) DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.EnableCache = true
		opts.CacheTimeout = timeout
	}
}

// WithoutCache disables caching
func WithoutDeploymentConfigCache() DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.EnableCache = false
	}
}

// WithEnvOverrides enables environment variable overrides
func WithDeploymentConfigEnvOverrides() DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.EnableEnvOverrides = true
	}
}

// WithVerbose enables verbose logging
func WithDeploymentConfigVerbose() DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.Verbose = true
	}
}

// WithRemoteOnly forces remote configuration only
func WithDeploymentConfigRemoteOnly() DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.PreferRemote = true
		opts.AllowFallback = false
	}
}

// WithLocalOnly forces local configuration only
func WithDeploymentConfigLocalOnly() DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.PreferRemote = false
		opts.AllowFallback = false
	}
}

// WithMergeStrategy sets the merge strategy for combining remote and local configs
func WithDeploymentConfigMergeStrategy(strategy DeploymentConfigMergeStrategy) DeploymentConfigLoadOption {
	return func(opts *DeploymentConfigLoadOptions) {
		opts.MergeStrategy = strategy
		// When merging, we need to load both sources
		if strategy != DeploymentConfigMergeNone {
			opts.AllowFallback = true // Ensure we try to load both
		}
	}
}

// WithRemotePriorityMerge enables merging with remote config taking priority
// Local config provides defaults for fields not set in remote
func WithDeploymentConfigRemotePriorityMerge() DeploymentConfigLoadOption {
	return WithDeploymentConfigMergeStrategy(DeploymentConfigMergeRemotePriority)
}

// WithLocalPriorityMerge enables merging with local config taking priority
// Remote config provides defaults for fields not set in local
func WithDeploymentConfigLocalPriorityMerge() DeploymentConfigLoadOption {
	return WithDeploymentConfigMergeStrategy(DeploymentConfigMergeLocalPriority)
}
