package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	nethttp "net/http"
	"os"
	"strings"
	"time"

	"github.com/dm-vev/nu/agent"
	httpserver "github.com/dm-vev/nu/internal/transport/http/server"
	"github.com/dm-vev/nu/internal/transport/ui/trace"
)

// New creates an HTTP server with the embedded UI.
func New(agent *agent.Agent, port int, config *Config) *Server {
	if config == nil {
		config = &Config{
			Enabled:     true,
			DefaultPath: "/",
			DevMode:     false,
			Theme:       "light",
			Features: Features{
				Chat:      true,
				Memory:    true,
				AgentInfo: true,
				Settings:  true,
				Traces:    false, // Disabled by default
			},
		}
	}

	// Set default tracing config if traces are enabled
	if config.Features.Traces && config.Tracing == nil {
		config.Tracing = &trace.Config{
			Enabled:         true,
			MaxBufferSizeKB: 10240, // 10MB
			MaxTraceAge:     "1h",
			RetentionCount:  100,
		}
	}

	// Extract the embedded UI files
	var uiFS fs.FS
	var err error
	uiFS, err = fs.Sub(defaultUIFiles, "ui-nextjs/out")
	if err != nil {
		// Fallback to serving from local directory in dev mode
		if config.DevMode {
			uiFS = os.DirFS("./internal/transport/ui/server/ui-nextjs/out")
		}
	}

	server := &Server{
		Server: httpserver.Server{
			Agent: agent,
			Port:  port,
		},
		uiConfig:            config,
		uiFS:                uiFS,
		conversationHistory: make([]MemoryEntry, 0),
	}

	// Initialize trace collector if enabled
	if config.Features.Traces && config.Tracing != nil && config.Tracing.Enabled {
		// Check if agent already has a TraceCollector
		if agent.GetTracer() != nil {
			if uiCollector, ok := agent.GetTracer().(*trace.Collector); ok {
				// Agent already has a TraceCollector, use it
				server.traceCollector = uiCollector
				log.Printf("[UI Server] Using existing TraceCollector from agent")
			} else {
				// Agent has a different tracer, wrap it with new TraceCollector
				server.traceCollector = trace.New(config.Tracing, agent.GetTracer(), agent.GetLogger())
				log.Printf("[UI Server] Created new TraceCollector wrapping agent's tracer")
			}
		} else {
			// Agent has no tracer, create new TraceCollector
			server.traceCollector = trace.New(config.Tracing, nil, agent.GetLogger())
			log.Printf("[UI Server] Created new TraceCollector (agent has no tracer)")
		}
	}

	return server
}

// Start starts the HTTP server with UI
func (h *Server) Start() error {
	mux := nethttp.NewServeMux()

	// Add CORS middleware
	corsHandler := h.AddCORS(mux)

	// Register API endpoints
	h.registerAPIEndpoints(mux)

	// Debug endpoint to list embedded files
	mux.HandleFunc("/debug/files", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if h.uiFS != nil {
			var files []string
			err := fs.WalkDir(h.uiFS, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				files = append(files, path)
				return nil
			})
			if err != nil {
				nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(files)
		} else {
			nethttp.Error(w, "No UI filesystem", nethttp.StatusNotFound)
		}
	})

	// Serve UI if enabled
	if h.uiConfig.Enabled && h.uiFS != nil {
		// Serve the embedded UI files
		fileServer := nethttp.FileServer(nethttp.FS(h.uiFS))

		// Handle static assets specifically
		mux.Handle("/_next/", fileServer)
		mux.Handle("/favicon.ico", fileServer)

		// Handle root and everything else
		mux.HandleFunc("/", func(w nethttp.ResponseWriter, r *nethttp.Request) {
			// For non-API requests, serve the index.html
			if !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/health") {
				// Try to serve the file first
				if file, err := h.uiFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err == nil {
					_ = file.Close()
					fileServer.ServeHTTP(w, r)
					return
				}
				// Fallback to index.html for SPA routing
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	h.Server.Server = &nethttp.Server{
		Addr:         fmt.Sprintf(":%d", h.Port),
		Handler:      corsHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 15 * time.Minute, // Longer timeout for streaming
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("HTTP server with UI starting on port %d\n", h.Port)
	if h.uiConfig.Enabled {
		fmt.Printf("UI available at: http://localhost:%d%s\n", h.Port, h.uiConfig.DefaultPath)
	}

	fmt.Printf("API endpoints available:\n")
	fmt.Printf("  - POST /api/v1/agent/run (non-streaming)\n")
	fmt.Printf("  - POST /api/v1/agent/stream (SSE streaming)\n")
	fmt.Printf("  - GET /api/v1/agent/metadata\n")
	fmt.Printf("  - GET /health\n")

	if h.uiConfig.Enabled {
		fmt.Printf("UI-specific endpoints:\n")
		fmt.Printf("  - GET /api/v1/agent/config\n")
		fmt.Printf("  - GET /api/v1/agent/subagents\n")
		fmt.Printf("  - POST /api/v1/agent/delegate\n")
		fmt.Printf("  - GET /api/v1/memory\n")
		fmt.Printf("  - GET /api/v1/memory/search\n")
		fmt.Printf("  - GET /api/v1/tools\n")

		if h.uiConfig.Features.Traces && h.traceCollector != nil {
			fmt.Printf("Trace endpoints:\n")
			fmt.Printf("  - GET /api/v1/traces\n")
			fmt.Printf("  - GET /api/v1/traces/{id}\n")
			fmt.Printf("  - DELETE /api/v1/traces/{id}\n")
			fmt.Printf("  - GET /api/v1/traces/stats\n")
		}
	}

	return h.Server.Server.ListenAndServe()
}

// registerAPIEndpoints registers all API endpoints
func (h *Server) registerAPIEndpoints(mux *nethttp.ServeMux) {
	// Health check (always available)
	mux.HandleFunc("/health", h.HandleHealth)

	// Core agent endpoints (always available)
	mux.HandleFunc("/api/v1/agent/run", h.withOrgContext(h.handleRun))
	mux.HandleFunc("/api/v1/agent/stream", h.withOrgContext(h.handleStream))
	mux.HandleFunc("/api/v1/agent/metadata", h.HandleMetadata)

	// UI-specific endpoints (only when UI is enabled)
	if h.uiConfig.Enabled {
		mux.HandleFunc("/api/v1/agent/config", h.handleConfig)
		mux.HandleFunc("/api/v1/agent/subagents", h.handleSubAgents)
		mux.HandleFunc("/api/v1/agent/delegate", h.withOrgContext(h.handleDelegate))
		mux.HandleFunc("/api/v1/memory", h.withOrgContext(h.handleMemory))
		mux.HandleFunc("/api/v1/memory/search", h.withOrgContext(h.handleMemorySearch))
		mux.HandleFunc("/api/v1/tools", h.handleTools)
		mux.HandleFunc("/ws/chat", h.handleWebSocketChat)

		// Trace endpoints (only when traces feature is enabled)
		if h.uiConfig.Features.Traces && h.traceCollector != nil {
			mux.HandleFunc("/api/v1/traces", h.handleTraces)
			mux.HandleFunc("/api/v1/traces/stats", h.handleTraceStats)
			// Pattern matching for /api/v1/traces/{id}
			mux.HandleFunc("/api/v1/traces/", h.handleTrace)
		}
	}
}
