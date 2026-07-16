package anthropic

import (
	"fmt"
	"strings"
	"sync"

	"cloud.google.com/go/auth"
	"golang.org/x/oauth2"
)

// VertexConfig contains configuration for Google Vertex AI
type VertexConfig struct {
	Enabled     bool
	ProjectID   string
	Region      string
	AccessToken string
	TokenSource oauth2.TokenSource
	Credentials *auth.Credentials

	regions            []string
	currentRegionIndex int
	mu                 sync.Mutex
}

// parseRegions splits the Region field by comma and stores as a list
func (vc *VertexConfig) parseRegions() {
	if vc.Region == "" {
		vc.regions = []string{}
		return
	}

	parts := strings.Split(vc.Region, ",")
	vc.regions = make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			vc.regions = append(vc.regions, trimmed)
		}
	}
	vc.currentRegionIndex = 0
}

// GetCurrentRegion returns the current region for round-robin rotation
func (vc *VertexConfig) GetCurrentRegion() string {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if len(vc.regions) == 0 {
		return vc.Region
	}
	return vc.regions[vc.currentRegionIndex]
}

// RotateRegion moves to the next region in round-robin fashion
func (vc *VertexConfig) RotateRegion() {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if len(vc.regions) <= 1 {
		return
	}

	vc.currentRegionIndex = (vc.currentRegionIndex + 1) % len(vc.regions)
}

// GetBaseURL returns the Vertex AI base URL for the configured region
func (vc *VertexConfig) GetBaseURL() string {
	if !vc.Enabled {
		return ""
	}

	if vc.GetCurrentRegion() == "global" {
		return "https://aiplatform.googleapis.com"
	}

	return fmt.Sprintf("https://%s-aiplatform.googleapis.com", vc.GetCurrentRegion())
}

// ValidateVertexConfig validates the Vertex AI configuration
func (vc *VertexConfig) ValidateVertexConfig() error {
	if !vc.Enabled {
		return nil
	}

	if vc.Region == "" {
		return fmt.Errorf("region is required for Vertex AI")
	}

	if vc.ProjectID == "" {
		return fmt.Errorf("projectID is required for Vertex AI")
	}

	if vc.TokenSource == nil && vc.AccessToken == "" {
		return fmt.Errorf("either TokenSource or AccessToken must be provided for Vertex AI")
	}

	return nil
}

// GetSupportedRegions returns a list of regions that support Anthropic models on Vertex AI
func GetSupportedRegions() []string {
	return []string{
		"us-central1",
		"us-east5",
		"europe-west1",
		"europe-west4",
		"asia-southeast1",
		"asia-northeast3",
	}
}

// IsRegionSupported checks if a region supports Anthropic models on Vertex AI
func IsRegionSupported(region string) bool {
	supportedRegions := GetSupportedRegions()
	for _, supported := range supportedRegions {
		if region == supported {
			return true
		}
	}
	return false
}
