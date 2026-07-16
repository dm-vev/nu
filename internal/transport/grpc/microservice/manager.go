package microservice

import (
	"fmt"
	"sync"
)

// Manager manages multiple agent microservices
type Manager struct {
	services map[string]*Service
	mu       sync.RWMutex
}

// NewManager creates a new microservice manager
func NewManager() *Manager {
	return &Manager{
		services: make(map[string]*Service),
	}
}

// RegisterService registers a microservice with the manager
func (mm *Manager) RegisterService(name string, service *Service) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.services[name]; exists {
		return fmt.Errorf("service with name %s already exists", name)
	}

	mm.services[name] = service
	return nil
}

// StartService starts a service by name
func (mm *Manager) StartService(name string) error {
	mm.mu.RLock()
	service, exists := mm.services[name]
	mm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	return service.Start()
}

// StopService stops a service by name
func (mm *Manager) StopService(name string) error {
	mm.mu.RLock()
	service, exists := mm.services[name]
	mm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	return service.Stop()
}

// StartAll starts all registered services
func (mm *Manager) StartAll() error {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	for name, service := range mm.services {
		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", name, err)
		}
	}

	return nil
}

// StopAll stops all running services
func (mm *Manager) StopAll() error {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var lastErr error
	for name, service := range mm.services {
		if err := service.Stop(); err != nil {
			lastErr = fmt.Errorf("failed to stop service %s: %w", name, err)
		}
	}

	return lastErr
}

// GetService returns a service by name
func (mm *Manager) GetService(name string) (*Service, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	service, exists := mm.services[name]
	return service, exists
}

// ListServices returns all registered service names
func (mm *Manager) ListServices() []string {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	names := make([]string, 0, len(mm.services))
	for name := range mm.services {
		names = append(names, name)
	}

	return names
}
