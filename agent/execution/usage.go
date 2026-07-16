package execution

import (
	"context"
	"sync"

	"github.com/dm-vev/nu/contracts"
)

type contextKey string

const trackerKey contextKey = "Tracker"

type Tracker struct {
	totalUsage   *contracts.TokenUsage
	execSummary  *contracts.ExecutionSummary
	detailed     bool
	primaryModel string
	mu           sync.Mutex
}

// Detailed reports whether the tracker collects execution details.
func (ut *Tracker) Detailed() bool { return ut.detailed }

func NewTracker(detailed bool) *Tracker {
	return &Tracker{
		totalUsage: &contracts.TokenUsage{},
		execSummary: &contracts.ExecutionSummary{
			UsedTools:     []string{},
			UsedSubAgents: []string{},
		},
		detailed: detailed,
	}
}

func (ut *Tracker) AddLLMUsage(usage *contracts.TokenUsage, model string) {
	if !ut.detailed || usage == nil {
		return
	}

	ut.mu.Lock()
	defer ut.mu.Unlock()

	ut.totalUsage.InputTokens += usage.InputTokens
	ut.totalUsage.OutputTokens += usage.OutputTokens
	ut.totalUsage.TotalTokens += usage.TotalTokens
	ut.totalUsage.ReasoningTokens += usage.ReasoningTokens
	ut.execSummary.LLMCalls++

	if ut.primaryModel == "" && model != "" {
		ut.primaryModel = model
	}
}

func (ut *Tracker) AddToolCall(toolName string) {
	if !ut.detailed {
		return
	}

	ut.mu.Lock()
	defer ut.mu.Unlock()

	for _, used := range ut.execSummary.UsedTools {
		if used == toolName {
			return
		}
	}

	ut.execSummary.UsedTools = append(ut.execSummary.UsedTools, toolName)
	ut.execSummary.ToolCalls++
}

// AddSubAgentCall records a sub-agent invocation and its name once.
func (ut *Tracker) AddSubAgentCall(agentName string) {
	if !ut.detailed {
		return
	}

	ut.mu.Lock()
	defer ut.mu.Unlock()

	ut.execSummary.SubAgentCalls++
	if agentName == "" {
		return
	}
	for _, used := range ut.execSummary.UsedSubAgents {
		if used == agentName {
			return
		}
	}
	ut.execSummary.UsedSubAgents = append(ut.execSummary.UsedSubAgents, agentName)
}

func (ut *Tracker) SetExecutionTime(timeMs int64) {
	if !ut.detailed {
		return
	}

	ut.mu.Lock()
	defer ut.mu.Unlock()

	ut.execSummary.ExecutionTimeMs = timeMs
}

func (ut *Tracker) Results() (*contracts.TokenUsage, *contracts.ExecutionSummary, string) {
	if !ut.detailed {
		return nil, nil, ""
	}

	ut.mu.Lock()
	defer ut.mu.Unlock()

	return ut.totalUsage, ut.execSummary, ut.primaryModel
}

func WithTracker(ctx context.Context, tracker *Tracker) context.Context {
	return context.WithValue(ctx, trackerKey, tracker)
}

func TrackerFromContext(ctx context.Context) *Tracker {
	tracker, _ := ctx.Value(trackerKey).(*Tracker)
	return tracker
}
