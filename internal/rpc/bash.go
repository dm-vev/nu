package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"nu/internal/tools/coding"
)

const defaultBashOutputLimit = 16 * 1024

func (s *Server) runBash(ctx context.Context, command string) (any, error) {
	args, err := json.Marshal(map[string]any{"command": command})
	if err != nil {
		return nil, fmt.Errorf("encode bash command: %w", err)
	}
	result, err := coding.RunBash(ctx, s.cwd, string(args), defaultBashOutputLimit)
	if err != nil {
		return nil, err
	}
	var data any
	if err := json.Unmarshal([]byte(result.Content), &data); err != nil {
		return result.Content, nil
	}
	return data, nil
}
