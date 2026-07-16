package config

import (
	"time"
)

// DeploymentConfigValueType represents the type of a deployment configuration value.
type DeploymentConfigValueType string

const (
	DeploymentConfigValueTypePlain  DeploymentConfigValueType = "plain"
	DeploymentConfigValueTypeSecret DeploymentConfigValueType = "secret"
)

// DeploymentConfigSecretRef references a secret in the secret manager.
type DeploymentConfigSecretRef struct {
	ProviderID string  `json:"provider_id"`
	Key        string  `json:"key"`
	Instance   *string `json:"instance,omitempty"`
}

// DeploymentConfigResolvedValue contains a resolved configuration value.
type DeploymentConfigResolvedValue struct {
	Type         DeploymentConfigValueType  `json:"type"`
	Value        string                     `json:"value"`
	SecretRef    *DeploymentConfigSecretRef `json:"secret_ref,omitempty"`
	StoreInVault bool                       `json:"store_in_vault,omitempty"`
}

// DeploymentConfigResponse is a resolved deployment configuration response.
type DeploymentConfigResponse struct {
	ID          string                        `json:"id"`
	OrgID       string                        `json:"org_id"`
	UserID      string                        `json:"user_id"`
	InstanceID  string                        `json:"instance_id"`
	Environment string                        `json:"environment"`
	Key         string                        `json:"key"`
	Value       DeploymentConfigResolvedValue `json:"value"`
	Description *string                       `json:"description,omitempty"`
	CreatedBy   *string                       `json:"created_by,omitempty"`
	UpdatedBy   *string                       `json:"updated_by,omitempty"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

// DeploymentConfigAgentResponse is a resolved agent configuration from the service.
type DeploymentConfigAgentResponse struct {
	AgentConfig struct {
		ID            string    `json:"id"`
		AgentName     string    `json:"agent_name"`
		Environment   string    `json:"environment"`
		DisplayName   string    `json:"display_name"`
		Description   string    `json:"description"`
		Goal          string    `json:"goal"`
		SystemPrompt  string    `json:"system_prompt"`
		SchemaVersion string    `json:"schema_version"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
	} `json:"agent_config"`
	GeneratedYAML     string            `json:"generated_yaml"`     // YAML generated from structured data
	ResolvedYAML      string            `json:"resolved_yaml"`      // YAML with variables resolved
	ResolvedVariables map[string]string `json:"resolved_variables"` // Variable mappings
	MissingVariables  []string          `json:"missing_variables"`  // Unresolved variables
}
