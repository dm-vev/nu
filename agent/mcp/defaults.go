package mcp

import (
	"context"
	"time"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/internal/mcp/builder"
	"github.com/dm-vev/nu/telemetry"
)

func applyGlobalDefaults(globalConfig *config.MCPGlobalConfig) *config.MCPGlobalConfig {
	trueVal := true
	if globalConfig == nil {
		return &config.MCPGlobalConfig{
			Timeout:         "30s",
			RetryAttempts:   3,
			HealthCheck:     &trueVal,
			EnableResources: &trueVal,
			EnablePrompts:   &trueVal,
			EnableSampling:  &trueVal,
			EnableSchemas:   &trueVal,
			LogLevel:        "info",
		}
	}

	if globalConfig.Timeout == "" {
		globalConfig.Timeout = "30s"
	}
	if globalConfig.RetryAttempts == 0 {
		globalConfig.RetryAttempts = 3
	}
	if globalConfig.HealthCheck == nil {
		globalConfig.HealthCheck = &trueVal
	}
	if globalConfig.EnableResources == nil {
		globalConfig.EnableResources = &trueVal
	}
	if globalConfig.EnablePrompts == nil {
		globalConfig.EnablePrompts = &trueVal
	}
	if globalConfig.EnableSampling == nil {
		globalConfig.EnableSampling = &trueVal
	}
	if globalConfig.EnableSchemas == nil {
		globalConfig.EnableSchemas = &trueVal
	}
	if globalConfig.LogLevel == "" {
		globalConfig.LogLevel = "info"
	}
	return globalConfig
}

func configureBuilder(ctx context.Context, mcpBuilder *builder.Builder, globalConfig *config.MCPGlobalConfig, logger telemetry.Logger) {
	if globalConfig.Timeout != "" {
		if timeout, err := time.ParseDuration(globalConfig.Timeout); err == nil {
			mcpBuilder.WithTimeout(timeout)
			if logger != nil {
				logger.Debug(ctx, "MCP timeout configured", map[string]interface{}{"timeout": globalConfig.Timeout})
			}
		}
	}
	if globalConfig.RetryAttempts > 0 {
		mcpBuilder.WithRetry(globalConfig.RetryAttempts, time.Second)
		if logger != nil {
			logger.Debug(ctx, "MCP retry attempts configured", map[string]interface{}{"retry_attempts": globalConfig.RetryAttempts})
		}
	}

	mcpBuilder.WithHealthCheck(*globalConfig.HealthCheck)
	if logger != nil {
		logger.Debug(ctx, "MCP global configuration applied", map[string]interface{}{
			"health_check":     *globalConfig.HealthCheck,
			"enable_resources": *globalConfig.EnableResources,
			"enable_prompts":   *globalConfig.EnablePrompts,
			"enable_sampling":  *globalConfig.EnableSampling,
			"enable_schemas":   *globalConfig.EnableSchemas,
			"log_level":        globalConfig.LogLevel,
			"timeout":          globalConfig.Timeout,
			"retry_attempts":   globalConfig.RetryAttempts,
		})
	}
}
