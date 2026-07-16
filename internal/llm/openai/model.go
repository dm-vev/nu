package openai

import (
	"context"
	"strings"
)

// openAIIsReasoningModel reports whether the model requires temperature 1.
func openAIIsReasoningModel(model string) bool {
	reasoningModels := []string{
		"o1-", "o1-mini", "o1-preview",
		"o3-", "o3-mini",
		"o4-", "o4-mini",
		"gpt-5", "gpt-5-mini", "gpt-5-nano",
	}

	for _, prefix := range reasoningModels {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

// getTemperatureForModel returns the temperature accepted by the model.
func (c *Client) getTemperatureForModel(requestedTemp float64) float64 {
	if openAIIsReasoningModel(c.Model) {
		if requestedTemp != 1.0 {
			c.logger.Debug(context.Background(), "Overriding temperature for reasoning model", map[string]interface{}{
				"model":                 c.Model,
				"requested_temperature": requestedTemp,
				"forced_temperature":    1.0,
				"reason":                "reasoning models only support temperature = 1",
			})
		}
		return 1.0
	}
	return requestedTemp
}
