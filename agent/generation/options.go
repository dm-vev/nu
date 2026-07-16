package generation

import (
	"context"
	"fmt"

	"github.com/dm-vev/nu/agent/execution"
	"github.com/dm-vev/nu/contracts"
)

func (s *Service) streamOptions() []contracts.GenerateOption {
	var options []contracts.GenerateOption
	if s.SystemPrompt != "" {
		options = append(options, func(opts *contracts.GenerateOptions) { opts.SystemMessage = s.SystemPrompt })
	}
	if s.LLMConfig != nil {
		options = append(options, func(opts *contracts.GenerateOptions) { opts.LLMConfig = s.LLMConfig })
	}
	if s.ResponseFormat != nil {
		options = append(options, func(opts *contracts.GenerateOptions) { opts.ResponseFormat = s.ResponseFormat })
	}
	if s.MaxIterations > 0 {
		options = append(options, contracts.WithMaxIterations(s.MaxIterations))
	}
	if s.Memory != nil {
		options = append(options, contracts.WithMemory(s.Memory))
	}
	if s.StreamConfig != nil {
		options = append(options, contracts.WithStreamConfig(*s.StreamConfig))
	}
	if s.CacheConfig != nil {
		options = append(options, func(opts *contracts.GenerateOptions) { opts.CacheConfig = s.CacheConfig })
	}
	return options
}

func withStreamForwarder(ctx context.Context, eventChan chan<- contracts.AgentStreamEvent) context.Context {
	forward := func(event contracts.AgentStreamEvent) {
		select {
		case eventChan <- event:
		case <-ctx.Done():
		}
	}
	return context.WithValue(ctx, contracts.StreamForwarderKey, contracts.StreamForwarder(forward))
}

func (s *Service) startStream(ctx context.Context, input string, tools []contracts.Tool, llm contracts.StreamingLLM, options []contracts.GenerateOption) (<-chan contracts.StreamEvent, error) {
	var (
		events <-chan contracts.StreamEvent
		err    error
	)
	if len(tools) > 0 {
		trackedTools := execution.WrapToolsWithTracker(tools, execution.TrackerFromContext(ctx))
		events, err = llm.GenerateWithToolsStream(ctx, input, trackedTools, options...)
	} else {
		events, err = llm.GenerateStream(ctx, input, options...)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to start LLM streaming: %w", err)
	}
	return events, nil
}
