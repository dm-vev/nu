package embedding

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

// Embed generates an embedding using Gemini API with default configuration
func (e *GeminiEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return e.EmbedWithConfig(ctx, text, e.config)
}

// EmbedWithConfig generates an embedding using Gemini API with custom configuration
func (e *GeminiEmbedder) EmbedWithConfig(ctx context.Context, text string, config Config) ([]float32, error) {
	model := config.Model
	if model == "" {
		model = e.model
	}

	// Build the embed content config
	embedConfig := &genai.EmbedContentConfig{}

	// Set task type if configured
	if e.taskType != "" {
		embedConfig.TaskType = e.taskType
	}

	// Set output dimensionality if specified and supported
	// #nosec G115 - dimensions are bounded by embedding model limits (typically < 10000)
	if config.Dimensions > 0 && config.Dimensions <= 32767 {
		dims := int32(config.Dimensions)
		embedConfig.OutputDimensionality = &dims
	}

	e.logger.Debug(ctx, "Generating embedding with Gemini", map[string]interface{}{
		"model":      model,
		"task_type":  e.taskType,
		"dimensions": config.Dimensions,
	})

	// Create content parts for embedding
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: text},
			},
		},
	}

	// Generate embedding
	result, err := e.client.Models.EmbedContent(ctx, model, contents, embedConfig)
	if err != nil {
		e.logger.Error(ctx, "Failed to generate embedding", map[string]interface{}{
			"error": err.Error(),
			"model": model,
		})
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if result == nil || len(result.Embeddings) == 0 || result.Embeddings[0] == nil || len(result.Embeddings[0].Values) == 0 {
		return nil, errors.New("no embedding data returned from Gemini API")
	}

	e.logger.Debug(ctx, "Successfully generated embedding", map[string]interface{}{
		"model":      model,
		"dimensions": len(result.Embeddings[0].Values),
	})

	return result.Embeddings[0].Values, nil
}

// EmbedBatch generates embeddings for multiple texts using default configuration
func (e *GeminiEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return e.EmbedBatchWithConfig(ctx, texts, e.config)
}

// EmbedBatchWithConfig generates embeddings for multiple texts with custom configuration
func (e *GeminiEmbedder) EmbedBatchWithConfig(ctx context.Context, texts []string, config Config) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	model := config.Model
	if model == "" {
		model = e.model
	}

	// Build the embed content config
	embedConfig := &genai.EmbedContentConfig{}

	// Set task type if configured
	if e.taskType != "" {
		embedConfig.TaskType = e.taskType
	}

	// Set output dimensionality if specified and supported
	// #nosec G115 - dimensions are bounded by embedding model limits (typically < 10000)
	if config.Dimensions > 0 && config.Dimensions <= 32767 {
		dims := int32(config.Dimensions)
		embedConfig.OutputDimensionality = &dims
	}

	e.logger.Debug(ctx, "Generating batch embeddings with Gemini", map[string]interface{}{
		"model":      model,
		"task_type":  e.taskType,
		"dimensions": config.Dimensions,
		"batch_size": len(texts),
	})

	// Create content parts for batch embedding
	contents := make([]*genai.Content, len(texts))
	for i, text := range texts {
		contents[i] = &genai.Content{
			Parts: []*genai.Part{
				{Text: text},
			},
		}
	}

	// Generate embeddings - EmbedContent handles multiple contents
	result, err := e.client.Models.EmbedContent(ctx, model, contents, embedConfig)
	if err != nil {
		e.logger.Error(ctx, "Failed to generate batch embeddings", map[string]interface{}{
			"error":      err.Error(),
			"model":      model,
			"batch_size": len(texts),
		})
		return nil, fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	if result == nil || len(result.Embeddings) == 0 {
		return nil, errors.New("no embedding data returned from Gemini API")
	}

	// Extract embeddings
	embeddings := make([][]float32, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		if emb == nil || len(emb.Values) == 0 {
			return nil, fmt.Errorf("empty embedding at index %d", i)
		}
		embeddings[i] = emb.Values
	}

	e.logger.Debug(ctx, "Successfully generated batch embeddings", map[string]interface{}{
		"model":      model,
		"batch_size": len(embeddings),
	})

	return embeddings, nil
}

// CalculateSimilarity calculates the similarity between two embeddings
func (e *GeminiEmbedder) CalculateSimilarity(vec1, vec2 []float32, metric string) (float32, error) {
	if metric == "" {
		metric = e.config.SimilarityMetric
	}
	return CalculateSimilarity(vec1, vec2, metric)
}

// GetConfig returns the current configuration
func (e *GeminiEmbedder) GetConfig() Config {
	return e.config
}

// GetModel returns the model name being used
func (e *GeminiEmbedder) GetModel() string {
	return e.model
}
