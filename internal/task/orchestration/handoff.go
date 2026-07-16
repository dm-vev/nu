package orchestration

// HandoffRequest represents a request to hand off to another agent
type HandoffRequest struct {
	// TargetAgentID is the ID of the agent to hand off to
	TargetAgentID string

	// Reason explains why the handoff is happening
	Reason string

	// Context contains additional context for the target agent
	Context map[string]interface{}

	// Query is the query to send to the target agent
	Query string

	// PreserveMemory indicates whether to copy memory to the target agent
	PreserveMemory bool
}

// HandoffResult represents the result of a handoff
type HandoffResult struct {
	// AgentID is the ID of the agent that handled the request
	AgentID string

	// Response is the response from the agent
	Response string

	// Completed indicates whether the task was completed
	Completed bool

	// NextHandoff is the next handoff request, if any
	NextHandoff *HandoffRequest
}
