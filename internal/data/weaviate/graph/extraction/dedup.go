package extraction

import (
	"context"
	"fmt"

	"nu/internal/contracts"
	"nu/internal/data/embedding"
)

// Deduplicate removes duplicate entities using embedding similarity.
func Deduplicate(
	ctx context.Context,
	entities []contracts.Entity,
	threshold float32,
	embedder embedding.Client,
) ([]contracts.Entity, error) {
	if len(entities) <= 1 {
		return entities, nil
	}

	// Generate embeddings for all entities
	texts := make([]string, len(entities))
	for i, e := range entities {
		texts[i] = e.Name + " " + e.Description
	}

	embeddings, err := embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Deduplicate using cosine similarity
	unique := []contracts.Entity{}
	uniqueEmbeddings := [][]float32{}

	for i, entity := range entities {
		isDuplicate := false

		for j, existing := range uniqueEmbeddings {
			similarity, err := embedder.CalculateSimilarity(embeddings[i], existing, "cosine")
			if err != nil {
				continue
			}

			if similarity >= threshold {
				// Merge: prefer longer names and combine properties
				if len(entity.Name) > len(unique[j].Name) {
					unique[j].Name = entity.Name
				}
				if len(entity.Description) > len(unique[j].Description) {
					unique[j].Description = entity.Description
				}
				// Merge properties
				if entity.Properties != nil {
					if unique[j].Properties == nil {
						unique[j].Properties = make(map[string]interface{})
					}
					for k, v := range entity.Properties {
						if _, exists := unique[j].Properties[k]; !exists {
							unique[j].Properties[k] = v
						}
					}
				}
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			unique = append(unique, entity)
			uniqueEmbeddings = append(uniqueEmbeddings, embeddings[i])
		}
	}

	return unique, nil
}
