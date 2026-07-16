package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dm-vev/nu/internal/data/embedding"
)

func TestOpenAIModelConstants(t *testing.T) {
	assert.Equal(t, "text-embedding-3-small", ModelTextEmbedding3Small)
	assert.Equal(t, "text-embedding-3-large", ModelTextEmbedding3Large)
	assert.Equal(t, "text-embedding-ada-002", ModelTextEmbeddingAda002)
	assert.Equal(t, ModelTextEmbedding3Small, DefaultModel)
}

func TestNew(t *testing.T) {
	embedder := New("test-api-key", "")
	assert.NotNil(t, embedder)
	assert.Equal(t, DefaultModel, embedder.model)
	assert.Equal(t, DefaultModel, embedder.config.Model)
}

func TestNewClientWithModel(t *testing.T) {
	embedder := New("test-api-key", ModelTextEmbedding3Large)
	assert.NotNil(t, embedder)
	assert.Equal(t, ModelTextEmbedding3Large, embedder.model)
}

func TestNewWithConfig(t *testing.T) {
	config := embedding.Config{
		Model:            ModelTextEmbedding3Large,
		Dimensions:       1024,
		SimilarityMetric: "dot_product",
	}
	embedder := NewWithConfig("test-api-key", config)
	assert.NotNil(t, embedder)
	assert.Equal(t, ModelTextEmbedding3Large, embedder.model)
	assert.Equal(t, 1024, embedder.config.Dimensions)
}

func TestNewWithConfig_DefaultModel(t *testing.T) {
	config := embedding.Config{
		Dimensions: 512,
	}
	embedder := NewWithConfig("test-api-key", config)
	assert.NotNil(t, embedder)
	assert.Equal(t, DefaultModel, embedder.model)
}

func TestClient_GetConfig(t *testing.T) {
	config := embedding.DefaultConfig(ModelTextEmbedding3Small)
	embedder := &Client{
		config: config,
	}
	assert.Equal(t, config, embedder.GetConfig())
}

func TestClient_GetModel(t *testing.T) {
	embedder := &Client{
		model: ModelTextEmbedding3Large,
	}
	assert.Equal(t, ModelTextEmbedding3Large, embedder.GetModel())
}

func TestClient_CalculateSimilarity(t *testing.T) {
	embedder := &Client{
		config: embedding.DefaultConfig(""),
	}

	t.Run("cosine similarity", func(t *testing.T) {
		vec1 := []float32{1.0, 0.0, 0.0}
		vec2 := []float32{1.0, 0.0, 0.0}

		similarity, err := embedder.CalculateSimilarity(vec1, vec2, "cosine")
		assert.NoError(t, err)
		assert.InDelta(t, 1.0, similarity, 0.01)
	})

	t.Run("default metric from config", func(t *testing.T) {
		vec1 := []float32{1.0, 0.0, 0.0}
		vec2 := []float32{1.0, 0.0, 0.0}

		similarity, err := embedder.CalculateSimilarity(vec1, vec2, "")
		assert.NoError(t, err)
		assert.InDelta(t, 1.0, similarity, 0.01)
	})
}
