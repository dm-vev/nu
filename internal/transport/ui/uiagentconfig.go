package ui

import (
	"context"
	"fmt"
	nethttp "net/http"
	"strings"

	"nu/internal/contracts"
	"nu/internal/multitenancy"
)

// withOrgContext adds organization context to HTTP requests
func (h *Server) withOrgContext(handler nethttp.HandlerFunc) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		ctx := r.Context()

		// Check if organization ID is already in context
		if !multitenancy.HasOrgID(ctx) {
			// Add default organization ID
			ctx = multitenancy.WithOrgID(ctx, "default-org")
			r = r.WithContext(ctx)
		}

		handler(w, r)
	}
}

// getToolNames extracts tool names from the agent
func (h *Server) getToolNames() []string {
	tools := h.Agent.GetTools()
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name())
	}
	return toolNames
}

// getModelName extracts the model name from the agent's LLM
func (h *Server) getModelName() string {
	// For remote agents, try to get LLM info from metadata
	if h.Agent.IsRemote() {
		if metadata, err := h.Agent.GetRemoteMetadata(); err == nil && metadata != nil {
			if llmModel, ok := metadata["llm_model"]; ok && llmModel != "" && llmModel != "unknown" {
				return llmModel
			}
			if llmName, ok := metadata["llm_name"]; ok && llmName != "" && llmName != "unknown" {
				return llmName + " (model not specified)"
			}
		}
		return "Remote agent - metadata unavailable"
	}

	// For local agents, get from LLM directly
	llm := h.Agent.GetLLM()
	if llm == nil {
		return "No LLM configured"
	}

	// Try to get model from LLM if it supports GetModel method
	if modelGetter, ok := llm.(interface{ GetModel() string }); ok {
		model := modelGetter.GetModel()
		if model != "" {
			// Special handling for Azure OpenAI deployments
			if llm.Name() == "azure-openai" {
				// Try to extract model name from deployment name
				if inferredModel := inferAzureModelFromDeployment(model); inferredModel != "" {
					return inferredModel + " (deployment: " + model + ")"
				}
				return "Azure OpenAI (deployment: " + model + ")"
			}
			return model
		}
	}

	// Fallback to LLM name if GetModel is not available or returns empty
	name := llm.Name()
	if name != "" {
		return name + " (model not specified)"
	}

	return "Unknown LLM"
}

// inferAzureModelFromDeployment attempts to infer the actual model name from Azure deployment name
func inferAzureModelFromDeployment(deployment string) string {
	deployment = strings.ToLower(deployment)

	// Common Azure OpenAI model patterns
	if strings.Contains(deployment, "gpt-4o") {
		if strings.Contains(deployment, "mini") {
			return "gpt-4o-mini"
		}
		return "gpt-4o"
	}
	if strings.Contains(deployment, "gpt-4-turbo") || strings.Contains(deployment, "gpt4-turbo") {
		return "gpt-4-turbo"
	}
	if strings.Contains(deployment, "gpt-4") || strings.Contains(deployment, "gpt4") {
		return "gpt-4"
	}
	if strings.Contains(deployment, "gpt-35-turbo") || strings.Contains(deployment, "gpt-3.5-turbo") {
		return "gpt-3.5-turbo"
	}
	if strings.Contains(deployment, "o1-preview") {
		return "o1-preview"
	}
	if strings.Contains(deployment, "o1-mini") {
		return "o1-mini"
	}
	if strings.Contains(deployment, "text-embedding") {
		return "text-embedding-ada-002"
	}
	if strings.Contains(deployment, "dall-e") || strings.Contains(deployment, "dalle") {
		return "dall-e-3"
	}

	// If no pattern matches, return empty string
	return ""
}

// getMemoryInfo extracts memory information from the agent
func (h *Server) getMemoryInfo() MemoryInfo {
	// For remote agents, try to get memory info from metadata
	if h.Agent.IsRemote() {
		if metadata, err := h.Agent.GetRemoteMetadata(); err == nil && metadata != nil {
			if memoryType, ok := metadata["memory"]; ok && memoryType != "" && memoryType != "none" {
				return MemoryInfo{
					Type:   memoryType,
					Status: "active",
					// Entry count not available from remote metadata yet
				}
			}
		}
		return MemoryInfo{
			Type:   "none",
			Status: "inactive",
		}
	}

	// For local agents, check memory directly
	mem := h.Agent.GetMemory()
	if mem == nil {
		// Check if there's a memory config that indicates the type
		// even if the instance hasn't been created yet
		if memConfig := h.Agent.GetMemoryConfig(); memConfig != nil {
			if memType, ok := memConfig["type"].(string); ok && memType != "" {
				return MemoryInfo{
					Type:   memType,
					Status: "configured", // Memory is configured but not instantiated
				}
			}
		}
		return MemoryInfo{
			Type:   "none",
			Status: "inactive",
		}
	}

	// Determine memory type by checking the concrete type
	memType := h.detectMemoryType(mem)

	memInfo := MemoryInfo{
		Type:   memType,
		Status: "active",
	}

	// Try to get entry count if the memory supports it
	ctx := context.Background()
	if messages, err := mem.GetMessages(ctx); err == nil {
		memInfo.EntryCount = len(messages)
	}

	return memInfo
}

// detectMemoryType determines the actual type of memory implementation
func (h *Server) detectMemoryType(mem contracts.Memory) string {
	// Check for specific memory types using type assertions
	// We use a type switch approach with interface checks

	// Check for RedisMemory by looking for Close method (specific to Redis)
	if _, ok := mem.(interface{ Close() error }); ok {
		return "redis"
	}

	// Check for ConversationSummary by looking for specific behavior
	// ConversationSummary wraps a buffer and has summarization
	memType := fmt.Sprintf("%T", mem)

	switch {
	case strings.Contains(memType, "RedisMemory"):
		return "redis"
	case strings.Contains(memType, "ConversationSummary"):
		return "buffer_summary"
	case strings.Contains(memType, "ConversationBuffer"):
		return "buffer"
	case strings.Contains(memType, "TracedMemory"):
		return "traced"
	default:
		// Fallback: if it implements AdminConversationMemory, it's likely redis or buffer
		if _, ok := mem.(contracts.AdminConversationMemory); ok {
			return "conversation"
		}
		return "memory"
	}
}

// getDataStoreInfo extracts datastore information from the agent
func (h *Server) getDataStoreInfo() DataStoreInfo {
	// For remote agents, try to get datastore info from metadata
	if h.Agent.IsRemote() {
		if metadata, err := h.Agent.GetRemoteMetadata(); err == nil && metadata != nil {
			if dsType, ok := metadata["datastore"]; ok && dsType != "" && dsType != "none" {
				return DataStoreInfo{
					Type:   dsType,
					Status: "active",
				}
			}
		}
		return DataStoreInfo{
			Type:   "none",
			Status: "inactive",
		}
	}

	// For local agents, check datastore directly
	ds := h.Agent.GetDataStore()
	if ds == nil {
		return DataStoreInfo{
			Type:   "none",
			Status: "inactive",
		}
	}

	// Determine datastore type by checking the concrete type
	dsType := h.detectDataStoreType(ds)

	return DataStoreInfo{
		Type:   dsType,
		Status: "active",
	}
}

// detectDataStoreType determines the actual type of datastore implementation
func (h *Server) detectDataStoreType(ds contracts.DataStore) string {
	// Use type name to determine the datastore type
	dsType := fmt.Sprintf("%T", ds)

	switch {
	case strings.Contains(dsType, "sql.PostgresClient"):
		return "postgres"
	case strings.Contains(dsType, "sql.SupabaseClient"):
		return "supabase"
	default:
		return "database"
	}
}

// getSystemPrompt gets system prompt, handling remote agents
func (h *Server) getSystemPrompt() string {
	// For remote agents, try to get from metadata
	if h.Agent.IsRemote() {
		if metadata, err := h.Agent.GetRemoteMetadata(); err == nil && metadata != nil {
			if systemPrompt, ok := metadata["system_prompt"]; ok && systemPrompt != "" {
				return systemPrompt
			}
		}
		return "Remote agent - system prompt unavailable"
	}

	// For local agents, get directly
	systemPrompt := h.Agent.GetSystemPrompt()
	if systemPrompt == "" {
		systemPrompt = "No system prompt configured"
	}
	return systemPrompt
}
