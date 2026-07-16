package grpc

import (
	"fmt"
	"sync"
)

// MicroserviceManager manages multiple agent microservices
type MicroserviceManager struct {
	services map[string]*Microservice
	mu       sync.RWMutex
}

// NewMicroserviceManager creates a new microservice manager
func NewMicroserviceManager() *MicroserviceManager {
	return &MicroserviceManager{
		services: make(map[string]*Microservice),
	}
}

// RegisterService registers a microservice with the manager
func (mm *MicroserviceManager) RegisterService(name string, service *Microservice) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.services[name]; exists {
		return fmt.Errorf("service with name %s already exists", name)
	}

	mm.services[name] = service
	return nil
}

// StartService starts a service by name
func (mm *MicroserviceManager) StartService(name string) error {
	mm.mu.RLock()
	service, exists := mm.services[name]
	mm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	return service.Start()
}

// StopService stops a service by name
func (mm *MicroserviceManager) StopService(name string) error {
	mm.mu.RLock()
	service, exists := mm.services[name]
	mm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	return service.Stop()
}

// StartAll starts all registered services
func (mm *MicroserviceManager) StartAll() error {
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
func (mm *MicroserviceManager) StopAll() error {
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
func (mm *MicroserviceManager) GetService(name string) (*Microservice, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	service, exists := mm.services[name]
	return service, exists
}

// ListServices returns all registered service names
func (mm *MicroserviceManager) ListServices() []string {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	names := make([]string, 0, len(mm.services))
	for name := range mm.services {
		names = append(names, name)
	}

	return names
}
