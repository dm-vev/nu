package ui

import (
	"encoding/json"
	nethttp "net/http"
	"strconv"
	"strings"
)

// handleTraces handles GET /api/v1/traces endpoint
func (h *Server) handleTraces(w nethttp.ResponseWriter, r *nethttp.Request) {
	switch r.Method {
	case "GET":
		// Get traces list with pagination
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		offset := 0
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		traces, total := h.traceCollector.GetTraces(limit, offset)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"traces": traces,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		}); err != nil {
			nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
		}

	default:
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
	}
}

// handleTrace handles GET/DELETE /api/v1/traces/{id} endpoint
func (h *Server) handleTrace(w nethttp.ResponseWriter, r *nethttp.Request) {
	// Extract trace ID from path
	path := r.URL.Path
	prefix := "/api/v1/traces/"
	if !strings.HasPrefix(path, prefix) {
		nethttp.Error(w, "Invalid path", nethttp.StatusBadRequest)
		return
	}

	traceID := strings.TrimPrefix(path, prefix)
	if traceID == "" || traceID == "stats" { // Skip stats endpoint
		nethttp.Error(w, "Trace ID required", nethttp.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		// Get specific trace
		trace, err := h.traceCollector.GetTrace(traceID)
		if err != nil {
			nethttp.Error(w, err.Error(), nethttp.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(trace); err != nil {
			nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
		}

	case "DELETE":
		// Delete specific trace
		if err := h.traceCollector.DeleteTrace(traceID); err != nil {
			nethttp.Error(w, err.Error(), nethttp.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "deleted",
			"id":     traceID,
		}); err != nil {
			nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
		}

	default:
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
	}
}

// handleTraceStats handles GET /api/v1/traces/stats endpoint
func (h *Server) handleTraceStats(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "GET" {
		nethttp.Error(w, "Method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	stats := h.traceCollector.GetStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		nethttp.Error(w, "Failed to encode response", nethttp.StatusInternalServerError)
	}
}
