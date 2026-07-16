package orchestration

import "context"

// Router determines which agent should handle a request
type HandoffRouter interface {
	Route(ctx context.Context, query string, context map[string]interface{}) (string, error)
}
