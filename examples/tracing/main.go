package main

import (
	"context"
	"log"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"nu/internal/telemetry"
	"nu/internal/telemetry/otel"
)

func main() {
	ctx := context.Background()
	provider := sdktrace.NewTracerProvider()
	tracer := otel.NewTracerWrapper(provider.Tracer("nu/examples/tracing"))

	ctx, request := telemetry.StartRequestTracing(ctx, tracer, "request-123")
	_, child := tracer.StartSpan(ctx, "process")
	child.SetAttribute("example.item_count", 3)
	child.AddEvent("items processed", map[string]interface{}{"count": 3})
	child.End()
	request.End()

	if err := provider.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
