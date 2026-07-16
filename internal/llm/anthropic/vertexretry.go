package anthropic

import (
	"context"
	"time"

	"github.com/dm-vev/nu/telemetry"
)

// AnthropicVertexRetryPolicy represents the retry policy (imported from pkg/retry)
type VertexRetryPolicy struct {
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaximumInterval    time.Duration
	MaximumAttempts    int32
}

// VertexRetryExecutor wraps retry execution with region rotation for Vertex AI
type VertexRetryExecutor struct {
	vertexConfig *VertexConfig
	policy       *VertexRetryPolicy
	logger       telemetry.Logger
}

// NewVertexRetryExecutor creates a new retry executor for Vertex AI with region rotation
func NewVertexRetryExecutor(vertexConfig *VertexConfig, policy *VertexRetryPolicy) *VertexRetryExecutor {
	return &VertexRetryExecutor{
		vertexConfig: vertexConfig,
		policy:       policy,
		logger:       telemetry.NewLogger(),
	}
}

// Execute executes the operation with retries and region rotation
func (e *VertexRetryExecutor) Execute(ctx context.Context, operation func() error) error {
	var lastErr error
	attempt := int32(0)
	currentInterval := e.policy.InitialInterval
	contextCancelled := false

	for attempt < e.policy.MaximumAttempts {
		select {
		case <-ctx.Done():
			if !contextCancelled {
				contextCancelled = true
				e.logger.Warn(ctx, "Context cancelled but continuing with retry attempts", map[string]interface{}{
					"attempt": attempt,
					"error":   ctx.Err(),
				})
			}
		default:
		}

		currentRegion := e.vertexConfig.GetCurrentRegion()
		e.logger.Debug(ctx, "Attempting operation", map[string]interface{}{
			"attempt":      attempt + 1,
			"max_attempts": e.policy.MaximumAttempts,
			"region":       currentRegion,
		})

		if err := operation(); err == nil {
			e.logger.Debug(ctx, "Operation succeeded", map[string]interface{}{
				"attempt": attempt + 1,
				"region":  currentRegion,
			})
			return nil
		} else {
			lastErr = err
			attempt++
			if attempt >= e.policy.MaximumAttempts {
				e.logger.Debug(ctx, "Maximum attempts reached", map[string]interface{}{
					"attempt": attempt,
					"error":   err.Error(),
					"region":  currentRegion,
				})
				break
			}

			e.vertexConfig.RotateRegion()
			nextRegion := e.vertexConfig.GetCurrentRegion()
			nextInterval := time.Duration(float64(currentInterval) * e.policy.BackoffCoefficient)
			if nextInterval > e.policy.MaximumInterval {
				nextInterval = e.policy.MaximumInterval
			}

			e.logger.Debug(ctx, "Operation failed, rotating region and scheduling retry", map[string]interface{}{
				"attempt":          attempt,
				"error":            err.Error(),
				"current_region":   currentRegion,
				"next_region":      nextRegion,
				"current_interval": currentInterval,
				"next_interval":    nextInterval,
			})

			select {
			case <-time.After(currentInterval):
				currentInterval = nextInterval
			case <-ctx.Done():
				if !contextCancelled {
					contextCancelled = true
					e.logger.Warn(ctx, "Context cancelled during retry delay, but continuing", map[string]interface{}{
						"attempt": attempt,
						"error":   ctx.Err(),
					})
				}
				time.Sleep(currentInterval)
				currentInterval = nextInterval
			}
		}
	}

	return lastErr
}
