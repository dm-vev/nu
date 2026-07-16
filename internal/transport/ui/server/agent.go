package server

import (
	"fmt"
	"strings"

	"nu/internal/agent"
)

// getSubAgentsList returns list of sub-agents
func (h *Server) getSubAgentsList() []SubAgentInfo {
	subAgents := []SubAgentInfo{}

	// Check if agent is remote
	if h.Agent.IsRemote() {
		// For remote agents, parse from system prompt
		systemPrompt := h.getSystemPrompt()
		toolNames := h.parseToolsFromSystemPrompt(systemPrompt)

		for _, toolName := range toolNames {
			if strings.HasSuffix(toolName, "_agent") {
				agentName := strings.TrimSuffix(toolName, "_agent")
				subAgent := SubAgentInfo{
					ID:           toolName,
					Name:         agentName,
					Description:  h.getToolDescriptionFromSystemPrompt(toolName, systemPrompt),
					Model:        "Remote",
					Status:       "active",
					Tools:        []string{toolName},
					Capabilities: []string{"Remote sub-agent"},
				}
				subAgents = append(subAgents, subAgent)
			}
		}
	} else {
		// Get sub-agents directly from the agent instance
		agentSubAgents := h.Agent.GetSubAgents()
		for _, subAgent := range agentSubAgents {
			subAgentInfo := SubAgentInfo{
				ID:           subAgent.GetName(),
				Name:         subAgent.GetName(),
				Description:  subAgent.GetDescription(),
				Model:        h.getSubAgentModel(subAgent),
				Status:       "active", // Sub-agents are active if they're registered
				Tools:        h.getSubAgentTools(subAgent),
				Capabilities: []string{"Sub-agent"},
			}
			subAgents = append(subAgents, subAgentInfo)
		}

		// Also check tools for sub-agent tools (tools that end with _agent)
		tools := h.Agent.GetTools()
		for _, tool := range tools {
			toolName := tool.Name()
			// Check if this tool represents a sub-agent (ends with _agent)
			if strings.HasSuffix(toolName, "_agent") {
				// Extract the agent name by removing _agent suffix
				agentName := strings.TrimSuffix(toolName, "_agent")

				// Check if we already have this sub-agent from GetSubAgents()
				found := false
				for _, existing := range subAgents {
					if existing.ID == toolName || existing.Name == agentName {
						found = true
						break
					}
				}

				if !found {
					subAgent := SubAgentInfo{
						ID:           toolName,
						Name:         agentName,
						Description:  tool.Description(),
						Model:        "Unknown",
						Status:       "active",
						Tools:        []string{toolName},
						Capabilities: []string{"Tool-based sub-agent"},
					}
					subAgents = append(subAgents, subAgent)
				}
			}
		}
	}

	return subAgents
}

// getSubAgentModel extracts model information from a sub-agent
func (h *Server) getSubAgentModel(subAgent *agent.Agent) string {
	if subAgent.IsRemote() {
		return "Remote Agent"
	}

	llm := subAgent.GetLLM()
	if llm == nil {
		return "No LLM"
	}

	// Try to get model from LLM if it supports GetModel method
	if modelGetter, ok := llm.(interface{ GetModel() string }); ok {
		model := modelGetter.GetModel()
		if model != "" {
			return model
		}
	}

	// Fallback to LLM name
	name := llm.Name()
	if name != "" {
		return name
	}

	return "Unknown"
}

// getSubAgentTools gets the tools available to a sub-agent
func (h *Server) getSubAgentTools(subAgent *agent.Agent) []string {
	tools := subAgent.GetTools()
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name())
	}
	return toolNames
}

// parseToolsFromSystemPrompt extracts tool names from system prompt for remote agents
func (h *Server) parseToolsFromSystemPrompt(systemPrompt string) []string {
	tools := []string{}

	// Look for common patterns in system prompt
	lines := strings.Split(systemPrompt, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for patterns like "### ToolName_agent" or "- **Usage**: `ToolName_agent`"
		if strings.Contains(line, "_agent") {
			// Extract tool names ending with _agent
			words := strings.Fields(line)
			for _, word := range words {
				// Clean up word (remove markdown, punctuation)
				word = strings.Trim(word, "#*`-:.,!?()[]{}\"'")
				if strings.HasSuffix(word, "_agent") {
					// Check if not already added
					found := false
					for _, existingTool := range tools {
						if existingTool == word {
							found = true
							break
						}
					}
					if !found {
						tools = append(tools, word)
					}
				}
			}
		}
	}

	return tools
}

// getToolDescriptionFromSystemPrompt extracts tool description from system prompt
func (h *Server) getToolDescriptionFromSystemPrompt(toolName, systemPrompt string) string {
	lines := strings.Split(systemPrompt, "\n")

	for i, line := range lines {
		if strings.Contains(line, toolName) {
			// Look for description in nearby lines
			for j := i; j < len(lines) && j < i+5; j++ {
				if strings.Contains(lines[j], "Purpose") && strings.Contains(lines[j], ":") {
					parts := strings.SplitN(lines[j], ":", 2)
					if len(parts) == 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
			// Fallback to generic description
			return fmt.Sprintf("%s sub-agent", strings.TrimSuffix(toolName, "_agent"))
		}
	}

	return "Sub-agent tool"
}
