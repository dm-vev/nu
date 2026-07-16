package graph

import (
	"context"
	"fmt"
	"strings"

	"nu/internal/contracts"
	"nu/internal/data/weaviate/graph/extraction"
)

// ExtractFromText extracts entities and relationships from text using an LLM.
func (s *Store) ExtractFromText(ctx context.Context, text string, llm contracts.LLM, opts ...contracts.ExtractionOption) (*contracts.ExtractionResult, error) {
	if llm == nil {
		return nil, ErrNoLLM
	}

	if text == "" {
		return &contracts.ExtractionResult{
			Entities:      []contracts.Entity{},
			Relationships: []contracts.Relationship{},
			SourceText:    text,
			Confidence:    0,
		}, nil
	}

	options := applyExtractionOptions(opts)

	// Build the extraction prompt
	prompt := s.buildExtractionPrompt(text, options)

	// Call LLM to extract entities and relationships
	response, err := llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	// Parse the LLM response
	result, err := extraction.ParseResponse(response, text, options)
	if err != nil {
		s.logger.Warn(ctx, "Failed to parse extraction response, attempting fallback", map[string]interface{}{
			"error": err.Error(),
		})
		// Try to extract what we can
		result = &contracts.ExtractionResult{
			Entities:      []contracts.Entity{},
			Relationships: []contracts.Relationship{},
			SourceText:    text,
			Confidence:    0.3,
		}
	}

	// Deduplicate entities if embedder is available and threshold is set
	if s.embedder != nil && options.DedupThreshold > 0 && len(result.Entities) > 1 {
		deduped, err := extraction.Deduplicate(ctx, result.Entities, options.DedupThreshold, s.embedder)
		if err != nil {
			s.logger.Warn(ctx, "Failed to deduplicate entities", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			result.Entities = deduped
		}
	}

	// Apply entity limit
	if options.MaxEntities > 0 && len(result.Entities) > options.MaxEntities {
		result.Entities = result.Entities[:options.MaxEntities]
	}

	s.logger.Info(ctx, "Extracted entities and relationships", map[string]interface{}{
		"entities":      len(result.Entities),
		"relationships": len(result.Relationships),
		"confidence":    result.Confidence,
	})

	return result, nil
}

// buildExtractionPrompt constructs the prompt for entity/relationship extraction.
func (s *Store) buildExtractionPrompt(text string, options *contracts.ExtractionOptions) string {
	var sb strings.Builder

	sb.WriteString("Extract entities and relationships from the following text.\n\n")

	if options.SchemaGuided && s.schema != nil {
		sb.WriteString("Use the following schema:\n\n")
		sb.WriteString("Entity Types:\n")
		for _, et := range s.schema.EntityTypes {
			fmt.Fprintf(&sb, "- %s: %s\n", et.Name, et.Description)
		}
		sb.WriteString("\nRelationship Types:\n")
		for _, rt := range s.schema.RelationshipTypes {
			fmt.Fprintf(&sb, "- %s: %s (from %v to %v)\n", rt.Name, rt.Description, rt.SourceTypes, rt.TargetTypes)
		}
		sb.WriteString("\n")
	} else if len(options.EntityTypes) > 0 || len(options.RelationshipTypes) > 0 {
		if len(options.EntityTypes) > 0 {
			fmt.Fprintf(&sb, "Focus on entity types: %s\n", strings.Join(options.EntityTypes, ", "))
		}
		if len(options.RelationshipTypes) > 0 {
			fmt.Fprintf(&sb, "Focus on relationship types: %s\n", strings.Join(options.RelationshipTypes, ", "))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Respond with a JSON object containing:
{
  "entities": [
    {
      "name": "entity name",
      "type": "entity type",
      "description": "brief description of the entity"
    }
  ],
  "relationships": [
    {
      "source": "source entity name",
      "target": "target entity name",
      "type": "RELATIONSHIP_TYPE",
      "description": "brief description of the relationship"
    }
  ],
  "confidence": 0.8
}

Text to analyze:
`)
	sb.WriteString(text)
	sb.WriteString("\n\nJSON response:")

	return sb.String()
}
