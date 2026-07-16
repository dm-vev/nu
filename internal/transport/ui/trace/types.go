package trace

import (
	"sync"
	"time"

	"nu/internal/contracts"
	"nu/internal/telemetry"
)

// Trace represents a trace in the UI
type Trace struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        *time.Time             `json:"end_time,omitempty"`
	Duration       int64                  `json:"duration_ms"`
	Status         string                 `json:"status"` // running, completed, error
	Spans          []TraceSpan            `json:"spans"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	OrgID          string                 `json:"org_id,omitempty"`
	SizeBytes      int                    `json:"size_bytes"`
}

// TraceSpan represents a span in a trace
type TraceSpan struct {
	ID         string                 `json:"id"`
	TraceID    string                 `json:"trace_id"`
	ParentID   string                 `json:"parent_id,omitempty"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // generation, tool_call, span, event
	StartTime  time.Time              `json:"start_time"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Duration   int64                  `json:"duration_ms"`
	Events     []TraceEvent           `json:"events,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Error      *TraceError            `json:"error,omitempty"`
	Input      string                 `json:"input,omitempty"`
	Output     string                 `json:"output,omitempty"`
}

// TraceEvent represents an event in a span
type TraceEvent struct {
	Name       string                 `json:"name"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// TraceError represents an error in a span
type TraceError struct {
	Message    string    `json:"message"`
	Type       string    `json:"type,omitempty"`
	Stacktrace string    `json:"stacktrace,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// Config contains configuration for UI tracing
type Config struct {
	Enabled         bool   `json:"enabled"`
	MaxBufferSizeKB int    `json:"max_buffer_size_kb"` // Default: 10240 (10MB)
	MaxTraceAge     string `json:"max_trace_age"`      // Default: "1h"
	RetentionCount  int    `json:"retention_count"`    // Default: 100 traces
}

// Collector collects traces for the UI
type Collector struct {
	config           *Config
	wrappedTracer    contracts.Tracer
	traces           map[string]*Trace
	activeSpans      map[string]*uiSpanContext
	mu               sync.RWMutex
	maxSizeBytes     int
	currentSizeBytes int
	maxAge           time.Duration
	logger           telemetry.Logger
}

// uiSpanContext holds context for an active span
type uiSpanContext struct {
	span        *TraceSpan
	trace       *Trace
	wrappedSpan contracts.Span
}

// uiCollectorSpan wraps a span and collects data
type uiCollectorSpan struct {
	collector   *Collector
	spanContext *uiSpanContext
}
