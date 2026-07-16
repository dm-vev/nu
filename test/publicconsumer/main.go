package consumer

import (
	"context"

	"github.com/dm-vev/nu/agent"
	"github.com/dm-vev/nu/contracts"
)

type llm struct{}

func (llm) Generate(context.Context, string, ...contracts.GenerateOption) (string, error) {
	return "", nil
}

func (llm) GenerateWithTools(context.Context, string, []contracts.Tool, ...contracts.GenerateOption) (string, error) {
	return "", nil
}

func (llm) GenerateDetailed(context.Context, string, ...contracts.GenerateOption) (*contracts.LLMResponse, error) {
	return &contracts.LLMResponse{}, nil
}

func (llm) GenerateWithToolsDetailed(
	context.Context,
	string,
	[]contracts.Tool,
	...contracts.GenerateOption,
) (*contracts.LLMResponse, error) {
	return &contracts.LLMResponse{}, nil
}

func (llm) Name() string {
	return "consumer-test"
}

func (llm) SupportsStreaming() bool {
	return false
}

func NewAgent() (*agent.Agent, error) {
	return agent.NewAgent(agent.WithLLM(llm{}), agent.WithSystemPrompt("external consumer"))
}
