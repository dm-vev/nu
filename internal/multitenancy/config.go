package multitenancy

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrTenantNotFound is returned when a tenant is not found
	ErrTenantNotFound = errors.New("tenant not found")
)

// TenantConfig represents the configuration for a tenant
type TenantConfig struct {
	// OrgID is the organization ID
	OrgID string

	// LLMAPIKeys maps LLM provider names to API keys
	LLMAPIKeys map[string]string

	// VectorStoreConfig contains vector store configuration
	VectorStoreConfig map[string]interface{}

	// DataStoreConfig contains data store configuration
	DataStoreConfig map[string]interface{}

	// Custom contains custom configuration values
	Custom map[string]interface{}
}

// ConfigManager manages tenant configurations
type ConfigManager struct {
	configs map[string]*TenantConfig
	mu      sync.RWMutex
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs: make(map[string]*TenantConfig),
	}
}

// RegisterTenant registers a new tenant
func (m *ConfigManager) RegisterTenant(config *TenantConfig) error {
	if config.OrgID == "" {
		return errors.New("organization ID is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[config.OrgID] = config
	return nil
}

// GetTenantConfig returns the configuration for a tenant
func (m *ConfigManager) GetTenantConfig(ctx context.Context) (*TenantConfig, error) {
	orgID, err := GetOrgID(ctx)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	config, ok := m.configs[orgID]
	if !ok {
		return nil, ErrTenantNotFound
	}

	return config, nil
}

// GetLLMAPIKey returns the API key for the given LLM provider
func (m *ConfigManager) GetLLMAPIKey(ctx context.Context, provider string) (string, error) {
	config, err := m.GetTenantConfig(ctx)
	if err != nil {
		return "", err
	}

	apiKey, ok := config.LLMAPIKeys[provider]
	if !ok {
		return "", errors.New("API key not found for provider: " + provider)
	}

	return apiKey, nil
}

// GetVectorStoreConfig returns the vector store configuration
func (m *ConfigManager) GetVectorStoreConfig(ctx context.Context) (map[string]interface{}, error) {
	config, err := m.GetTenantConfig(ctx)
	if err != nil {
		return nil, err
	}

	return config.VectorStoreConfig, nil
}

// GetDataStoreConfig returns the data store configuration
func (m *ConfigManager) GetDataStoreConfig(ctx context.Context) (map[string]interface{}, error) {
	config, err := m.GetTenantConfig(ctx)
	if err != nil {
		return nil, err
	}

	return config.DataStoreConfig, nil
}

// GetCustomConfig returns a custom configuration value
func (m *ConfigManager) GetCustomConfig(ctx context.Context, key string) (interface{}, error) {
	config, err := m.GetTenantConfig(ctx)
	if err != nil {
		return nil, err
	}

	value, ok := config.Custom[key]
	if !ok {
		return nil, errors.New("custom config not found for key: " + key)
	}

	return value, nil
}
