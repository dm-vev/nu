package vector

import (
	"nu/internal/data/embedding"
	"nu/internal/data/weaviate/vector/metadata"
)

// FilterToWeaviateFormat converts metadata filters to Weaviate's filter format.
func FilterToWeaviateFormat(group embedding.FilterGroup) map[string]interface{} {
	return metadata.FilterToWeaviateFormat(group)
}

// CreateWeaviateFilter creates a simple Weaviate filter for one field.
func CreateWeaviateFilter(field, operator string, value interface{}) map[string]interface{} {
	return metadata.CreateWeaviateFilter(field, operator, value)
}

// CreateWeaviateAndFilter creates a Weaviate AND filter.
func CreateWeaviateAndFilter(conditions ...map[string]interface{}) map[string]interface{} {
	return metadata.CreateWeaviateAndFilter(conditions...)
}

// CreateWeaviateOrFilter creates a Weaviate OR filter.
func CreateWeaviateOrFilter(conditions ...map[string]interface{}) map[string]interface{} {
	return metadata.CreateWeaviateOrFilter(conditions...)
}
