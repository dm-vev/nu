package client

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const TraceParentAttribute = "traceparent"

func tracingMiddleware(h mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (result mcp.Result, err error) {
		if method == "tools/call" {
			spanCtx := trace.SpanContextFromContext(ctx)
			if !spanCtx.IsValid() {
				return h(ctx, method, req)
			}
			propagator := propagation.TraceContext{}
			headers := make(http.Header)
			propagator.Inject(ctx, propagation.HeaderCarrier(headers))
			traceparentValue := headers.Get(TraceParentAttribute)
			if rp, ok := req.GetParams().(mcp.RequestParams); ok {
				if rp.GetMeta() == nil {
					rp.SetMeta(map[string]any{TraceParentAttribute: traceparentValue})
				} else {
					rp.GetMeta()[TraceParentAttribute] = traceparentValue
				}
			}
		}
		return h(ctx, method, req)
	}
}
