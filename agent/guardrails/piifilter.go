package guardrails

import (
	"context"
	"regexp"
)

// PIIFilterGuardrail filters personally identifiable information.
type PIIFilterGuardrail struct {
	patterns map[string]*regexp.Regexp
	action   GuardrailAction
}

// NewPIIFilterGuardrail creates a PII filter guardrail.
func NewPIIFilterGuardrail(action GuardrailAction) *PIIFilterGuardrail {
	patterns := map[string]*regexp.Regexp{
		"email":       regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
		"phone":       regexp.MustCompile(`\b(\+\d{1,2}\s)?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}\b`),
		"ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		"credit_card": regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`),
		"ip_address":  regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
	}

	return &PIIFilterGuardrail{
		patterns: patterns,
		action:   action,
	}
}

// Type returns the type of guardrail
func (p *PIIFilterGuardrail) Type() GuardrailType {
	return GuardrailTypePIIFilter
}

// CheckRequest checks if a request violates the guardrail
func (p *PIIFilterGuardrail) CheckRequest(ctx context.Context, request string) (bool, string, error) {
	modified := request
	triggered := false

	for name, pattern := range p.patterns {
		if pattern.MatchString(modified) {
			triggered = true
			modified = pattern.ReplaceAllString(modified, "[REDACTED "+name+"]")
		}
	}

	return triggered, modified, nil
}

// CheckResponse checks if a response violates the guardrail
func (p *PIIFilterGuardrail) CheckResponse(ctx context.Context, response string) (bool, string, error) {
	modified := response
	triggered := false

	for name, pattern := range p.patterns {
		if pattern.MatchString(modified) {
			triggered = true
			modified = pattern.ReplaceAllString(modified, "[REDACTED "+name+"]")
		}
	}

	return triggered, modified, nil
}

// Action returns the action to take when the guardrail is triggered
func (p *PIIFilterGuardrail) Action() GuardrailAction {
	return p.action
}
