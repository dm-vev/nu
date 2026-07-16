package server

import (
	"net/http"
	"time"

	"github.com/dm-vev/nu/telemetry"
)

// Option configures an A2A server.
type Option func(*Server)

// WithAddress sets the listen address for the A2A server.
func WithAddress(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithBasePath sets the JSON-RPC endpoint base path.
func WithBasePath(path string) Option {
	return func(s *Server) {
		s.basePath = path
	}
}

// WithLogger sets a logger for the A2A server.
func WithLogger(logger telemetry.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithShutdownTimeout sets the graceful shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = d
	}
}

// WithReadHeaderTimeout sets the maximum time for reading request headers.
func WithReadHeaderTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.readHeaderTimeout = d
	}
}

// WithReadTimeout sets the maximum time for reading a request.
func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = d
	}
}

// WithWriteTimeout sets the maximum time for writing a response.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = d
	}
}

// WithIdleTimeout sets the keep-alive idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.idleTimeout = d
	}
}

// WithMiddleware adds an HTTP middleware to the A2A server.
func WithMiddleware(middleware func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, middleware)
	}
}
