package langfuse

import (
	"context"
	"fmt"
	"time"

	"nu/internal/config"
	"nu/internal/contracts"
	"nu/internal/telemetry/otel"
)

// Tracer implements tracing using Langfuse via OTEL (backward compatibility wrapper)
// This replaces the old buggy henomis/langfuse-go implementation with our reliable OTEL-based one
type Tracer struct {
	otelTracer *OTELTracer
	enabled    bool
}

// Config contains configuration for Langfuse
type Config struct {
	// Enabled determines whether Langfuse tracing is enabled
	Enabled bool

	// SecretKey is the Langfuse secret key
	SecretKey string

	// PublicKey is the Langfuse public key
	PublicKey string

	// Host is the Langfuse host (optional)
	Host string

	// Environment is the environment name (e.g., "production", "staging")
	Environment string

	// IncludeContent determines whether actual prompt/response content is included in traces
	// When false (default), only hashes are stored for privacy
	IncludeContent bool
}

// NewTracer creates a new Langfuse tracer (backward compatibility wrapper)
// This now uses the reliable OTEL-based implementation internally
func NewTracer(customConfig ...Config) (*Tracer, error) {
	// Get global configuration
	cfg := config.Get()

	// Use custom config if provided, otherwise use global config
	var tracerConfig Config
	if len(customConfig) > 0 {
		tracerConfig = customConfig[0]
	} else {
		tracerConfig = Config{
			Enabled:        cfg.Tracing.Langfuse.Enabled,
			SecretKey:      cfg.Tracing.Langfuse.SecretKey,
			PublicKey:      cfg.Tracing.Langfuse.PublicKey,
			Host:           cfg.Tracing.Langfuse.Host,
			Environment:    cfg.Tracing.Langfuse.Environment,
			IncludeContent: cfg.Tracing.Langfuse.IncludeContent,
		}
	}

	if !tracerConfig.Enabled {
		return &Tracer{
			enabled: false,
		}, nil
	}

	// Create the new OTEL-based Langfuse tracer internally
	otelTracer, err := NewOTELTracer(tracerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTEL Langfuse tracer: %w", err)
	}

	return &Tracer{
		otelTracer: otelTracer,
		enabled:    true,
	}, nil
}

// TraceGeneration traces an LLM generation (delegates to OTEL implementation)
func (t *Tracer) TraceGeneration(ctx context.Context, modelName string, prompt string, response string, startTime time.Time, endTime time.Time, metadata map[string]interface{}) (string, error) {
	if !t.enabled || t.otelTracer == nil {
		return "", nil
	}
	return t.otelTracer.TraceGeneration(ctx, modelName, prompt, response, startTime, endTime, metadata)
}

// TraceSpan traces a span of execution (delegates to OTEL implementation)
func (t *Tracer) TraceSpan(ctx context.Context, name string, startTime time.Time, endTime time.Time, metadata map[string]interface{}, parentID string) (string, error) {
	if !t.enabled || t.otelTracer == nil {
		return "", nil
	}
	return t.otelTracer.TraceSpan(ctx, name, startTime, endTime, metadata, parentID)
}

// TraceEvent traces an event (delegates to OTEL implementation)
func (t *Tracer) TraceEvent(ctx context.Context, name string, input interface{}, output interface{}, level string, metadata map[string]interface{}, parentID string) (string, error) {
	if !t.enabled || t.otelTracer == nil {
		return "", nil
	}
	return t.otelTracer.TraceEvent(ctx, name, input, output, level, metadata, parentID)
}

// Flush flushes the Langfuse tracer (delegates to OTEL implementation)
func (t *Tracer) Flush() error {
	if !t.enabled || t.otelTracer == nil {
		return nil
	}
	return t.otelTracer.Flush()
}

// Shutdown shuts down the tracer (delegates to OTEL implementation)
func (t *Tracer) Shutdown() error {
	if !t.enabled || t.otelTracer == nil {
		return nil
	}
	return t.otelTracer.Shutdown()
}

// AsInterfaceTracer returns a contracts.Tracer compatible adapter
// This allows the backward-compatible tracer to work with Agents
func (t *Tracer) AsInterfaceTracer() contracts.Tracer {
	if !t.enabled || t.otelTracer == nil {
		return nil
	}
	return NewTracerAdapter(t.otelTracer)
}

// @deprecated Use NewTracedLLM - removing in v1.0.0
func NewLLMMiddleware(llm contracts.LLM, tracer *Tracer) contracts.LLM {
	return otel.NewTracedLLM(llm, tracer.AsInterfaceTracer())
}

// @deprecated Use NewTracedLLM - removing in v1.0.0
func NewOTELMiddleware(llm contracts.LLM, tracer *OTELTracer) contracts.LLM {
	return otel.NewTracedLLM(llm, tracer)
}
