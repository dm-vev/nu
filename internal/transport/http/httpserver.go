package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nu/internal/agent"
)

// NewServer creates a new HTTP server for agent streaming
func NewServer(agent *agent.Agent, port int) *Server {
	return &Server{
		Agent: agent,
		Port:  port,
	}
}

// Start starts the HTTP server
func (h *Server) Start() error {
	mux := http.NewServeMux()

	// Add CORS middleware
	corsHandler := h.AddCORS(mux)

	// Register endpoints
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/api/v1/agent/run", h.HandleRun)
	mux.HandleFunc("/api/v1/agent/stream", h.HandleStream)
	mux.HandleFunc("/api/v1/agent/metadata", h.HandleMetadata)

	// Serve static files for browser example (if they exist)
	mux.Handle("/", http.FileServer(http.Dir("./web/")))

	h.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.Port),
		Handler:      corsHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 15 * time.Minute, // Longer timeout for streaming
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("HTTP server starting on port %d\n", h.Port)
	fmt.Printf("Endpoints available:\n")
	fmt.Printf("  - POST /api/v1/agent/run (non-streaming)\n")
	fmt.Printf("  - POST /api/v1/agent/stream (SSE streaming)\n")
	fmt.Printf("  - GET /api/v1/agent/metadata\n")
	fmt.Printf("  - GET /health\n")

	return h.Server.ListenAndServe()
}

// Stop stops the HTTP server
func (h *Server) Stop(ctx context.Context) error {
	if h.Server != nil {
		return h.Server.Shutdown(ctx)
	}
	return nil
}

// addCORS adds CORS headers to allow browser access
func (h *Server) AddCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
