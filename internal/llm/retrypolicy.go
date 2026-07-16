package llm

import "time"

// RetryPolicy defines the retry policy configuration
type RetryPolicy struct {
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaximumInterval    time.Duration
	MaximumAttempts    int32
}

// RetryOption represents a retry policy option
type RetryOption func(*RetryPolicy)

// WithRetryInitialInterval sets the initial interval for retries
func WithRetryInitialInterval(interval time.Duration) RetryOption {
	return func(p *RetryPolicy) {
		p.InitialInterval = interval
	}
}

// WithRetryBackoffCoefficient sets the backoff coefficient
func WithRetryBackoffCoefficient(coefficient float64) RetryOption {
	return func(p *RetryPolicy) {
		p.BackoffCoefficient = coefficient
	}
}

// WithRetryMaximumInterval sets the maximum interval between retries
func WithRetryMaximumInterval(interval time.Duration) RetryOption {
	return func(p *RetryPolicy) {
		p.MaximumInterval = interval
	}
}

// WithRetryMaxAttempts sets the maximum number of retry attempts
func WithRetryMaxAttempts(attempts int32) RetryOption {
	return func(p *RetryPolicy) {
		p.MaximumAttempts = attempts
	}
}

// NewRetryPolicy creates a new retry policy with default values
func NewRetryPolicy(opts ...RetryOption) *RetryPolicy {
	policy := &RetryPolicy{
		InitialInterval:    time.Second,       // Default 1s
		BackoffCoefficient: 2.0,               // Default exponential backoff
		MaximumInterval:    time.Second * 100, // Default 100s
		MaximumAttempts:    3,                 // Default 3 attempts
	}

	for _, opt := range opts {
		opt(policy)
	}

	return policy
}
