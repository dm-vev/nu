package agent

import "fmt"

func (a *Agent) validateSubAgents() error {
	return a.checkCircularDependency(make(map[string]bool), make(map[string]bool))
}

func (a *Agent) checkCircularDependency(visited, recursionStack map[string]bool) error {
	agentID := a.getUniqueID()
	visited[agentID] = true
	recursionStack[agentID] = true
	for _, subAgent := range a.subAgents {
		subAgentID := subAgent.getUniqueID()
		if recursionStack[subAgentID] {
			return fmt.Errorf("circular dependency detected: %s -> %s", a.name, subAgent.name)
		}
		if !visited[subAgentID] {
			if err := subAgent.checkCircularDependency(visited, recursionStack); err != nil {
				return err
			}
		}
	}
	delete(recursionStack, agentID)
	return nil
}

func (a *Agent) getUniqueID() string {
	if a.name != "" {
		return a.name
	}
	return fmt.Sprintf("%p", a)
}

// HasSubAgent reports whether a direct sub-agent has the given name.
func (a *Agent) HasSubAgent(name string) bool {
	for _, subAgent := range a.subAgents {
		if subAgent.name == name {
			return true
		}
	}
	return false
}

// GetSubAgent retrieves a direct sub-agent by name.
func (a *Agent) GetSubAgent(name string) (*Agent, bool) {
	for _, subAgent := range a.subAgents {
		if subAgent.name == name {
			return subAgent, true
		}
	}
	return nil, false
}

func validateAgentTree(root *Agent, maxDepth int) error {
	if err := root.validateSubAgents(); err != nil {
		return err
	}
	depth := calculateMaxDepth(root, 0)
	if depth > maxDepth {
		return fmt.Errorf("agent tree depth %d exceeds maximum allowed depth %d", depth, maxDepth)
	}
	return validateAgentComponents(root)
}

func calculateMaxDepth(agent *Agent, currentDepth int) int {
	if len(agent.subAgents) == 0 {
		return currentDepth
	}
	maxDepth := currentDepth
	for _, subAgent := range agent.subAgents {
		if depth := calculateMaxDepth(subAgent, currentDepth+1); depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

func validateAgentComponents(agent *Agent) error {
	if !agent.isRemote && agent.llm == nil {
		return fmt.Errorf("agent %s is missing required LLM", agent.name)
	}
	if agent.name == "" {
		return fmt.Errorf("agent is missing a name, which is recommended for sub-agents")
	}
	for _, subAgent := range agent.subAgents {
		if err := validateAgentComponents(subAgent); err != nil {
			return err
		}
	}
	return nil
}
