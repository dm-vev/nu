package app

import (
	"context"

	sdkagent "nu/internal/agent"
	agent "nu/internal/agentui"
	"nu/internal/contracts"
	sdkmemory "nu/internal/memory"
	"nu/internal/telemetry"
	"nu/internal/tools/coding"
)

func newAgent(opts Options, emit func(agent.Event)) *agent.Agent {
	agentTools := opts.Tools
	if agentTools == nil {
		agentTools = coding.Builtins(opts.CWD)
	}
	memory := opts.Memory
	if memory == nil {
		memory = sdkmemory.NewConversationBuffer()
	}
	runner := opts.Runner
	if runner == nil && opts.LLM != nil {
		var err error
		runner, err = newSDKAgent(opts.LLM, memory, agentTools)
		if err != nil {
			return nil
		}
	}
	if runner == nil {
		return nil
	}
	builder := agent.Builder(nil)
	if opts.BuildLLM != nil {
		builder = func(ctx context.Context, config agent.Config, memory contracts.Memory) (contracts.StreamingAgent, error) {
			llm, err := opts.BuildLLM(ctx, config)
			if err != nil {
				return nil, err
			}
			return newSDKAgent(llm, memory, agentTools)
		}
	}
	return agent.New(agent.Options{
		Runner:  runner,
		Builder: builder,
		Memory:  memory,
		Config:  agent.Config{ProviderID: opts.ProviderID, API: opts.API, Model: opts.Model},
		Emit:    emit,
	})
}

func newSDKAgent(llm contracts.LLM, memory contracts.Memory, tools []contracts.Tool) (*sdkagent.Agent, error) {
	streamConfig := contracts.DefaultStreamConfig()
	// The upstream OpenAI tool stream suppresses even final text when this is false
	// and tools were offered but not called.
	streamConfig.IncludeIntermediateMessages = true
	return sdkagent.NewAgent(
		sdkagent.WithLLM(llm),
		sdkagent.WithMemory(memory),
		sdkagent.WithTools(tools...),
		sdkagent.WithRequirePlanApproval(false),
		sdkagent.WithMaxIterations(16),
		sdkagent.WithStreamConfig(&streamConfig),
		sdkagent.WithName("nu"),
		sdkagent.WithLogger(discardSDKLogger{}),
	)
}

type discardSDKLogger struct{}

func (discardSDKLogger) Info(context.Context, string, map[string]any)  {}
func (discardSDKLogger) Warn(context.Context, string, map[string]any)  {}
func (discardSDKLogger) Error(context.Context, string, map[string]any) {}
func (discardSDKLogger) Debug(context.Context, string, map[string]any) {}

var _ telemetry.Logger = discardSDKLogger{}
