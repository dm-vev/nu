package guardrails

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dm-vev/nu/internal/multitenancy"
)

// RateLimitGuardrail limits requests per organization.
type RateLimitGuardrail struct {
	requestsPerMinute int
	requestCounts     map[string][]time.Time
	mu                sync.Mutex
	action            GuardrailAction
}

// NewRateLimitGuardrail creates a rate limit guardrail.
func NewRateLimitGuardrail(requestsPerMinute int, action GuardrailAction) *RateLimitGuardrail {
	return &RateLimitGuardrail{
		requestsPerMinute: requestsPerMinute,
		requestCounts:     make(map[string][]time.Time),
		action:            action,
	}
}

// Type returns the type of guardrail
func (r *RateLimitGuardrail) Type() GuardrailType {
	return GuardrailTypeRateLimit
}

// CheckRequest checks if a request violates the guardrail
func (r *RateLimitGuardrail) CheckRequest(ctx context.Context, request string) (bool, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get organization ID from context
	orgID, err := multitenancy.GetOrgID(ctx)
	if err != nil {
		// If no organization ID is found, use a default key
		orgID = "default"
	}

	// Get current time
	now := time.Now()

	// Clean up old requests (older than 1 minute)
	var recentRequests []time.Time
	for _, t := range r.requestCounts[orgID] {
		if now.Sub(t) < time.Minute {
			recentRequests = append(recentRequests, t)
		}
	}
	r.requestCounts[orgID] = recentRequests

	// Check if rate limit is exceeded
	if len(recentRequests) >= r.requestsPerMinute {
		return true, fmt.Sprintf("Rate limit exceeded: %d requests per minute", r.requestsPerMinute), nil
	}

	// Add current request to count
	r.requestCounts[orgID] = append(r.requestCounts[orgID], now)

	return false, request, nil
}

// CheckResponse checks if a response violates the guardrail
func (r *RateLimitGuardrail) CheckResponse(ctx context.Context, response string) (bool, string, error) {
	// Rate limits typically apply to requests, not responses
	return false, response, nil
}

// Action returns the action to take when the guardrail is triggered
func (r *RateLimitGuardrail) Action() GuardrailAction {
	return r.action
}
