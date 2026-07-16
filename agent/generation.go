package agent

import (
	"context"

	"github.com/dm-vev/nu/agent/generation"
	"github.com/dm-vev/nu/contracts"
)

func (a *Agent) generationService() *generation.Service {
	return generation.NewService(generation.Config{
		LLM:                 a.llm,
		Memory:              a.memory,
		Guardrails:          a.guardrails,
		Logger:              a.logger,
		SystemPrompt:        a.systemPrompt,
		Name:                a.name,
		ResponseFormat:      a.responseFormat,
		LLMConfig:           a.llmConfig,
		MaxIterations:       a.maxIterations,
		DisableFinalSummary: a.disableFinalSummary,
		CacheConfig:         a.cacheConfig,
		StreamConfig:        a.streamConfig,
	})
}

func (a *Agent) runWithoutExecutionPlanWithToolsTracked(ctx context.Context, input string, tools []contracts.Tool) (string, error) {
	return a.generationService().Generate(ctx, input, tools)
}

func (a *Agent) isAskingAboutRole(input string) bool {
	return a.generationService().IsAskingAboutRole(input)
}

func (a *Agent) generateRoleResponse() string {
	return a.generationService().GenerateRoleResponse()
}

func isStructuredJSONResponse(content string) bool {
	return generation.IsStructuredJSONResponse(content)
}

func convertToHumanReadable(content string) string {
	return generation.ConvertToHumanReadable(content)
}

func (a *Agent) runStreamingGeneration(
	ctx context.Context,
	input string,
	allTools []contracts.Tool,
	streamingLLM contracts.StreamingLLM,
	eventChan chan<- contracts.AgentStreamEvent,
) (int64, error) {
	return a.generationService().Stream(ctx, input, allTools, streamingLLM, eventChan)
}
