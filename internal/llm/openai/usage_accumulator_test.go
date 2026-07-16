package openai

import (
	"context"
	"testing"
)

func TestOpenAIUsageAccumulator_AccumulatesAcrossCalls(t *testing.T) {
	acc := &openAIUsageAccumulator{}
	acc.add(10, 5, 15, 0, "gpt-4o")
	acc.add(20, 8, 28, 2, "gpt-4o")

	usage, model, ok := acc.snapshot()
	if !ok {
		t.Fatal("expected accumulator to be marked touched")
	}
	if usage.InputTokens != 30 {
		t.Errorf("InputTokens = %d, want 30", usage.InputTokens)
	}
	if usage.OutputTokens != 13 {
		t.Errorf("OutputTokens = %d, want 13", usage.OutputTokens)
	}
	if usage.TotalTokens != 43 {
		t.Errorf("TotalTokens = %d, want 43", usage.TotalTokens)
	}
	if usage.ReasoningTokens != 2 {
		t.Errorf("ReasoningTokens = %d, want 2", usage.ReasoningTokens)
	}
	if model != "gpt-4o" {
		t.Errorf("model = %q, want gpt-4o", model)
	}
}

func TestOpenAIUsageAccumulator_UntouchedReturnsNil(t *testing.T) {
	acc := &openAIUsageAccumulator{}
	usage, _, ok := acc.snapshot()
	if ok {
		t.Errorf("expected !ok for untouched accumulator")
	}
	if usage != nil {
		t.Errorf("expected nil usage, got %+v", usage)
	}
}

func TestOpenAIUsageAccumulator_ContextRoundtrip(t *testing.T) {
	ctx := context.Background()
	if got := openAIGetUsageAccumulator(ctx); got != nil {
		t.Errorf("expected nil from empty ctx, got %p", got)
	}

	acc := &openAIUsageAccumulator{}
	ctx = openAIWithUsageAccumulator(ctx, acc)
	if got := openAIGetUsageAccumulator(ctx); got != acc {
		t.Errorf("expected to retrieve the installed accumulator")
	}
}
