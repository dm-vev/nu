package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nu/internal/contracts"
	"nu/internal/llm/openai"
)

func (a *Agent) runWithoutExecutionPlanWithToolsTracked(ctx context.Context, input string, tools []contracts.Tool) (string, error) {
	prompt := input

	var response string
	var err error

	generateOptions := []contracts.GenerateOption{}
	if a.systemPrompt != "" {
		a.logger.Debug(context.Background(), fmt.Sprintf("Using system prompt (length=%d)", len(a.systemPrompt)), nil)
		generateOptions = append(generateOptions, openai.WithSystemMessage(a.systemPrompt))
	} else {
		a.logger.Warn(context.Background(), fmt.Sprintf("No system prompt set for agent %s", a.name), nil)
	}

	if a.responseFormat != nil {
		generateOptions = append(generateOptions, openai.WithResponseFormat(*a.responseFormat))
	}

	if a.llmConfig != nil {
		generateOptions = append(generateOptions, func(options *contracts.GenerateOptions) {
			options.LLMConfig = a.llmConfig
		})
	}

	generateOptions = append(generateOptions, contracts.WithMaxIterations(a.maxIterations))
	generateOptions = append(generateOptions, contracts.WithDisableFinalSummary(a.disableFinalSummary))

	if a.memory != nil {
		generateOptions = append(generateOptions, contracts.WithMemory(a.memory))
	}

	if a.cacheConfig != nil {
		generateOptions = append(generateOptions, func(options *contracts.GenerateOptions) {
			options.CacheConfig = a.cacheConfig
		})
	}

	tracker := getUsageTracker(ctx)

	if len(tools) > 0 {
		// Record tool invocations as the LLM actually calls them, not the
		// full set of available tools (#305).
		toolsForLLM := wrapToolsWithTracker(tools, tracker)

		if tracker != nil && tracker.detailed {
			llmResp, err := a.llm.GenerateWithToolsDetailed(ctx, prompt, toolsForLLM, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
			response = llmResp.Content
			tracker.addLLMUsage(llmResp.Usage, llmResp.Model)
		} else {
			response, err = a.llm.GenerateWithTools(ctx, prompt, toolsForLLM, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
		}
	} else {
		if tracker != nil && tracker.detailed {
			llmResp, err := a.llm.GenerateDetailed(ctx, prompt, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
			response = llmResp.Content
			tracker.addLLMUsage(llmResp.Usage, llmResp.Model)
		} else {
			response, err = a.llm.Generate(ctx, prompt, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
		}
	}

	// Apply guardrails to output if available
	if a.guardrails != nil {
		guardedResponse, err := a.guardrails.ProcessOutput(ctx, response)
		if err != nil {
			return "", fmt.Errorf("guardrails error: %w", err)
		}
		response = guardedResponse
	}

	// Add agent message to memory
	if a.memory != nil {
		if err := a.memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: response,
		}); err != nil {
			return "", fmt.Errorf("failed to add agent message to memory: %w", err)
		}
	}

	return response, nil
}

// isStructuredJSONResponse checks if a message content is a structured JSON response
func isStructuredJSONResponse(content string) bool {
	trimmed := strings.TrimSpace(content)
	return strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
}

// convertToHumanReadable converts a JSON response to a human-readable format
// to avoid confusing the LLM with raw JSON in conversation history
func convertToHumanReadable(jsonContent string) string {
	// Try to parse the JSON to extract key information
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &jsonMap); err != nil {
		// If parsing fails, return a generic summary
		return "[Generated structured response]"
	}

	// Convert JSON to human-readable format - works with any JSON structure
	var parts []string

	for key, value := range jsonMap {
		switch v := value.(type) {
		case string:
			if v != "" && v != "null" {
				parts = append(parts, fmt.Sprintf("%s: %s", key, v))
			}
		case []interface{}:
			if len(v) > 0 {
				if str, ok := v[0].(string); ok && str != "" {
					parts = append(parts, fmt.Sprintf("%s: %s", key, str))
				}
			}
		case bool:
			parts = append(parts, fmt.Sprintf("%s: %t", key, v))
		case float64, int:
			parts = append(parts, fmt.Sprintf("%s: %v", key, v))
		}
	}

	if len(parts) == 0 {
		return "[Generated structured response]"
	}

	// Limit to most important parts to keep summary concise
	if len(parts) > 3 {
		parts = parts[:3]
	}

	return "[AI: " + strings.Join(parts, ", ") + "]"
}

// isAskingAboutRole determines if the user is asking about the agent's role or identity
func (a *Agent) isAskingAboutRole(input string) bool {
	// Convert input to lowercase for case-insensitive matching
	lowerInput := strings.ToLower(input)

	// Common phrases that indicate a user asking about the agent's role
	roleQueries := []string{
		"what are you",
		"who are you",
		"what is your role",
		"what do you do",
		"what can you do",
		"what is your purpose",
		"what is your function",
		"tell me about yourself",
		"introduce yourself",
		"what are your capabilities",
		"what are you designed to do",
		"what's your job",
		"what kind of assistant are you",
		"your role",
		"your expertise",
		"what are you expert in",
		"what are you specialized in",
		"your specialty",
		"what's your specialty",
	}

	// Check if any of the role query phrases are in the input
	for _, query := range roleQueries {
		if strings.Contains(lowerInput, query) {
			return true
		}
	}

	return false
}

// generateRoleResponse creates a response based on the agent's system prompt
func (a *Agent) generateRoleResponse() string {
	// If the prompt is empty, return a generic response
	if a.systemPrompt == "" || a.llm == nil {
		return "I'm an AI assistant designed to help you with various tasks and answer your questions. How can I assist you today?"
	}

	// Create a prompt that asks the LLM to generate a role description based on the system prompt
	agentName := "an AI assistant"
	if a.name != "" {
		agentName = a.name
	}

	prompt := fmt.Sprintf(`Based on the following system prompt that defines your role and capabilities,
generate a brief, natural-sounding response (3-5 sentences) introducing yourself to a user who asked what you can do.
You are named "%s".
Do not directly quote from the system prompt, but create a conversational first-person response that captures your
purpose, expertise, and how you can help. The response should feel like a natural conversation, not like reading documentation.

System prompt:
%s

Your response should:
1. Introduce yourself using first-person perspective, mentioning your name ("%s")
2. Briefly explain your specialization or purpose
3. Mention 2-3 key areas you can help with
4. End with a friendly question about how you can assist the user

Response:`, agentName, a.systemPrompt, agentName)

	// Generate a response using the LLM with the system prompt as context
	generateOptions := []contracts.GenerateOption{}

	// Use the same system prompt to ensure consistent persona
	generateOptions = append(generateOptions, openai.WithSystemMessage(a.systemPrompt))

	// Generate the response
	response, err := a.llm.Generate(context.Background(), prompt, generateOptions...)
	if err != nil {
		// Fallback to a simple response in case of errors
		if a.name != "" {
			return fmt.Sprintf("I'm %s, an AI assistant based on the role defined in my system prompt. How can I help you today?", a.name)
		}
		return "I'm an AI assistant based on the role defined in my system prompt. How can I help you today?"
	}

	return response
}
