package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/tools/calculator"
)

// Factory creates tools from YAML configuration
type Factory struct {
	builtinFactories    map[string]func(map[string]interface{}) (contracts.Tool, error)
	customFactories     map[string]func(map[string]interface{}) (contracts.Tool, error)
	remoteClientFactory func(string) contracts.RemoteAgentClient
	agentFactory        AgentFactory
}

// SubAgent is the part of an agent required by the agent tool wrapper.
type SubAgent interface {
	Run(context.Context, string) (string, error)
}

// AgentFactory creates an agent tool target from its remote configuration.
type AgentFactory func(url, name, description string, timeout time.Duration, requirePlanApproval *bool) (SubAgent, error)

// NewFactory creates a new tool factory with builtin tools registered
func NewFactory(remoteFactories ...func(string) contracts.RemoteAgentClient) *Factory {
	tf := &Factory{
		builtinFactories: make(map[string]func(map[string]interface{}) (contracts.Tool, error)),
		customFactories:  make(map[string]func(map[string]interface{}) (contracts.Tool, error)),
	}
	if len(remoteFactories) > 0 {
		tf.remoteClientFactory = remoteFactories[0]
	}

	// Register builtin tools
	tf.registerBuiltinTools()

	return tf
}

// WithAgentFactory sets the runtime-specific factory used for agent tools.
func (tf *Factory) WithAgentFactory(factory AgentFactory) *Factory {
	tf.agentFactory = factory
	return tf
}

// registerBuiltinTools registers all available builtin tools
func (tf *Factory) registerBuiltinTools() {
	// Calculator tool
	tf.builtinFactories["calculator"] = func(config map[string]interface{}) (contracts.Tool, error) {
		return calculator.NewCalculator(), nil
	}

	// Add other builtin tools as they become available
	// tf.builtinFactories["web_search"] = func(config map[string]interface{}) (contracts.Tool, error) {
	//     // Implementation depends on available web search tool
	//     return nil, fmt.Errorf("web_search tool not implemented yet")
	// }
}

// CreateTool creates a tool from YAML configuration
func (tf *Factory) CreateTool(config config.ToolConfigYAML) (contracts.Tool, error) {
	return tf.CreateToolWithParentConfig(config, nil)
}

// CreateToolWithParentConfig creates a tool from YAML configuration with access to parent agent config
func (tf *Factory) CreateToolWithParentConfig(config config.ToolConfigYAML, parentConfig *config.AgentConfig) (contracts.Tool, error) {
	switch config.Type {
	case "builtin":
		return tf.createBuiltinTool(config)
	case "custom":
		return tf.createCustomTool(config)
	case "agent":
		return tf.createAgentToolWithParentConfig(config, parentConfig)
	case "mcp":
		return tf.createMCPTool(config)
	default:
		return nil, fmt.Errorf("unknown tool type: %s", config.Type)
	}
}

// createBuiltinTool creates a builtin tool
func (tf *Factory) createBuiltinTool(config config.ToolConfigYAML) (contracts.Tool, error) {
	factory, exists := tf.builtinFactories[config.Name]
	if !exists {
		return nil, fmt.Errorf("unknown builtin tool: %s", config.Name)
	}

	return factory(config.Config)
}

// createCustomTool creates a custom tool
func (tf *Factory) createCustomTool(config config.ToolConfigYAML) (contracts.Tool, error) {
	factory, exists := tf.customFactories[config.Name]
	if !exists {
		return nil, fmt.Errorf("unknown custom tool: %s. Register it first using RegisterCustomTool()", config.Name)
	}

	return factory(config.Config)
}

// createAgentToolWithParentConfig creates a tool that wraps a remote agent with parent config inheritance
func (tf *Factory) createAgentToolWithParentConfig(config config.ToolConfigYAML, parentConfig *config.AgentConfig) (contracts.Tool, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("agent tool requires URL")
	}
	if tf.agentFactory == nil {
		return nil, fmt.Errorf("agent tool requires an agent factory")
	}

	timeout := 30 * time.Second
	if config.Timeout != "" {
		if t, err := time.ParseDuration(config.Timeout); err == nil {
			timeout = t
		}
	}

	// Inherit plan approval setting from parent if available
	var requirePlanApproval *bool
	if parentConfig != nil && parentConfig.RequirePlanApproval != nil {
		requirePlanApproval = parentConfig.RequirePlanApproval
	}

	// Delegate agent construction to runtime so this tools package stays independent.
	remoteAgent, err := tf.agentFactory(config.URL, config.Name, config.Description, timeout, requirePlanApproval)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote agent tool: %w", err)
	}

	return &AgentToolWrapper{agent: remoteAgent, name: config.Name, description: config.Description}, nil
}

// createMCPTool creates an MCP tool (placeholder for future implementation)
func (tf *Factory) createMCPTool(config config.ToolConfigYAML) (contracts.Tool, error) {
	return nil, fmt.Errorf("MCP tool creation from YAML not implemented yet - use MCP section in agent config instead")
}

// RegisterCustomTool allows external registration of custom tools
func (tf *Factory) RegisterCustomTool(name string, factory func(map[string]interface{}) (contracts.Tool, error)) {
	tf.customFactories[name] = factory
}
