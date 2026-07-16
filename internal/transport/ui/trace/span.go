package trace

import (
	"context"
	"fmt"
	"time"
)

// End ends the span
func (s *uiCollectorSpan) End() {
	ctx := context.Background()
	s.collector.logger.Debug(ctx, "End() called for span", map[string]interface{}{
		"span_name": s.spanContext.span.Name,
	})

	if s.spanContext.wrappedSpan != nil {
		s.spanContext.wrappedSpan.End()
	}

	endTime := time.Now()

	s.collector.mu.Lock()
	defer s.collector.mu.Unlock()

	// Find the actual span in the trace and update it
	found := false
	for i := range s.spanContext.trace.Spans {
		if s.spanContext.trace.Spans[i].ID == s.spanContext.span.ID {
			s.spanContext.trace.Spans[i].EndTime = &endTime
			s.spanContext.trace.Spans[i].Duration = endTime.Sub(s.spanContext.trace.Spans[i].StartTime).Milliseconds()
			s.collector.logger.Debug(ctx, "Updated span with duration", map[string]interface{}{
				"span_id":     s.spanContext.span.ID,
				"duration_ms": s.spanContext.trace.Spans[i].Duration,
			})
			found = true
			break
		}
	}
	if !found {
		s.collector.logger.Warn(ctx, "Span not found in trace", map[string]interface{}{
			"span_id":  s.spanContext.span.ID,
			"trace_id": s.spanContext.trace.ID,
		})
	}

	// Remove from active spans
	delete(s.collector.activeSpans, s.spanContext.span.ID)
	s.collector.logger.Debug(ctx, "Removed span from active spans", map[string]interface{}{
		"span_id": s.spanContext.span.ID,
	})

	// Update trace if all spans are complete
	isComplete := s.collector.isTraceComplete(s.spanContext.trace)
	s.collector.logger.Debug(ctx, "Trace completion status", map[string]interface{}{
		"trace_id":    s.spanContext.trace.ID,
		"is_complete": isComplete,
	})
	if isComplete {
		// Only set status to completed if it's not already an error
		if s.spanContext.trace.Status != "error" {
			s.spanContext.trace.Status = "completed"
		}
		traceEndTime := s.collector.getTraceEndTime(s.spanContext.trace)
		s.spanContext.trace.EndTime = &traceEndTime
		s.spanContext.trace.Duration = traceEndTime.Sub(s.spanContext.trace.StartTime).Milliseconds()
		s.collector.logger.Debug(ctx, "Trace completed with duration", map[string]interface{}{
			"trace_id":    s.spanContext.trace.ID,
			"duration_ms": s.spanContext.trace.Duration,
		})
	}

	// Update trace size
	s.collector.updateTraceSize(s.spanContext.trace)
	s.collector.logger.Debug(ctx, "Trace size updated", map[string]interface{}{
		"trace_id":     s.spanContext.trace.ID,
		"total_traces": len(s.collector.traces),
	})
}

// AddEvent adds an event to the span
func (s *uiCollectorSpan) AddEvent(name string, attributes map[string]interface{}) {
	if s.spanContext.wrappedSpan != nil {
		s.spanContext.wrappedSpan.AddEvent(name, attributes)
	}

	event := TraceEvent{
		Name:       name,
		Timestamp:  time.Now(),
		Attributes: attributes,
	}

	s.collector.mu.Lock()
	defer s.collector.mu.Unlock()

	// Find the actual span in the trace and update it
	for i := range s.spanContext.trace.Spans {
		if s.spanContext.trace.Spans[i].ID == s.spanContext.span.ID {
			s.spanContext.trace.Spans[i].Events = append(s.spanContext.trace.Spans[i].Events, event)
			break
		}
	}
	s.collector.updateTraceSize(s.spanContext.trace)
}

// SetAttribute sets an attribute on the span
func (s *uiCollectorSpan) SetAttribute(key string, value interface{}) {
	if s.spanContext.wrappedSpan != nil {
		s.spanContext.wrappedSpan.SetAttribute(key, value)
	}

	s.collector.mu.Lock()
	defer s.collector.mu.Unlock()

	// Find the actual span in the trace and update it
	for i := range s.spanContext.trace.Spans {
		if s.spanContext.trace.Spans[i].ID == s.spanContext.span.ID {
			s.spanContext.trace.Spans[i].Attributes[key] = value

			// Special handling for certain attributes
			switch key {
			case "input", "prompt":
				if str, ok := value.(string); ok {
					s.spanContext.trace.Spans[i].Input = str
				}
			case "output", "response", "completion":
				if str, ok := value.(string); ok {
					s.spanContext.trace.Spans[i].Output = str
				}
			case "error":
				s.spanContext.trace.Status = "error"
			}
			break
		}
	}

	s.collector.updateTraceSize(s.spanContext.trace)
}

// RecordError records an error on the span
func (s *uiCollectorSpan) RecordError(err error) {
	if s.spanContext.wrappedSpan != nil {
		s.spanContext.wrappedSpan.RecordError(err)
	}

	if err == nil {
		return
	}

	s.collector.mu.Lock()
	defer s.collector.mu.Unlock()

	// Find the actual span in the trace and update it
	for i := range s.spanContext.trace.Spans {
		if s.spanContext.trace.Spans[i].ID == s.spanContext.span.ID {
			s.spanContext.trace.Spans[i].Error = &TraceError{
				Message:   err.Error(),
				Type:      fmt.Sprintf("%T", err),
				Timestamp: time.Now(),
			}
			// Update trace status to error
			s.spanContext.trace.Status = "error"
			break
		}
	}

	s.collector.updateTraceSize(s.spanContext.trace)
}
