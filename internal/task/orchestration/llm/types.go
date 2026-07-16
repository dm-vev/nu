package llm

// Step represents a single step in an orchestration plan
type Step struct {
	// AgentID is the ID of the agent to execute
	AgentID string `json:"agent_id"`

	// Input is the input to provide to the agent
	Input string `json:"input"`

	// Description explains the purpose of this step
	Description string `json:"description"`

	// DependsOn lists the IDs of steps that must complete before this one
	DependsOn []string `json:"depends_on,omitempty"`
}

// Plan represents an orchestration plan
type Plan struct {
	// Steps is the list of steps in the plan
	Steps []Step `json:"steps"`

	// FinalAgentID is the ID of the agent that should provide the final response
	FinalAgentID string `json:"final_agent_id"`
}
