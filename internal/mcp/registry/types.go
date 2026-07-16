package registry

import "time"

// RegistryServer represents a server entry in the MCP Registry
type RegistryServer struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Namespace   string         `json:"namespace"`
	Version     string         `json:"version"`
	Tags        []string       `json:"tags,omitempty"`
	Category    string         `json:"category,omitempty"`
	Author      RegistryAuthor `json:"author"`
	Repository  RepositoryInfo `json:"repository,omitempty"`
	License     string         `json:"license,omitempty"`
	Homepage    string         `json:"homepage,omitempty"`

	Installation  InstallationInfo   `json:"installation"`
	Configuration ConfigurationInfo  `json:"configuration,omitempty"`
	Tools         []RegistryTool     `json:"tools,omitempty"`
	Resources     []RegistryResource `json:"resources,omitempty"`
	Prompts       []RegistryPrompt   `json:"prompts,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Downloads     int                `json:"downloads,omitempty"`
	Rating        float64            `json:"rating,omitempty"`
	Verified      bool               `json:"verified"`
}

type RegistryAuthor struct {
	Name   string `json:"name"`
	Email  string `json:"email,omitempty"`
	URL    string `json:"url,omitempty"`
	GitHub string `json:"github,omitempty"`
}

type RepositoryInfo struct {
	Type string `json:"type"`
	URL  string `json:"url"`
	Ref  string `json:"ref,omitempty"`
}

type InstallationInfo struct {
	Type    string                 `json:"type"`
	Command string                 `json:"command"`
	Args    []string               `json:"args,omitempty"`
	Env     map[string]string      `json:"env,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

type ConfigurationInfo struct {
	Required []ConfigOption `json:"required,omitempty"`
	Optional []ConfigOption `json:"optional,omitempty"`
}

type ConfigOption struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Required    bool        `json:"required"`
	Sensitive   bool        `json:"sensitive"`
}

type RegistryTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Category    string                 `json:"category,omitempty"`
}

type RegistryResource struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Pattern     string   `json:"pattern,omitempty"`
	MimeTypes   []string `json:"mime_types,omitempty"`
}

type RegistryPrompt struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Variables   []string `json:"variables,omitempty"`
	Category    string   `json:"category,omitempty"`
}

type SearchOptions struct {
	Query    string   `json:"query,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Category string   `json:"category,omitempty"`
	Author   string   `json:"author,omitempty"`
	Verified bool     `json:"verified,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}

type SearchResponse struct {
	Servers []RegistryServer `json:"servers"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}
