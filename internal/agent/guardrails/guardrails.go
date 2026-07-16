package guardrails

import (
	"context"
	"fmt"

	"nu/internal/telemetry"
)

// GuardrailType represents the type of guardrail
type GuardrailType string

const (
	// ContentFilterGuardrail filters content for inappropriate material
	GuardrailTypeContentFilter GuardrailType = "content_filter"

	// TokenLimitGuardrail limits the number of tokens in a request or response
	GuardrailTypeTokenLimit GuardrailType = "token_limit"

	// PiiFilterGuardrail filters personally identifiable information
	GuardrailTypePIIFilter GuardrailType = "pii_filter"

	// ToolRestrictionGuardrail restricts which tools can be used
	GuardrailTypeToolRestriction GuardrailType = "tool_restriction"

	// RateLimitGuardrail limits the rate of requests
	GuardrailTypeRateLimit GuardrailType = "rate_limit"
)

// GuardrailAction represents the action taken when a guardrail is triggered.
type GuardrailAction string

const (
	// BlockAction blocks the request or response
	GuardrailActionBlock GuardrailAction = "block"

	// RedactAction redacts the sensitive content
	GuardrailActionRedact GuardrailAction = "redact"

	// WarnAction allows the content but logs a warning
	GuardrailActionWarn GuardrailAction = "warn"
)

// Guardrail represents a guardrail that can be applied to requests and responses
type Guardrail interface {
	// Type returns the type of guardrail
	Type() GuardrailType

	// CheckRequest checks if a request violates the guardrail
	CheckRequest(ctx context.Context, request string) (bool, string, error)

	// CheckResponse checks if a response violates the guardrail
	CheckResponse(ctx context.Context, response string) (bool, string, error)

	// Action returns the action to take when the guardrail is triggered
	Action() GuardrailAction
}

// GuardrailPipeline applies guardrails in order.
type GuardrailPipeline struct {
	guardrails []Guardrail
	logger     telemetry.Logger
}

// NewGuardrailPipeline creates a guardrail pipeline.
func NewGuardrailPipeline(guardrails []Guardrail, logger telemetry.Logger) *GuardrailPipeline {
	return &GuardrailPipeline{
		guardrails: guardrails,
		logger:     logger,
	}
}

// ProcessRequest processes a request through the guardrails pipeline
func (p *GuardrailPipeline) ProcessRequest(ctx context.Context, request string) (string, error) {
	processedRequest := request

	for _, guardrail := range p.guardrails {
		triggered, modified, err := guardrail.CheckRequest(ctx, processedRequest)
		if err != nil {
			p.logger.Error(ctx, "Guardrail check failed", map[string]interface{}{
				"guardrail_type": guardrail.Type(),
				"error":          err.Error(),
			})
			return "", fmt.Errorf("guardrail check failed: %w", err)
		}

		if triggered {
			p.logger.Info(ctx, "Guardrail triggered", map[string]interface{}{
				"guardrail_type": guardrail.Type(),
				"action":         guardrail.Action(),
			})

			switch guardrail.Action() {
			case GuardrailActionBlock:
				return "", fmt.Errorf("request blocked by %s guardrail", guardrail.Type())
			case GuardrailActionRedact:
				processedRequest = modified
			case GuardrailActionWarn:
				// Continue with original request but log warning
				p.logger.Warn(ctx, "Guardrail warning", map[string]interface{}{
					"guardrail_type": guardrail.Type(),
					"original":       processedRequest,
					"modified":       modified,
				})
			}
		}
	}

	return processedRequest, nil
}

// ProcessResponse processes a response through the guardrails pipeline
func (p *GuardrailPipeline) ProcessResponse(ctx context.Context, response string) (string, error) {
	processedResponse := response

	for _, guardrail := range p.guardrails {
		triggered, modified, err := guardrail.CheckResponse(ctx, processedResponse)
		if err != nil {
			p.logger.Error(ctx, "Guardrail check failed", map[string]interface{}{
				"guardrail_type": guardrail.Type(),
				"error":          err.Error(),
			})
			return "", fmt.Errorf("guardrail check failed: %w", err)
		}

		if triggered {
			p.logger.Info(ctx, "Guardrail triggered", map[string]interface{}{
				"guardrail_type": guardrail.Type(),
				"action":         guardrail.Action(),
			})

			switch guardrail.Action() {
			case GuardrailActionBlock:
				return "", fmt.Errorf("response blocked by %s guardrail", guardrail.Type())
			case GuardrailActionRedact:
				processedResponse = modified
			case GuardrailActionWarn:
				// Continue with original response but log warning
				p.logger.Warn(ctx, "Guardrail warning", map[string]interface{}{
					"guardrail_type": guardrail.Type(),
					"original":       processedResponse,
					"modified":       modified,
				})
			}
		}
	}

	return processedResponse, nil
}

// AddGuardrail adds a guardrail to the pipeline
func (p *GuardrailPipeline) AddGuardrail(guardrail Guardrail) {
	p.guardrails = append(p.guardrails, guardrail)
}
