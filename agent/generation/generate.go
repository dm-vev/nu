package generation

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/agent/execution"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/llm/openai"
)

func (s *Service) Generate(ctx context.Context, input string, tools []contracts.Tool) (string, error) {
	prompt := input

	var response string
	var err error

	generateOptions := []contracts.GenerateOption{}
	if s.SystemPrompt != "" {
		s.Logger.Debug(context.Background(), fmt.Sprintf("Using system prompt (length=%d)", len(s.SystemPrompt)), nil)
		generateOptions = append(generateOptions, openai.WithSystemMessage(s.SystemPrompt))
	} else {
		s.Logger.Warn(context.Background(), fmt.Sprintf("No system prompt set for agent %s", s.Name), nil)
	}

	if s.ResponseFormat != nil {
		generateOptions = append(generateOptions, openai.WithResponseFormat(*s.ResponseFormat))
	}

	if s.LLMConfig != nil {
		generateOptions = append(generateOptions, func(options *contracts.GenerateOptions) {
			options.LLMConfig = s.LLMConfig
		})
	}

	generateOptions = append(generateOptions, contracts.WithMaxIterations(s.MaxIterations))
	generateOptions = append(generateOptions, contracts.WithDisableFinalSummary(s.DisableFinalSummary))

	if s.Memory != nil {
		generateOptions = append(generateOptions, contracts.WithMemory(s.Memory))
	}

	if s.CacheConfig != nil {
		generateOptions = append(generateOptions, func(options *contracts.GenerateOptions) {
			options.CacheConfig = s.CacheConfig
		})
	}

	tracker := execution.TrackerFromContext(ctx)

	if len(tools) > 0 {
		// Record tool invocations as the LLM actually calls them, not the
		// full set of available tools (#305).
		toolsForLLM := execution.WrapToolsWithTracker(tools, tracker)

		if tracker != nil && tracker.Detailed() {
			llmResp, err := s.LLM.GenerateWithToolsDetailed(ctx, prompt, toolsForLLM, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
			response = llmResp.Content
			tracker.AddLLMUsage(llmResp.Usage, llmResp.Model)
		} else {
			response, err = s.LLM.GenerateWithTools(ctx, prompt, toolsForLLM, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
		}
	} else {
		if tracker != nil && tracker.Detailed() {
			llmResp, err := s.LLM.GenerateDetailed(ctx, prompt, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
			response = llmResp.Content
			tracker.AddLLMUsage(llmResp.Usage, llmResp.Model)
		} else {
			response, err = s.LLM.Generate(ctx, prompt, generateOptions...)
			if err != nil {
				return "", fmt.Errorf("failed to generate response: %w", err)
			}
		}
	}

	// Apply guardrails to output if available
	if s.Guardrails != nil {
		guardedResponse, err := s.Guardrails.ProcessOutput(ctx, response)
		if err != nil {
			return "", fmt.Errorf("guardrails error: %w", err)
		}
		response = guardedResponse
	}

	// Add agent message to memory
	if s.Memory != nil {
		if err := s.Memory.AddMessage(ctx, contracts.Message{
			Role:    contracts.RoleAssistant,
			Content: response,
		}); err != nil {
			return "", fmt.Errorf("failed to add agent message to memory: %w", err)
		}
	}

	return response, nil
}
