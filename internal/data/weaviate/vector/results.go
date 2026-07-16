package vector

import (
	"context"

	"github.com/weaviate/weaviate/entities/models"

	"nu/internal/contracts"
)

func (s *Store) parseSearchResults(result *models.GraphQLResponse, className string) ([]contracts.SearchResult, error) {
	var searchResults []contracts.SearchResult

	// Add debug logging
	s.logger.Info(context.Background(), "Parsing search results", map[string]interface{}{
		"className":  className,
		"resultData": result.Data,
	})

	// Check if result.Data is nil
	if result.Data == nil {
		s.logger.Warn(context.Background(), "Empty response data from Weaviate", nil)
		return []contracts.SearchResult{}, nil // Return empty results instead of error
	}

	// Get the results array
	getMap, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		// Log the actual structure for debugging
		s.logger.Error(context.Background(), "Invalid response format", map[string]interface{}{
			"data": result.Data,
		})
		// Return empty results instead of error for production use
		return []contracts.SearchResult{}, nil
	}

	results, ok := getMap[className].([]interface{})
	if !ok {
		// Return empty results if no matches found
		s.logger.Info(context.Background(), "No results found for class", map[string]interface{}{
			"className": className,
			"getMap":    getMap,
		})
		return searchResults, nil
	}

	for _, r := range results {
		result := r.(map[string]interface{})
		additional, ok := result["_additional"].(map[string]interface{})
		if !ok {
			s.logger.Warn(context.Background(), "Missing _additional field in result", map[string]interface{}{
				"result": result,
			})
			continue
		}

		content, ok := result["content"].(string)
		if !ok {
			s.logger.Warn(context.Background(), "Missing content field in result", map[string]interface{}{
				"result": result,
			})
			continue
		}

		id, ok := additional["id"].(string)
		if !ok {
			s.logger.Warn(context.Background(), "Missing id field in result", map[string]interface{}{
				"additional": additional,
			})
			continue
		}

		certainty, ok := additional["certainty"].(float64)
		if !ok {
			s.logger.Warn(context.Background(), "Missing certainty field in result", map[string]interface{}{
				"additional": additional,
			})
			// Use a default certainty value
			certainty = 0.5
		}

		doc := contracts.Document{
			ID:       id,
			Content:  content,
			Metadata: make(map[string]interface{}),
		}

		// Copy all properties except content and _additional to metadata
		for k, v := range result {
			if k != "content" && k != "_additional" {
				doc.Metadata[k] = v
			}
		}

		searchResults = append(searchResults, contracts.SearchResult{
			Document: doc,
			Score:    float32(certainty),
		})
	}

	return searchResults, nil
}
