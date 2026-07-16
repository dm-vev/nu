package mcp

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// PromptManager provides high-level operations for MCP prompts
type PromptManager struct {
	servers []contracts.MCPServer
	logger  telemetry.Logger
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(servers []contracts.MCPServer) *PromptManager {
	return &PromptManager{
		servers: servers,
		logger:  telemetry.NewLogger(),
	}
}

// ListAllPrompts lists prompts from all MCP servers
func (pm *PromptManager) ListAllPrompts(ctx context.Context) (map[string][]contracts.MCPPrompt, error) {
	result := make(map[string][]contracts.MCPPrompt)

	for i, server := range pm.servers {
		serverName := fmt.Sprintf("server-%d", i)

		prompts, err := server.ListPrompts(ctx)
		if err != nil {
			pm.logger.Warn(ctx, "Failed to list prompts from server", map[string]interface{}{
				"server": serverName,
				"error":  err.Error(),
			})
			continue
		}

		result[serverName] = prompts
		pm.logger.Debug(ctx, "Listed prompts from server", map[string]interface{}{
			"server":       serverName,
			"prompt_count": len(prompts),
		})
	}

	return result, nil
}

// FindPrompts searches for prompts by name pattern across all servers
func (pm *PromptManager) FindPrompts(ctx context.Context, pattern string) ([]PromptMatch, error) {
	var matches []PromptMatch

	for i, server := range pm.servers {
		serverName := fmt.Sprintf("server-%d", i)

		prompts, err := server.ListPrompts(ctx)
		if err != nil {
			pm.logger.Warn(ctx, "Failed to list prompts from server", map[string]interface{}{
				"server": serverName,
				"error":  err.Error(),
			})
			continue
		}

		for _, prompt := range prompts {
			if pm.matchesPattern(prompt, pattern) {
				matches = append(matches, PromptMatch{
					Server: server,
					Prompt: prompt,
				})
			}
		}
	}

	pm.logger.Debug(ctx, "Found matching prompts", map[string]interface{}{
		"pattern":     pattern,
		"match_count": len(matches),
	})

	return matches, nil
}

// GetPrompt retrieves a prompt by name from any server that has it
func (pm *PromptManager) GetPrompt(ctx context.Context, name string, variables map[string]interface{}) (*PromptResult, error) {
	for i, server := range pm.servers {
		serverName := fmt.Sprintf("server-%d", i)

		// First check if this server has the prompt
		prompts, err := server.ListPrompts(ctx)
		if err != nil {
			continue
		}

		var foundPrompt *contracts.MCPPrompt
		for _, p := range prompts {
			if p.Name == name {
				foundPrompt = &p
				break
			}
		}

		if foundPrompt == nil {
			continue
		}

		// Get the prompt with variables
		result, err := server.GetPrompt(ctx, name, variables)
		if err != nil {
			pm.logger.Warn(ctx, "Failed to get prompt from server", map[string]interface{}{
				"server": serverName,
				"prompt": name,
				"error":  err.Error(),
			})
			continue
		}

		pm.logger.Debug(ctx, "Successfully retrieved prompt", map[string]interface{}{
			"server": serverName,
			"prompt": name,
		})

		return &PromptResult{
			Server: server,
			Prompt: *foundPrompt,
			Result: *result,
		}, nil
	}

	return nil, fmt.Errorf("prompt not found on any server: %s", name)
}

// ExecutePromptTemplate executes a prompt template with variables and returns rendered content
func (pm *PromptManager) ExecutePromptTemplate(ctx context.Context, promptName string, variables map[string]interface{}) (string, error) {
	promptResult, err := pm.GetPrompt(ctx, promptName, variables)
	if err != nil {
		return "", err
	}

	// If we have a single prompt string, return it
	if promptResult.Result.Prompt != "" {
		return promptResult.Result.Prompt, nil
	}

	// If we have messages, combine them into a single string
	if len(promptResult.Result.Messages) > 0 {
		var parts []string
		for _, msg := range promptResult.Result.Messages {
			if msg.Role != "" {
				parts = append(parts, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
			} else {
				parts = append(parts, msg.Content)
			}
		}
		return strings.Join(parts, "\n"), nil
	}

	return "", fmt.Errorf("prompt %s returned no content", promptName)
}

// GetPromptsByCategory returns prompts filtered by category (from metadata)
func (pm *PromptManager) GetPromptsByCategory(ctx context.Context, category string) ([]PromptMatch, error) {
	var matches []PromptMatch

	for _, server := range pm.servers {
		prompts, err := server.ListPrompts(ctx)
		if err != nil {
			continue
		}

		for _, prompt := range prompts {
			if pm.matchesCategory(prompt, category) {
				matches = append(matches, PromptMatch{
					Server: server,
					Prompt: prompt,
				})
			}
		}
	}

	return matches, nil
}

// ValidatePromptVariables checks if all required variables are provided
func (pm *PromptManager) ValidatePromptVariables(prompt contracts.MCPPrompt, variables map[string]interface{}) error {
	var missingRequired []string

	for _, arg := range prompt.Arguments {
		if arg.Required {
			if _, exists := variables[arg.Name]; !exists {
				missingRequired = append(missingRequired, arg.Name)
			}
		}
	}

	if len(missingRequired) > 0 {
		return fmt.Errorf("missing required variables: %s", strings.Join(missingRequired, ", "))
	}

	return nil
}

// BuildVariablesFromTemplate builds variables from a Go template string
func (pm *PromptManager) BuildVariablesFromTemplate(templateStr string, data interface{}) (map[string]interface{}, error) {
	tmpl, err := template.New("variables").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// For now, return empty map - this would need more sophisticated parsing
	// to extract variable values from the rendered template
	return make(map[string]interface{}), nil
}
