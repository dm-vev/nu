package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	agentconfig "nu/internal/agent/config"
)

// ExampleDeploymentConfigBasicUsage shows automatic source detection.
func ExampleDeploymentConfigBasicUsage() {
	ctx := context.Background()

	// Example 1: Load agent with automatic source detection (recommended)
	agentInstance, err := LoadAgentFromDeploymentConfig(ctx, "research-assistant", "production")
	if err != nil {
		log.Fatalf("Failed to load agent: %v", err)
	}

	fmt.Printf("Agent loaded from source: %s\n", agentInstance.GetConfig().ConfigSource.Type)
}

// ExampleDeploymentConfigExplicitSources shows how to force a source.
func ExampleDeploymentConfigExplicitSources() {
	ctx := context.Background()

	// Example 2: Force remote configuration only
	remoteAgent, err := LoadAgentFromRemoteDeploymentConfig(ctx, "research-assistant", "production")
	if err != nil {
		log.Printf("Remote loading failed: %v", err)
		// Handle fallback or error
		return
	}

	// Example 3: Force local configuration only
	localAgent, err := LoadAgentFromLocalDeploymentConfig(ctx, "research-assistant", "production")
	if err != nil {
		log.Printf("Local loading failed: %v", err)
		// Handle error
		return
	}

	fmt.Printf("Remote agent loaded from: %s\n", remoteAgent.GetConfig().ConfigSource.Source)
	fmt.Printf("Local agent loaded from: %s\n", localAgent.GetConfig().ConfigSource.Source)
}

// ExampleDeploymentConfigPreview shows how to preview a configuration.
func ExampleDeploymentConfigPreview() {
	ctx := context.Background()

	// Example 4: Preview configuration without creating agent
	config, err := agentconfig.PreviewDeploymentAgentConfig(ctx, "research-assistant", "production")
	if err != nil {
		log.Fatalf("Failed to preview config: %v", err)
	}

	fmt.Printf("Config loaded from: %s\n", config.ConfigSource.Type)
	fmt.Printf("Resolved variables: %v\n", config.ConfigSource.Variables)
	fmt.Printf("Agent role: %s\n", config.Role)
	fmt.Printf("Agent goal: %s\n", config.Goal)
}

// ExampleDeploymentConfigAdvancedOptions shows loading options.
func ExampleDeploymentConfigAdvancedOptions() {
	ctx := context.Background()

	// Load with custom options
	loadOptions := []agentconfig.DeploymentConfigLoadOption{
		agentconfig.WithDeploymentConfigLocalFallback("./configs/research.yaml"), // Specific fallback file
		agentconfig.WithDeploymentConfigCache(10 * time.Minute),                  // Longer cache
		agentconfig.WithDeploymentConfigEnvOverrides(),
		agentconfig.WithDeploymentConfigVerbose(),
	}

	// Agent options for customization
	agentOptions := []Option{
		WithMaxIterations(5),
		WithRequirePlanApproval(false),
	}

	agentInstance, err := LoadAgentFromDeploymentConfigWithOptions(ctx, "research-assistant", "staging", loadOptions, agentOptions...)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Configuration loaded successfully from %s\n", agentInstance.GetConfig().ConfigSource.Source)
}

// ExampleDeploymentConfigWithVariables shows custom variables.
func ExampleDeploymentConfigWithVariables() {
	ctx := context.Background()

	// Custom variables for template substitution
	variables := map[string]string{
		"topic":        "artificial intelligence",
		"search_depth": "comprehensive",
	}

	agentInstance, err := LoadAgentFromDeploymentConfigWithVariables(ctx, "research-assistant", "development", variables)
	if err != nil {
		log.Fatalf("Failed to load agent with variables: %v", err)
	}

	fmt.Printf("Agent loaded with custom variables: %v\n", variables)
	fmt.Printf("Agent backstory: %s\n", agentInstance.GetConfig().Backstory)
}

// ExampleDeploymentConfigErrorHandling shows error handling.
func ExampleDeploymentConfigErrorHandling() {
	ctx := context.Background()

	// Try loading with error handling
	agentInstance, err := LoadAgentFromDeploymentConfig(ctx, "nonexistent-agent", "production")
	if err != nil {
		// Check if it's a specific error type
		if err.Error() == "failed to load agent config from any source" {
			log.Printf("Agent not found in any configuration source")
			// Try creating a default agent or prompt user
		} else {
			log.Printf("Configuration error: %v", err)
		}
		return
	}

	fmt.Printf("Agent loaded successfully: %s\n", agentInstance.GetConfig().ConfigSource.Type)
}

// ExampleDeploymentConfigMigration shows file-to-service migration.
func ExampleDeploymentConfigMigration() {
	ctx := context.Background()

	// OLD WAY (still works):
	// configs, err := LoadAgentConfigsFromFile("./agents.yaml")
	// instance, err := NewAgentFromConfig("research-assistant", configs, nil)

	// NEW WAY (recommended):
	agentInstance, err := LoadAgentFromDeploymentConfig(ctx, "research-assistant", "production")
	if err != nil {
		log.Fatalf("Failed to load agent: %v", err)
	}

	fmt.Printf("Migrated to new configuration system: %s\n", agentInstance.GetConfig().ConfigSource.Type)
}
