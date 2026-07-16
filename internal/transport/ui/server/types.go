package server

import (
	"embed"
	"io/fs"

	httpserver "nu/internal/transport/http/server"
	"nu/internal/transport/ui/trace"
)

// Config represents UI configuration options
type Config struct {
	Enabled     bool          `json:"enabled"`
	DefaultPath string        `json:"default_path"`
	DevMode     bool          `json:"dev_mode"`
	Theme       string        `json:"theme"`
	Features    Features      `json:"features"`
	Tracing     *trace.Config `json:"tracing,omitempty"`
}

// Features represents available UI features
type Features struct {
	Chat      bool `json:"chat"`
	Memory    bool `json:"memory"`
	AgentInfo bool `json:"agent_info"`
	Settings  bool `json:"settings"`
	Traces    bool `json:"traces"`
}

// Server extends the HTTP transport with the embedded UI.
type Server struct {
	httpserver.Server
	uiConfig *Config
	uiFS     fs.FS

	// Simple in-memory conversation storage
	conversationHistory []MemoryEntry

	// Trace collector for UI
	traceCollector *trace.Collector
}

// SubAgentInfo represents sub-agent information for UI.
type SubAgentInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Model        string   `json:"model"`
	Status       string   `json:"status"`
	Tools        []string `json:"tools"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// AgentConfigResponse represents detailed agent configuration.
type AgentConfigResponse struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Model        string                 `json:"model"`
	SystemPrompt string                 `json:"system_prompt"`
	Tools        []string               `json:"tools"`
	Memory       MemoryInfo             `json:"memory"`
	DataStore    DataStoreInfo          `json:"datastore"`
	SubAgents    []SubAgentInfo         `json:"sub_agents,omitempty"`
	Features     Features               `json:"features"`
	UITheme      string                 `json:"ui_theme,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryInfo represents memory system information.
type MemoryInfo struct {
	Type        string `json:"type"`
	Status      string `json:"status"`
	EntryCount  int    `json:"entry_count,omitempty"`
	MaxCapacity int    `json:"max_capacity,omitempty"`
}

// DataStoreInfo represents datastore/database connection information.
type DataStoreInfo struct {
	Type   string `json:"type"`   // "postgres", "supabase", "none"
	Status string `json:"status"` // "active", "inactive"
}

// MemoryEntry represents a memory entry for the browser.
type MemoryEntry struct {
	ID             string                 `json:"id"`
	Role           string                 `json:"role"`
	Content        string                 `json:"content"`
	Timestamp      int64                  `json:"timestamp"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationInfo represents conversation metadata.
type ConversationInfo struct {
	ID           string `json:"id"`
	MessageCount int    `json:"message_count"`
	LastActivity int64  `json:"last_activity"`
	LastMessage  string `json:"last_message,omitempty"`
}

// MemoryResponse represents the response structure for memory endpoints.
type MemoryResponse struct {
	Mode           string             `json:"mode"` // "conversations" or "messages"
	Conversations  []ConversationInfo `json:"conversations,omitempty"`
	Messages       []MemoryEntry      `json:"messages,omitempty"`
	Total          int                `json:"total"`
	Limit          int                `json:"limit"`
	Offset         int                `json:"offset"`
	ConversationID string             `json:"conversation_id,omitempty"`
}

// DelegateRequest represents a request to delegate to a sub-agent.
type DelegateRequest struct {
	SubAgentID     string            `json:"sub_agent_id"`
	Task           string            `json:"task"`
	Context        map[string]string `json:"context,omitempty"`
	ConversationID string            `json:"conversation_id,omitempty"`
}

// Embed UI files (will be populated at build time)
//
//go:embed all:ui-nextjs/out
var defaultUIFiles embed.FS

// GetTraceCollector returns the UI trace collector if enabled
func (h *Server) GetTraceCollector() *trace.Collector {
	return h.traceCollector
}
