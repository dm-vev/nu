package mcp

import (
	"fmt"
	"strings"

	"nu/internal/contracts"
)

type PromptMatch struct {
	Server contracts.MCPServer
	Prompt contracts.MCPPrompt
}

type PromptResult struct {
	Server contracts.MCPServer
	Prompt contracts.MCPPrompt
	Result contracts.MCPPromptResult
}

type PromptCategory struct {
	Name        string
	Description string
	Prompts     []PromptMatch
}

func (pm *PromptManager) matchesPattern(prompt contracts.MCPPrompt, pattern string) bool {
	pattern = strings.ToLower(pattern)
	if strings.Contains(strings.ToLower(prompt.Name), pattern) ||
		strings.Contains(strings.ToLower(prompt.Description), pattern) {
		return true
	}
	for key, value := range prompt.Metadata {
		if strings.Contains(strings.ToLower(key), pattern) ||
			strings.Contains(strings.ToLower(value), pattern) {
			return true
		}
	}
	return false
}

func (pm *PromptManager) matchesCategory(prompt contracts.MCPPrompt, category string) bool {
	if prompt.Metadata == nil {
		return false
	}
	promptCategory, exists := prompt.Metadata["category"]
	if !exists {
		for key, value := range prompt.Metadata {
			if strings.ToLower(key) == "category" || strings.ToLower(key) == "type" || strings.ToLower(key) == "group" {
				promptCategory = value
				break
			}
		}
	}
	if promptCategory == "" {
		return false
	}
	return strings.EqualFold(promptCategory, category)
}

func GetPromptParameterInfo(prompt contracts.MCPPrompt) string {
	if len(prompt.Arguments) == 0 {
		return "No parameters required"
	}
	var parts []string
	for _, arg := range prompt.Arguments {
		paramInfo := arg.Name
		if arg.Type != "" {
			paramInfo += fmt.Sprintf(" (%s)", arg.Type)
		}
		if arg.Required {
			paramInfo += " *required*"
		}
		if arg.Description != "" {
			paramInfo += fmt.Sprintf(" - %s", arg.Description)
		}
		if arg.Default != nil {
			paramInfo += fmt.Sprintf(" (default: %v)", arg.Default)
		}
		parts = append(parts, paramInfo)
	}
	return "Parameters:\n" + strings.Join(parts, "\n")
}

func SuggestPromptVariables(prompt contracts.MCPPrompt, context map[string]interface{}) map[string]interface{} {
	suggested := make(map[string]interface{})
	for _, arg := range prompt.Arguments {
		if arg.Default != nil {
			suggested[arg.Name] = arg.Default
			continue
		}
		if value, exists := context[arg.Name]; exists {
			suggested[arg.Name] = value
			continue
		}
		switch strings.ToLower(arg.Name) {
		case "name", "username", "user":
			if user, exists := context["user"]; exists {
				suggested[arg.Name] = user
			}
		case "project", "repo", "repository":
			if project, exists := context["project"]; exists {
				suggested[arg.Name] = project
			}
		case "language", "lang":
			suggested[arg.Name] = "go"
		case "format", "output_format":
			suggested[arg.Name] = "markdown"
		}
	}
	return suggested
}
