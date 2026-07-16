package extraction

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"nu/internal/contracts"
)

// ParseResponse parses the LLM response into an ExtractionResult.
func ParseResponse(response, sourceText string, options *contracts.ExtractionOptions) (*contracts.ExtractionResult, error) {
	// Try to extract JSON from the response
	jsonStr := ExtractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var parsed struct {
		Entities []struct {
			Name        string                 `json:"name"`
			Type        string                 `json:"type"`
			Description string                 `json:"description"`
			Properties  map[string]interface{} `json:"properties,omitempty"`
		} `json:"entities"`
		Relationships []struct {
			Source      string                 `json:"source"`
			Target      string                 `json:"target"`
			Type        string                 `json:"type"`
			Description string                 `json:"description"`
			Strength    float32                `json:"strength,omitempty"`
			Properties  map[string]interface{} `json:"properties,omitempty"`
		} `json:"relationships"`
		Confidence float32 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to entities
	now := time.Now()
	entityNameToID := make(map[string]string)
	entities := make([]contracts.Entity, 0, len(parsed.Entities))

	for _, e := range parsed.Entities {
		if e.Name == "" {
			continue
		}

		// Filter by min confidence if specified
		if options.MinConfidence > 0 && parsed.Confidence < options.MinConfidence {
			continue
		}

		// Filter by entity types if specified
		if len(options.EntityTypes) > 0 {
			found := false
			for _, t := range options.EntityTypes {
				if strings.EqualFold(e.Type, t) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		id := uuid.New().String()
		entityNameToID[e.Name] = id

		entity := contracts.Entity{
			ID:          id,
			Name:        e.Name,
			Type:        e.Type,
			Description: e.Description,
			Properties:  e.Properties,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		entities = append(entities, entity)
	}

	// Convert to relationships
	relationships := make([]contracts.Relationship, 0, len(parsed.Relationships))

	for _, r := range parsed.Relationships {
		if r.Source == "" || r.Target == "" || r.Type == "" {
			continue
		}

		// Filter by relationship types if specified
		if len(options.RelationshipTypes) > 0 {
			found := false
			for _, t := range options.RelationshipTypes {
				if strings.EqualFold(r.Type, t) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		sourceID, ok := entityNameToID[r.Source]
		if !ok {
			// Source entity not found, skip
			continue
		}

		targetID, ok := entityNameToID[r.Target]
		if !ok {
			// Target entity not found, skip
			continue
		}

		strength := r.Strength
		if strength == 0 {
			strength = 1.0
		}

		rel := contracts.Relationship{
			ID:          uuid.New().String(),
			SourceID:    sourceID,
			TargetID:    targetID,
			Type:        strings.ToUpper(r.Type),
			Description: r.Description,
			Strength:    strength,
			Properties:  r.Properties,
			CreatedAt:   now,
		}
		relationships = append(relationships, rel)
	}

	confidence := parsed.Confidence
	if confidence == 0 {
		confidence = 0.7 // Default confidence if not provided
	}

	return &contracts.ExtractionResult{
		Entities:      entities,
		Relationships: relationships,
		SourceText:    sourceText,
		Confidence:    confidence,
	}, nil
}

// ExtractJSON extracts JSON content from a text response.
func ExtractJSON(text string) string {
	// Try to find JSON object in the response
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}

	// Find matching closing brace
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}

	return ""
}
