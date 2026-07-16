package ui

import (
	"fmt"
	"sort"
)

// GetTraces returns all traces (newest first)
func (c *TraceCollector) GetTraces(limit, offset int) ([]Trace, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert map to slice and sort by start time (newest first)
	traces := make([]Trace, 0, len(c.traces))
	for _, trace := range c.traces {
		traces = append(traces, *trace)
	}

	sort.Slice(traces, func(i, j int) bool {
		return traces[i].StartTime.After(traces[j].StartTime)
	})

	total := len(traces)

	// Apply pagination
	if offset >= total {
		return []Trace{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return traces[offset:end], total
}

// GetTrace returns a specific trace by ID
func (c *TraceCollector) GetTrace(traceID string) (*Trace, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	trace, exists := c.traces[traceID]
	if !exists {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}

	// Return a copy
	traceCopy := *trace
	return &traceCopy, nil
}

// DeleteTrace deletes a trace by ID
func (c *TraceCollector) DeleteTrace(traceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	trace, exists := c.traces[traceID]
	if !exists {
		return fmt.Errorf("trace not found: %s", traceID)
	}

	c.currentSizeBytes -= trace.SizeBytes
	delete(c.traces, traceID)
	return nil
}

// GetStats returns trace statistics
func (c *TraceCollector) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalDuration int64
	var errorCount int
	toolUsage := make(map[string]int)

	for _, trace := range c.traces {
		if trace.Duration > 0 {
			totalDuration += trace.Duration
		}
		if trace.Status == "error" {
			errorCount++
		}

		// Count tool usage
		for _, span := range trace.Spans {
			if span.Type == "tool_call" {
				toolName := span.Name
				if name, ok := span.Attributes["tool_name"].(string); ok {
					toolName = name
				}
				toolUsage[toolName]++
			} else if uiContains(span.Name, []string{"tool", "function", "call"}) {
				// Also count spans with tool-like names
				toolName := span.Name
				if name, ok := span.Attributes["tool_name"].(string); ok {
					toolName = name
				}
				toolUsage[toolName]++
			}
		}
	}

	avgDuration := int64(0)
	if len(c.traces) > 0 {
		avgDuration = totalDuration / int64(len(c.traces))
	}

	return map[string]interface{}{
		"total_traces":      len(c.traces),
		"running_traces":    c.countRunningTraces(),
		"error_count":       errorCount,
		"error_rate":        float64(errorCount) / float64(max(len(c.traces), 1)),
		"avg_duration_ms":   avgDuration,
		"buffer_size_bytes": c.currentSizeBytes,
		"buffer_usage":      float64(c.currentSizeBytes) / float64(c.maxSizeBytes),
		"tool_usage":        toolUsage,
	}
}
