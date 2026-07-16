package search

import (
	"github.com/weaviate/weaviate/entities/models"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/data/weaviate/graph/entity"
)

// ParseResults parses GraphQL response into GraphSearchResult slice.
func ParseResults(result *models.GraphQLResponse, className string) []contracts.GraphSearchResult {
	results := []contracts.GraphSearchResult{}

	if result.Data == nil {
		return results
	}

	getData, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		return results
	}

	classData, ok := getData[className].([]interface{})
	if !ok {
		return results
	}

	for _, item := range classData {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse entity
		entities := entity.ParseResults(result, className)
		if len(entities) == 0 {
			// Parse entity from itemMap directly
			parsedEntity := entity.ParseMap(itemMap)
			if parsedEntity.ID == "" {
				continue
			}

			searchResult := contracts.GraphSearchResult{
				Entity: parsedEntity,
			}

			// Extract score from _additional
			if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
				if certainty, ok := additional["certainty"].(float64); ok {
					searchResult.Score = float32(certainty)
				} else if score, ok := additional["score"].(float64); ok {
					// Normalize BM25 score to 0-1 range (approximate)
					searchResult.Score = float32(score / (score + 1))
				}
			}

			results = append(results, searchResult)
		}
	}

	// If we didn't get results from the loop, try parsing entities directly
	if len(results) == 0 {
		entities := entity.ParseResults(result, className)
		for _, parsedEntity := range entities {
			results = append(results, contracts.GraphSearchResult{
				Entity: parsedEntity,
				Score:  0.5, // Default score if not available
			})
		}
	}

	return results
}
