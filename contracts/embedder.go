package contracts

import (
	"context"
)

// Embedder represents a service that can convert text into embeddings
type Embedder interface {
	// Embed generates an embedding for the given text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// CalculateSimilarity calculates the similarity between two embeddings
	CalculateSimilarity(vec1, vec2 []float32, metric string) (float32, error)
}
