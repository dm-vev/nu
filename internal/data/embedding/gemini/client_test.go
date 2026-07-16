package gemini

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"

	"github.com/dm-vev/nu/internal/data/embedding"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("with empty model", func(t *testing.T) {
		config := DefaultConfig("")
		assert.Equal(t, DefaultModel, config.Model)
		assert.Equal(t, 768, config.Dimensions)
		assert.Equal(t, "float", config.EncodingFormat)
		assert.Equal(t, "cosine", config.SimilarityMetric)
	})

	t.Run("with custom model", func(t *testing.T) {
		config := DefaultConfig(ModelTextEmbedding005)
		assert.Equal(t, ModelTextEmbedding005, config.Model)
	})
}

func TestNewValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("missing API key for Gemini API", func(t *testing.T) {
		_, err := New(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API key is required")
	})

	t.Run("multiple credential types", func(t *testing.T) {
		_, err := New(ctx,
			WithBackend(genai.BackendVertexAI),
			WithCredentialsFile("/path/to/file.json"),
			WithCredentialsJSON([]byte(`{"key": "value"}`)),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only one credential type")
	})

	t.Run("missing Vertex AI credentials", func(t *testing.T) {
		_, err := New(ctx,
			WithBackend(genai.BackendVertexAI),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project ID, credentials file, credentials JSON, or API key are required")
	})
}

func TestOptions(t *testing.T) {
	embedder := &Client{}

	t.Run("WithModel", func(t *testing.T) {
		WithModel(ModelTextEmbedding005)(embedder)
		assert.Equal(t, ModelTextEmbedding005, embedder.model)
		assert.Equal(t, ModelTextEmbedding005, embedder.config.Model)
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		WithAPIKey("test-api-key")(embedder)
		assert.Equal(t, "test-api-key", embedder.apiKey)
	})

	t.Run("WithBackend", func(t *testing.T) {
		WithBackend(genai.BackendVertexAI)(embedder)
		assert.Equal(t, genai.BackendVertexAI, embedder.backend)
	})

	t.Run("WithProjectID", func(t *testing.T) {
		WithProjectID("my-project")(embedder)
		assert.Equal(t, "my-project", embedder.projectID)
	})

	t.Run("WithLocation", func(t *testing.T) {
		WithLocation("europe-west1")(embedder)
		assert.Equal(t, "europe-west1", embedder.location)
	})

	t.Run("WithCredentialsFile", func(t *testing.T) {
		WithCredentialsFile("/path/to/creds.json")(embedder)
		assert.Equal(t, "/path/to/creds.json", embedder.credentialsFile)
	})

	t.Run("WithCredentialsJSON", func(t *testing.T) {
		creds := []byte(`{"key": "value"}`)
		WithCredentialsJSON(creds)(embedder)
		assert.Equal(t, creds, embedder.credentialsJSON)
	})

	t.Run("WithTaskType", func(t *testing.T) {
		WithTaskType("RETRIEVAL_QUERY")(embedder)
		assert.Equal(t, "RETRIEVAL_QUERY", embedder.taskType)
	})

	t.Run("WithConfig", func(t *testing.T) {
		config := embedding.Config{
			Model:            "custom-model",
			Dimensions:       512,
			SimilarityMetric: "dot_product",
		}
		WithConfig(config)(embedder)
		assert.Equal(t, config, embedder.config)
		assert.Equal(t, "custom-model", embedder.model)
	})
}

func TestClient_CalculateSimilarity(t *testing.T) {
	embedder := &Client{
		config: DefaultConfig(""),
	}

	t.Run("cosine similarity", func(t *testing.T) {
		vec1 := []float32{1.0, 0.0, 0.0}
		vec2 := []float32{1.0, 0.0, 0.0}

		similarity, err := embedder.CalculateSimilarity(vec1, vec2, "cosine")
		require.NoError(t, err)
		assert.InDelta(t, 1.0, similarity, 0.01)
	})

	t.Run("euclidean distance", func(t *testing.T) {
		vec1 := []float32{1.0, 0.0, 0.0}
		vec2 := []float32{1.0, 0.0, 0.0}

		similarity, err := embedder.CalculateSimilarity(vec1, vec2, "euclidean")
		require.NoError(t, err)
		assert.InDelta(t, 1.0, similarity, 0.01)
	})

	t.Run("dot product", func(t *testing.T) {
		vec1 := []float32{1.0, 2.0, 3.0}
		vec2 := []float32{4.0, 5.0, 6.0}

		similarity, err := embedder.CalculateSimilarity(vec1, vec2, "dot_product")
		require.NoError(t, err)
		assert.InDelta(t, 32.0, similarity, 0.01)
	})

	t.Run("default metric from config", func(t *testing.T) {
		vec1 := []float32{1.0, 0.0, 0.0}
		vec2 := []float32{1.0, 0.0, 0.0}

		similarity, err := embedder.CalculateSimilarity(vec1, vec2, "")
		require.NoError(t, err)
		assert.InDelta(t, 1.0, similarity, 0.01)
	})

	t.Run("dimension mismatch", func(t *testing.T) {
		vec1 := []float32{1.0, 0.0}
		vec2 := []float32{1.0, 0.0, 0.0}

		_, err := embedder.CalculateSimilarity(vec1, vec2, "cosine")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "same dimensions")
	})

	t.Run("unsupported metric", func(t *testing.T) {
		vec1 := []float32{1.0, 0.0}
		vec2 := []float32{1.0, 0.0}

		_, err := embedder.CalculateSimilarity(vec1, vec2, "unsupported")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported similarity metric")
	})
}

func TestClient_GetConfig(t *testing.T) {
	config := DefaultConfig(ModelTextEmbedding005)
	embedder := &Client{
		config: config,
	}

	assert.Equal(t, config, embedder.GetConfig())
}

func TestClient_GetModel(t *testing.T) {
	embedder := &Client{
		model: ModelTextMultilingualEmbedding002,
	}

	assert.Equal(t, ModelTextMultilingualEmbedding002, embedder.GetModel())
}

func TestGeminiModelConstants(t *testing.T) {
	assert.Equal(t, "text-embedding-004", ModelTextEmbedding004)
	assert.Equal(t, "text-embedding-005", ModelTextEmbedding005)
	assert.Equal(t, "text-multilingual-embedding-002", ModelTextMultilingualEmbedding002)
	assert.Equal(t, ModelTextEmbedding004, DefaultModel)
}
