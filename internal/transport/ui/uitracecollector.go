package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// NewTraceCollector creates a trace collector for the UI.
func NewTraceCollector(config *TracingConfig, wrappedTracer contracts.Tracer, logger telemetry.Logger) *TraceCollector {
	if config == nil {
		config = &TracingConfig{
			Enabled:         true,
			MaxBufferSizeKB: 10240, // 10MB
			MaxTraceAge:     "1h",
			RetentionCount:  100,
		}
	}

	// Parse max age duration
	maxAge, err := time.ParseDuration(config.MaxTraceAge)
	if err != nil {
		maxAge = time.Hour // Default to 1 hour
	}

	// Use default logger if none provided
	if logger == nil {
		logger = telemetry.NewLogger()
	}

	return &TraceCollector{
		config:        config,
		wrappedTracer: wrappedTracer,
		traces:        make(map[string]*Trace),
		activeSpans:   make(map[string]*uiSpanContext),
		maxSizeBytes:  config.MaxBufferSizeKB * 1024,
		maxAge:        maxAge,
		logger:        logger,
	}
}

// StartSpan starts a new span and collects it
func (c *TraceCollector) StartSpan(ctx context.Context, name string) (context.Context, contracts.Span) {
	c.logger.Debug(ctx, "StartSpan called", map[string]interface{}{
		"name": name,
	})

	if !c.config.Enabled {
		if c.wrappedTracer != nil {
			return c.wrappedTracer.StartSpan(ctx, name)
		}
		return ctx, &noOpSpan{}
	}

	// Start span in wrapped tracer if available
	var wrappedSpan contracts.Span
	if c.wrappedTracer != nil {
		ctx, wrappedSpan = c.wrappedTracer.StartSpan(ctx, name)
	}

	// Create UI span
	spanID := uuid.New().String()
	span := &TraceSpan{
		ID:         spanID,
		Name:       name,
		Type:       c.inferSpanType(name),
		StartTime:  time.Now(),
		Events:     []TraceEvent{},
		Attributes: make(map[string]interface{}),
	}

	// Find or create trace
	var trace *Trace

	// Try to get parent span from context
	if parentSpanID := c.getParentSpanID(ctx); parentSpanID != "" {
		c.mu.RLock()
		if parentContext, exists := c.activeSpans[parentSpanID]; exists {
			trace = parentContext.trace
			span.TraceID = trace.ID
			span.ParentID = parentSpanID
		}
		c.mu.RUnlock()
	}

	// If no parent found, create new trace
	if trace == nil {
		traceID := uuid.New().String()
		trace = &Trace{
			ID:        traceID,
			Name:      name,
			StartTime: time.Now(),
			Status:    "running",
			Spans:     []TraceSpan{},
			Metadata:  make(map[string]interface{}),
		}
		span.TraceID = traceID

		// Extract context metadata
		if conversationID := c.getConversationID(ctx); conversationID != "" {
			trace.ConversationID = conversationID
		}
		if orgID := c.getOrgID(ctx); orgID != "" {
			trace.OrgID = orgID
		}

		c.mu.Lock()
		c.traces[traceID] = trace
		c.enforceRetentionLimits()
		c.mu.Unlock()
	}

	// Store span context
	spanContext := &uiSpanContext{
		span:        span,
		trace:       trace,
		wrappedSpan: wrappedSpan,
	}

	c.mu.Lock()
	c.activeSpans[spanID] = spanContext
	trace.Spans = append(trace.Spans, *span)
	c.updateTraceSize(trace)
	c.mu.Unlock()

	// Store span ID in context for child spans
	ctx = context.WithValue(ctx, spanIDKey{}, spanID)

	return ctx, &uiCollectorSpan{
		collector:   c,
		spanContext: spanContext,
	}
}

// StartTraceSession starts a root trace session
func (c *TraceCollector) StartTraceSession(ctx context.Context, contextID string) (context.Context, contracts.Span) {
	c.logger.Debug(ctx, "StartTraceSession called", map[string]interface{}{
		"context_id": contextID,
	})

	if !c.config.Enabled {
		c.logger.Debug(ctx, "Tracing disabled, delegating to wrapped tracer", nil)
		if c.wrappedTracer != nil {
			return c.wrappedTracer.StartTraceSession(ctx, contextID)
		}
		return ctx, &noOpSpan{}
	}

	// Create a root span with the session name
	sessionName := fmt.Sprintf("session:%s", contextID)
	c.logger.Debug(ctx, "Creating root span", map[string]interface{}{
		"session_name": sessionName,
	})
	ctx, span := c.StartSpan(ctx, sessionName)

	// Add session metadata
	span.SetAttribute("session_id", contextID)
	span.SetAttribute("is_root", true)

	c.logger.Debug(ctx, "Root trace session started successfully", nil)
	return ctx, span
}
