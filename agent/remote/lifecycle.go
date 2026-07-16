package remote

import (
	"context"
	"fmt"
)

// Initialize connects to the remote agent and resolves missing metadata.
func (s Service) Initialize() (name, description string) {
	if err := s.Client.Connect(); err != nil {
		if s.Logger != nil {
			s.Logger.Warn(context.Background(), fmt.Sprintf("Failed to connect to remote agent %s during initialization: %v (will retry on first use)", s.URL, err), nil)
		}
		if s.Name == "" {
			return "Remote-Agent", s.Description
		}
		return s.Name, s.Description
	}

	name, description = s.Name, s.Description
	if name == "" || description == "" {
		metadata, err := s.Client.GetMetadata(context.Background())
		if err != nil {
			if s.Logger != nil {
				s.Logger.Warn(context.Background(), fmt.Sprintf("Failed to fetch metadata from remote agent %s: %v", s.URL, err), nil)
			}
			return name, description
		}
		if name == "" {
			name = metadata.Name
		}
		if description == "" {
			description = metadata.Description
		}
	}
	return name, description
}

// Disconnect closes the remote agent connection.
func (s Service) Disconnect() error {
	return s.Client.Disconnect()
}

// Metadata returns remote agent metadata as string values.
func (s Service) Metadata() (map[string]string, error) {
	metadata, err := s.Client.GetMetadata(context.Background())
	if err != nil {
		return nil, err
	}
	result := map[string]string{
		"name":          metadata.Name,
		"description":   metadata.Description,
		"system_prompt": metadata.SystemPrompt,
	}
	for key, value := range metadata.Properties {
		result[key] = value
	}
	return result, nil
}
