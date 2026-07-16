package guardrails

import (
	"context"
	"fmt"
	"strings"
)

// GuardrailTokenCounter counts tokens for a token limit guardrail.
type GuardrailTokenCounter interface {
	CountTokens(text string) (int, error)
}

// GuardrailWordTokenCounter approximates tokens by counting fields.
type GuardrailWordTokenCounter struct{}

// CountTokens counts tokens in text (simple approximation)
func (s *GuardrailWordTokenCounter) CountTokens(text string) (int, error) {
	// Simple approximation: count words and punctuation
	return len(strings.Fields(text)), nil
}

// TokenLimitGuardrail limits the number of tokens in content.
type TokenLimitGuardrail struct {
	maxTokens    int
	counter      GuardrailTokenCounter
	action       GuardrailAction
	truncateMode string // "start", "end", or "middle"
}

// NewTokenLimitGuardrail creates a token limit guardrail.
func NewTokenLimitGuardrail(maxTokens int, counter GuardrailTokenCounter, action GuardrailAction, truncateMode string) *TokenLimitGuardrail {
	if counter == nil {
		counter = &GuardrailWordTokenCounter{}
	}

	if truncateMode == "" {
		truncateMode = "end"
	}

	return &TokenLimitGuardrail{
		maxTokens:    maxTokens,
		counter:      counter,
		action:       action,
		truncateMode: truncateMode,
	}
}

// Type returns the type of guardrail
func (t *TokenLimitGuardrail) Type() GuardrailType {
	return GuardrailTypeTokenLimit
}

// CheckRequest checks if a request violates the guardrail
func (t *TokenLimitGuardrail) CheckRequest(ctx context.Context, request string) (bool, string, error) {
	tokens, err := t.counter.CountTokens(request)
	if err != nil {
		return false, request, fmt.Errorf("failed to count tokens: %w", err)
	}

	if tokens > t.maxTokens {
		modified, err := t.truncate(request)
		if err != nil {
			return false, request, err
		}
		return true, modified, nil
	}

	return false, request, nil
}

// CheckResponse checks if a response violates the guardrail
func (t *TokenLimitGuardrail) CheckResponse(ctx context.Context, response string) (bool, string, error) {
	tokens, err := t.counter.CountTokens(response)
	if err != nil {
		return false, response, fmt.Errorf("failed to count tokens: %w", err)
	}

	if tokens > t.maxTokens {
		modified, err := t.truncate(response)
		if err != nil {
			return false, response, err
		}
		return true, modified, nil
	}

	return false, response, nil
}

// Action returns the action to take when the guardrail is triggered
func (t *TokenLimitGuardrail) Action() GuardrailAction {
	return t.action
}

// truncate truncates text to the maximum token limit
func (t *TokenLimitGuardrail) truncate(text string) (string, error) {
	words := strings.Fields(text)

	if len(words) <= t.maxTokens {
		return text, nil
	}

	switch t.truncateMode {
	case "start":
		return strings.Join(words[len(words)-t.maxTokens:], " "), nil
	case "middle":
		half := t.maxTokens / 2
		return strings.Join(words[:half], " ") + " ... " + strings.Join(words[len(words)-half:], " "), nil
	case "end":
		fallthrough
	default:
		return strings.Join(words[:t.maxTokens], " ") + " ...", nil
	}
}
