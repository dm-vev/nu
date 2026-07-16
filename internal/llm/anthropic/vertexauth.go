package anthropic

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/auth/credentials"
	"cloud.google.com/go/auth/oauth2adapt"
)

// NewVertexConfig creates a new VertexConfig using Application Default Credentials
func NewVertexConfig(ctx context.Context, region, projectID string) (*VertexConfig, error) {
	if region == "" {
		return nil, fmt.Errorf("region is required for Vertex AI")
	}
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required for Vertex AI")
	}

	// Detect Application Default Credentials
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find default credentials: %w", err)
	}

	config := &VertexConfig{
		Enabled:     true,
		ProjectID:   projectID,
		Region:      region,
		TokenSource: oauth2adapt.TokenSourceFromTokenProvider(creds),
		Credentials: creds,
	}
	config.parseRegions()
	return config, nil
}

// NewVertexConfigWithCredentials creates a new VertexConfig with explicit credentials file
func NewVertexConfigWithCredentials(ctx context.Context, region, projectID, credentialsPath string) (*VertexConfig, error) {
	if region == "" {
		return nil, fmt.Errorf("region is required for Vertex AI")
	}
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required for Vertex AI")
	}
	if credentialsPath == "" {
		return nil, fmt.Errorf("credentialsPath is required")
	}

	// Read credentials file
	credentialsFile, err := os.Open(credentialsPath) // #nosec G304 - credentialsPath is validated and comes from trusted source
	if err != nil {
		return nil, fmt.Errorf("failed to open credentials file %s: %w", credentialsPath, err)
	}
	defer func() {
		_ = credentialsFile.Close()
	}()

	credentialsJSON, err := io.ReadAll(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file %s: %w", credentialsPath, err)
	}

	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		CredentialsJSON: credentialsJSON,
		Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials from %s: %w", credentialsPath, err)
	}

	config := &VertexConfig{
		Enabled:     true,
		ProjectID:   projectID,
		Region:      region,
		TokenSource: oauth2adapt.TokenSourceFromTokenProvider(creds),
		Credentials: creds,
	}
	config.parseRegions()
	return config, nil
}

// NewVertexConfigWithCredentialsContent creates a new VertexConfig with explicit credentials content
func NewVertexConfigWithCredentialsContent(ctx context.Context, region, projectID, credentialsContent string) (*VertexConfig, error) {
	if region == "" {
		return nil, fmt.Errorf("region is required for Vertex AI")
	}
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required for Vertex AI")
	}
	if credentialsContent == "" {
		return nil, fmt.Errorf("credentialsContent is required")
	}

	// Try to parse credentials as JSON first
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		CredentialsJSON: []byte(credentialsContent),
		Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
	if err != nil {
		// If JSON parsing fails, try base64 decoding first
		decodedContent, decodeErr := base64.StdEncoding.DecodeString(credentialsContent)
		if decodeErr == nil {
			// Successfully decoded, try parsing the decoded content as JSON
			creds, err = credentials.DetectDefault(&credentials.DetectOptions{
				CredentialsJSON: decodedContent,
				Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to load credentials from decoded base64 content: %w", err)
			}
		} else {
			// Base64 decode also failed, return original JSON error
			return nil, fmt.Errorf("failed to load credentials from content: %w", err)
		}
	}

	config := &VertexConfig{
		Enabled:     true,
		ProjectID:   projectID,
		Region:      region,
		TokenSource: oauth2adapt.TokenSourceFromTokenProvider(creds),
		Credentials: creds,
	}
	config.parseRegions()
	return config, nil
}

// GetAuthHeaders returns the authentication headers for Vertex AI requests
func (vc *VertexConfig) GetAuthHeaders(ctx context.Context) (map[string]string, error) {
	if !vc.Enabled {
		return nil, fmt.Errorf("vertex AI is not enabled")
	}

	var token string
	if vc.AccessToken != "" {
		token = vc.AccessToken
	} else if vc.TokenSource != nil {
		oauthToken, err := vc.TokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
		token = oauthToken.AccessToken
	} else {
		return nil, fmt.Errorf("no authentication method available")
	}

	return map[string]string{"Authorization": "Bearer " + token}, nil
}
