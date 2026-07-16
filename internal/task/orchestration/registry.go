package orchestration

import "github.com/dm-vev/nu/agent"

// AgentRegistry maintains a registry of available agents
type AgentRegistry struct {
	agents map[string]*agent.Agent
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*agent.Agent),
	}
}

// Register registers an agent with the registry
func (r *AgentRegistry) Register(id string, agent *agent.Agent) {
	r.agents[id] = agent
}

// Get retrieves an agent from the registry
func (r *AgentRegistry) Get(id string) (*agent.Agent, bool) {
	agent, ok := r.agents[id]
	return agent, ok
}

// List returns all registered agents
func (r *AgentRegistry) List() map[string]*agent.Agent {
	return r.agents
}
