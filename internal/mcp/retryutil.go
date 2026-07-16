package mcp

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"nu/internal/telemetry"
)

func containsIgnoreCase(str, substr string) bool {
	return len(str) >= len(substr) && strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func randomFloat() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}

// RetryWithExponentialBackoff is a utility function for retrying any operation
func RetryWithExponentialBackoff(ctx context.Context, operation func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}
	logger := telemetry.NewLogger()
	delay := config.InitialDelay
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		if attempt >= config.MaxAttempts {
			return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, err)
		}
		logger.Debug(ctx, "Retrying operation", map[string]interface{}{
			"attempt": attempt, "max_attempts": config.MaxAttempts,
			"delay_ms": delay.Milliseconds(), "error": err.Error(),
		})
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		delay = time.Duration(math.Min(
			float64(delay)*config.BackoffMultiplier,
			float64(config.MaxDelay),
		))
	}
	return fmt.Errorf("max retry attempts reached")
}
