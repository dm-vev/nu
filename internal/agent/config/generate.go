package config

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"nu/internal/contracts"
)

// GenerateConfigFromSystemPrompt uses the LLM to generate agent and task configurations from a system prompt
func GenerateConfigFromSystemPrompt(ctx context.Context, llm contracts.LLM, systemPrompt string) (AgentConfig, []TaskConfig, error) {
	if systemPrompt == "" {
		return AgentConfig{}, nil, fmt.Errorf("system prompt cannot be empty")
	}

	// Create a prompt for the LLM to generate agent and task configurations
	prompt := fmt.Sprintf(`
Based on the following system prompt that defines an AI agent's role, create YAML configurations for the agent and potential tasks it can perform.

System prompt:
%s

I need you to create:
1. An agent configuration with role, goal, and backstory
2. At least 2 task configurations that this agent can perform, with description and expected output

Format your response as valid YAML with the following structure (no prose, just YAML):

agent:
  role: >
    [Agent's role/title]
  goal: >
    [Agent's primary goal]
  backstory: >
    [Agent's backstory]

tasks:
  task1_name:
    description: >
      [Description of the first task]
    expected_output: >
      [Expected output format and content]

  task2_name:
    description: >
      [Description of the second task]
    expected_output: >
      [Expected output format and content]
    output_file: task2_output.md  # Optional
`, systemPrompt)

	// Generate the configurations using the LLM
	response, err := llm.Generate(ctx, prompt)
	if err != nil {
		return AgentConfig{}, nil, fmt.Errorf("failed to generate configurations: %w", err)
	}

	// Parse the YAML response
	var configs struct {
		Agent AgentConfig           `yaml:"agent"`
		Tasks map[string]TaskConfig `yaml:"tasks"`
	}

	if err := yaml.Unmarshal([]byte(response), &configs); err != nil {
		// Try to extract just the YAML part if there's prose around it
		yamlStart := strings.Index(response, "agent:")
		if yamlStart == -1 {
			return AgentConfig{}, nil, fmt.Errorf("failed to find agent configuration in response: %w", err)
		}

		// Find the end of the YAML block
		var yamlEnd int
		lines := strings.Split(response[yamlStart:], "\n")
		for i, line := range lines {
			if line == "```" || line == "---" {
				yamlEnd = yamlStart + strings.Index(response[yamlStart:], line)
				break
			}
			if i == len(lines)-1 {
				yamlEnd = len(response)
			}
		}

		yamlContent := response[yamlStart:yamlEnd]

		if err := yaml.Unmarshal([]byte(yamlContent), &configs); err != nil {
			return AgentConfig{}, nil, fmt.Errorf("failed to parse generated configurations: %w", err)
		}
	}

	// Convert tasks map to slice
	taskConfigs := make([]TaskConfig, 0, len(configs.Tasks))
	for name, taskConfig := range configs.Tasks {
		// Set the agent name field to the task name since we're creating these for the same agent
		taskConfig.Agent = name
		taskConfigs = append(taskConfigs, taskConfig)
	}

	return configs.Agent, taskConfigs, nil
}
