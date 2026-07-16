package registry

import (
	"sync"

	"github.com/dm-vev/nu/contracts"
)

// Registry implements the ToolRegistry interface
type Registry struct {
	tools map[string]contracts.Tool
	mu    sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]contracts.Tool),
	}
}

// Register registers a tool with the registry
func (r *Registry) Register(tool contracts.Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// Get returns a tool by name
func (r *Registry) Get(name string) (contracts.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []contracts.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var tools []contracts.Tool
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}
