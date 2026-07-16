package openai

import (
	"context"
	"sync"

	"nu/internal/contracts"
)

// openAIUsageAccumulator collects token usage across the multiple API calls a
// single GenerateWithTools invocation makes (one per tool-loop iteration
// plus the final summary). Lets GenerateWithToolsDetailed return a
// total that reflects every underlying chat completion, not just the
// last one (#276).
type openAIUsageAccumulator struct {
	mu      sync.Mutex
	total   contracts.TokenUsage
	model   string
	touched bool
}

func (u *openAIUsageAccumulator) add(input, output, total, reasoning int, model string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.total.InputTokens += input
	u.total.OutputTokens += output
	u.total.TotalTokens += total
	u.total.ReasoningTokens += reasoning
	if u.model == "" {
		u.model = model
	}
	u.touched = true
}

func (u *openAIUsageAccumulator) snapshot() (*contracts.TokenUsage, string, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if !u.touched {
		return nil, "", false
	}
	t := u.total
	return &t, u.model, true
}

type openAIUsageCtxKey struct{}

func openAIWithUsageAccumulator(ctx context.Context, acc *openAIUsageAccumulator) context.Context {
	return context.WithValue(ctx, openAIUsageCtxKey{}, acc)
}

func openAIGetUsageAccumulator(ctx context.Context) *openAIUsageAccumulator {
	acc, _ := ctx.Value(openAIUsageCtxKey{}).(*openAIUsageAccumulator)
	return acc
}
