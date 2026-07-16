package testkit

import (
	"context"

	"github.com/stretchr/testify/mock"

	"nu/internal/contracts"
)

// Server is a configurable MCP server double for domain package tests.
type Server struct {
	mock.Mock
}

func (s *Server) Initialize(ctx context.Context) error {
	return s.Called(ctx).Error(0)
}

func (s *Server) ListTools(ctx context.Context) ([]contracts.MCPTool, error) {
	args := s.Called(ctx)
	if value := args.Get(0); value != nil {
		return value.([]contracts.MCPTool), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) CallTool(ctx context.Context, name string, toolArgs interface{}) (*contracts.MCPToolResponse, error) {
	args := s.Called(ctx, name, toolArgs)
	if value := args.Get(0); value != nil {
		return value.(*contracts.MCPToolResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) ListResources(ctx context.Context) ([]contracts.MCPResource, error) {
	args := s.Called(ctx)
	if value := args.Get(0); value != nil {
		return value.([]contracts.MCPResource), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) GetResource(ctx context.Context, uri string) (*contracts.MCPResourceContent, error) {
	args := s.Called(ctx, uri)
	if value := args.Get(0); value != nil {
		return value.(*contracts.MCPResourceContent), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) WatchResource(ctx context.Context, uri string) (<-chan contracts.MCPResourceUpdate, error) {
	args := s.Called(ctx, uri)
	if value := args.Get(0); value != nil {
		return value.(<-chan contracts.MCPResourceUpdate), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) ListPrompts(ctx context.Context) ([]contracts.MCPPrompt, error) {
	args := s.Called(ctx)
	if value := args.Get(0); value != nil {
		return value.([]contracts.MCPPrompt), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) GetPrompt(ctx context.Context, name string, variables map[string]interface{}) (*contracts.MCPPromptResult, error) {
	args := s.Called(ctx, name, variables)
	if value := args.Get(0); value != nil {
		return value.(*contracts.MCPPromptResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) CreateMessage(ctx context.Context, request *contracts.MCPSamplingRequest) (*contracts.MCPSamplingResponse, error) {
	args := s.Called(ctx, request)
	if value := args.Get(0); value != nil {
		return value.(*contracts.MCPSamplingResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) GetServerInfo() (*contracts.MCPServerInfo, error) {
	args := s.Called()
	if value := args.Get(0); value != nil {
		return value.(*contracts.MCPServerInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) GetCapabilities() (*contracts.MCPServerCapabilities, error) {
	args := s.Called()
	if value := args.Get(0); value != nil {
		return value.(*contracts.MCPServerCapabilities), args.Error(1)
	}
	return nil, args.Error(1)
}

func (s *Server) Close() error {
	return s.Called().Error(0)
}
