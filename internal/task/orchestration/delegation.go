package orchestration

import (
	"context"
	"fmt"
	"reflect"

	"github.com/dm-vev/nu/agent"
	"github.com/dm-vev/nu/contracts"
)

// DelegationAgent is an agent that can delegate tasks to other agents
type HandoffDelegationAgent struct {
	*agent.Agent
	registry *AgentRegistry
}

// NewDelegationAgent creates a new delegation agent
func NewHandoffDelegationAgent(baseAgent *agent.Agent, registry *AgentRegistry) *HandoffDelegationAgent {
	return &HandoffDelegationAgent{
		Agent:    baseAgent,
		registry: registry,
	}
}

// Delegate delegates a task to another agent
func (a *HandoffDelegationAgent) Delegate(ctx context.Context, targetAgentID string, query string, preserveMemory bool) (string, error) {
	// Get the target agent
	targetAgent, ok := a.registry.Get(targetAgentID)
	if !ok {
		return "", fmt.Errorf("agent not found: %s", targetAgentID)
	}

	// Copy memory if needed
	if preserveMemory {
		err := a.copyMemory(ctx, targetAgent)
		if err != nil {
			return "", fmt.Errorf("failed to copy memory: %w", err)
		}
	}

	// Run the target agent
	response, err := targetAgent.Run(ctx, query)
	if err != nil {
		return "", fmt.Errorf("agent execution failed: %w", err)
	}

	return response, nil
}

// copyMemory copies memory from this agent to the target agent
func (a *HandoffDelegationAgent) copyMemory(ctx context.Context, targetAgent *agent.Agent) error {
	// Get source memory using reflection since there's no exported getter
	agentValue := reflect.ValueOf(a.Agent).Elem()
	memoryField := agentValue.FieldByName("memory")

	if !memoryField.IsValid() || memoryField.IsNil() {
		return fmt.Errorf("source agent has no memory")
	}

	sourceMemory := memoryField.Interface().(contracts.Memory)

	// Get target memory using reflection
	targetValue := reflect.ValueOf(targetAgent).Elem()
	targetMemoryField := targetValue.FieldByName("memory")

	if !targetMemoryField.IsValid() || targetMemoryField.IsNil() {
		return fmt.Errorf("target agent has no memory")
	}

	targetMemory := targetMemoryField.Interface().(contracts.Memory)

	// Get messages from source memory
	messages, err := sourceMemory.GetMessages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get messages from source memory: %w", err)
	}

	// Add messages to target memory
	for _, msg := range messages {
		err := targetMemory.AddMessage(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to add message to target memory: %w", err)
		}
	}

	return nil
}

// AgentPool represents a pool of specialized agents
type AgentPool struct {
	agents       map[string]*agent.Agent
	descriptions map[string]string
}

// NewAgentPool creates a new agent pool
func NewAgentPool() *AgentPool {
	return &AgentPool{
		agents:       make(map[string]*agent.Agent),
		descriptions: make(map[string]string),
	}
}

// Add adds an agent to the pool
func (p *AgentPool) Add(id string, agent *agent.Agent, description string) {
	p.agents[id] = agent
	p.descriptions[id] = description
}

// Get retrieves an agent from the pool
func (p *AgentPool) Get(id string) (*agent.Agent, bool) {
	agent, ok := p.agents[id]
	return agent, ok
}

// GetDescription retrieves an agent's description
func (p *AgentPool) GetDescription(id string) (string, bool) {
	desc, ok := p.descriptions[id]
	return desc, ok
}

// List returns all agents in the pool
func (p *AgentPool) List() map[string]*agent.Agent {
	return p.agents
}

// ListDescriptions returns all agent descriptions
func (p *AgentPool) ListDescriptions() map[string]string {
	return p.descriptions
}
