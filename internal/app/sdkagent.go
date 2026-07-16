package app

import (
	"context"

	"github.com/dm-vev/nu/agent"
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/agentui"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/tools/coding"
	"github.com/dm-vev/nu/telemetry"
)

func newAgent(opts Options, emit func(agentui.Event)) *agentui.Agent {
	agentTools := opts.Tools
	if agentTools == nil {
		agentTools = coding.Builtins(opts.CWD)
	}
	memory := opts.Memory
	if memory == nil {
		memory = conversation.NewConversationBuffer()
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
	builder := agentui.Builder(nil)
	if opts.BuildLLM != nil {
		builder = func(ctx context.Context, config agentui.Config, memory contracts.Memory) (contracts.StreamingAgent, error) {
			llm, err := opts.BuildLLM(ctx, config)
			if err != nil {
				return nil, err
			}
			return newSDKAgent(llm, memory, agentTools)
		}
	}
	return agentui.New(agentui.Options{
		Runner:  runner,
		Builder: builder,
		Memory:  memory,
		Config:  agentui.Config{ProviderID: opts.ProviderID, API: opts.API, Model: opts.Model},
		Emit:    emit,
	})
}

func newSDKAgent(llm contracts.LLM, memory contracts.Memory, tools []contracts.Tool) (*agent.Agent, error) {
	streamConfig := contracts.DefaultStreamConfig()
	// The upstream OpenAI tool stream suppresses even final text when this is false
	// and tools were offered but not called.
	streamConfig.IncludeIntermediateMessages = true
	return agent.NewAgent(
		agent.WithLLM(llm),
		agent.WithMemory(memory),
		agent.WithTools(tools...),
		agent.WithRequirePlanApproval(false),
		agent.WithMaxIterations(16),
		agent.WithStreamConfig(&streamConfig),
		agent.WithName("nu"),
		agent.WithLogger(discardSDKLogger{}),
	)
}

type discardSDKLogger struct{}

func (discardSDKLogger) Info(context.Context, string, map[string]any)  {}
func (discardSDKLogger) Warn(context.Context, string, map[string]any)  {}
func (discardSDKLogger) Error(context.Context, string, map[string]any) {}
func (discardSDKLogger) Debug(context.Context, string, map[string]any) {}

var _ telemetry.Logger = discardSDKLogger{}
