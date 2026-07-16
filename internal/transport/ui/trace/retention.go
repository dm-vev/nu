package trace

import (
	"encoding/json"
	"sort"
	"time"
)

// Helper methods

func (c *Collector) inferSpanType(name string) string {
	// Infer span type from name patterns
	if uiContains(name, []string{"generation", "llm", "completion", "chat"}) {
		return "generation"
	}
	if uiContains(name, []string{"tool", "function", "call"}) {
		return "tool_call"
	}
	if uiContains(name, []string{"event"}) {
		return "event"
	}
	return "span"
}

func (c *Collector) updateTraceSize(trace *Trace) {
	// Calculate approximate size of trace in bytes
	data, _ := json.Marshal(trace)
	oldSize := trace.SizeBytes
	trace.SizeBytes = len(data)
	c.currentSizeBytes += (trace.SizeBytes - oldSize)
}

func (c *Collector) enforceRetentionLimits() {
	// Remove old traces beyond max age
	cutoffTime := time.Now().Add(-c.maxAge)
	for id, trace := range c.traces {
		if trace.StartTime.Before(cutoffTime) {
			c.currentSizeBytes -= trace.SizeBytes
			delete(c.traces, id)
		}
	}

	// Remove oldest traces if over retention count
	if len(c.traces) > c.config.RetentionCount {
		// Get sorted trace IDs by start time
		type traceTime struct {
			id        string
			startTime time.Time
		}

		traceTimes := make([]traceTime, 0, len(c.traces))
		for id, trace := range c.traces {
			traceTimes = append(traceTimes, traceTime{id: id, startTime: trace.StartTime})
		}

		sort.Slice(traceTimes, func(i, j int) bool {
			return traceTimes[i].startTime.Before(traceTimes[j].startTime)
		})

		// Remove oldest traces
		toRemove := len(c.traces) - c.config.RetentionCount
		for i := 0; i < toRemove; i++ {
			trace := c.traces[traceTimes[i].id]
			c.currentSizeBytes -= trace.SizeBytes
			delete(c.traces, traceTimes[i].id)
		}
	}

	// Remove oldest traces if over size limit
	for c.currentSizeBytes > c.maxSizeBytes && len(c.traces) > 0 {
		// Find oldest trace
		var oldestID string
		var oldestTime time.Time
		for id, trace := range c.traces {
			if oldestID == "" || trace.StartTime.Before(oldestTime) {
				oldestID = id
				oldestTime = trace.StartTime
			}
		}

		if oldestID != "" {
			trace := c.traces[oldestID]
			c.currentSizeBytes -= trace.SizeBytes
			delete(c.traces, oldestID)
		}
	}
}

func (c *Collector) isTraceComplete(trace *Trace) bool {
	// Check if all spans in trace are complete
	for _, span := range trace.Spans {
		if span.EndTime == nil {
			// Check if span is still active
			if _, exists := c.activeSpans[span.ID]; exists {
				return false
			}
		}
	}
	return true
}

func (c *Collector) getTraceEndTime(trace *Trace) time.Time {
	var latestEnd time.Time
	for _, span := range trace.Spans {
		if span.EndTime != nil && span.EndTime.After(latestEnd) {
			latestEnd = *span.EndTime
		}
	}
	if latestEnd.IsZero() {
		return time.Now()
	}
	return latestEnd
}

func (c *Collector) countRunningTraces() int {
	count := 0
	for _, trace := range c.traces {
		if trace.Status == "running" {
			count++
		}
	}
	return count
}
