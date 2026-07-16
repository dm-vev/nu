package guardrails

import (
	"context"
	"regexp"
	"strings"
)

// ContentFilterGuardrail filters configured words from content.
type ContentFilterGuardrail struct {
	blockedWords []string
	action       GuardrailAction
	regex        *regexp.Regexp
}

// NewContentFilterGuardrail creates a content filter guardrail.
func NewContentFilterGuardrail(blockedWords []string, action GuardrailAction) *ContentFilterGuardrail {
	// Escape special characters and join with OR
	pattern := strings.Join(blockedWords, "|")
	regex := regexp.MustCompile(`(?i)\b(` + pattern + `)\b`)

	return &ContentFilterGuardrail{
		blockedWords: blockedWords,
		action:       action,
		regex:        regex,
	}
}

// Type returns the type of guardrail
func (c *ContentFilterGuardrail) Type() GuardrailType {
	return GuardrailTypeContentFilter
}

// CheckRequest checks if a request violates the guardrail
func (c *ContentFilterGuardrail) CheckRequest(ctx context.Context, request string) (bool, string, error) {
	if c.regex.MatchString(request) {
		modified := c.regex.ReplaceAllString(request, "****")
		return true, modified, nil
	}
	return false, request, nil
}

// CheckResponse checks if a response violates the guardrail
func (c *ContentFilterGuardrail) CheckResponse(ctx context.Context, response string) (bool, string, error) {
	if c.regex.MatchString(response) {
		modified := c.regex.ReplaceAllString(response, "****")
		return true, modified, nil
	}
	return false, response, nil
}

// Action returns the action to take when the guardrail is triggered
func (c *ContentFilterGuardrail) Action() GuardrailAction {
	return c.action
}
