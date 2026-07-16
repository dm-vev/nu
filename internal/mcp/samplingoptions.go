package mcp

import (
	"fmt"
	"strings"

	"nu/internal/contracts"
)

type SamplingOption func(*contracts.MCPSamplingRequest)

func WithSystemPrompt(prompt string) SamplingOption {
	return func(req *contracts.MCPSamplingRequest) { req.SystemPrompt = prompt }
}

func WithMaxTokens(maxTokens int) SamplingOption {
	return func(req *contracts.MCPSamplingRequest) { req.MaxTokens = &maxTokens }
}

func WithTemperature(temperature float64) SamplingOption {
	return func(req *contracts.MCPSamplingRequest) { req.Temperature = &temperature }
}

func WithModelHint(modelName string) SamplingOption {
	return func(req *contracts.MCPSamplingRequest) {
		if req.ModelPreferences == nil {
			req.ModelPreferences = &contracts.MCPModelPreferences{}
		}
		req.ModelPreferences.Hints = append(req.ModelPreferences.Hints, contracts.MCPModelHint{Name: modelName})
	}
}

func WithModelPreferences(cost, speed, intelligence float64) SamplingOption {
	return func(req *contracts.MCPSamplingRequest) {
		if req.ModelPreferences == nil {
			req.ModelPreferences = &contracts.MCPModelPreferences{}
		}
		req.ModelPreferences.CostPriority = cost
		req.ModelPreferences.SpeedPriority = speed
		req.ModelPreferences.IntelligencePriority = intelligence
	}
}

func WithStopSequences(sequences ...string) SamplingOption {
	return func(req *contracts.MCPSamplingRequest) { req.StopSequences = sequences }
}

func WithIncludeContext(context string) SamplingOption {
	return func(req *contracts.MCPSamplingRequest) { req.IncludeContext = context }
}

func ExtractCodeFromResponse(response *contracts.MCPSamplingResponse) string {
	if response == nil || response.Content.Type != "text" {
		return ""
	}
	text := response.Content.Text
	lines := strings.Split(text, "\n")
	var codeLines []string
	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			codeLines = append(codeLines, line)
		}
	}
	if len(codeLines) > 0 {
		return strings.Join(codeLines, "\n")
	}
	return text
}

func FormatConversationHistory(messages []contracts.MCPMessage) string {
	var formatted []string
	for _, msg := range messages {
		// Preserve the existing indexing order, including its behavior for an empty role.
		role := strings.ToUpper(msg.Role[:1]) + strings.ToLower(msg.Role[1:])
		if len(msg.Role) == 0 {
			role = ""
		}
		content := msg.Content.Text
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		formatted = append(formatted, fmt.Sprintf("%s: %s", role, content))
	}
	return strings.Join(formatted, "\n")
}
