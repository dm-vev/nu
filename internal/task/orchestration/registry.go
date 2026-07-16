package orchestration

import "nu/internal/agent"

// AgentRegistry maintains a registry of available agents
type OrchestratorAgentRegistry struct {
	agents map[string]*agent.Agent
}

// NewAgentRegistry creates a new agent registry
func NewOrchestratorAgentRegistry() *OrchestratorAgentRegistry {
	return &OrchestratorAgentRegistry{
		agents: make(map[string]*agent.Agent),
	}
}

// Register registers an agent with the registry
func (r *OrchestratorAgentRegistry) Register(id string, agent *agent.Agent) {
	r.agents[id] = agent
}

// Get retrieves an agent from the registry
func (r *OrchestratorAgentRegistry) Get(id string) (*agent.Agent, bool) {
	agent, ok := r.agents[id]
	return agent, ok
}

// List returns all registered agents
func (r *OrchestratorAgentRegistry) List() map[string]*agent.Agent {
	return r.agents
}
