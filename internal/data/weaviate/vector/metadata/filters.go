package metadata

import (
	"fmt"
	"strings"

	"nu/internal/data/embedding"
)

// FilterToWeaviateFormat converts a FilterGroup to a Weaviate-compatible filter format.
// This is the recommended function to use for Weaviate vector store filters.
func FilterToWeaviateFormat(group embedding.FilterGroup) map[string]interface{} {
	// Simple case: single filter
	if len(group.Filters) == 1 && len(group.SubGroups) == 0 {
		filter := group.Filters[0]
		return weaviateFilterFromSingle(filter)
	}

	// Complex case: multiple filters or sub-groups
	var conditions []map[string]interface{}

	// Add filters
	for _, filter := range group.Filters {
		conditions = append(conditions, weaviateFilterFromSingle(filter))
	}

	// Add sub-groups
	for _, subGroup := range group.SubGroups {
		conditions = append(conditions, FilterToWeaviateFormat(subGroup))
	}

	// Combine with operator
	if len(conditions) > 0 {
		op := "And"
		if strings.ToLower(group.Operator) == "or" {
			op = "Or"
		}
		return map[string]interface{}{
			"operator": op,
			"operands": conditions,
		}
	}

	return nil
}

// weaviateFilterFromSingle converts a single Filter to Weaviate format
func weaviateFilterFromSingle(filter embedding.Filter) map[string]interface{} {
	result := map[string]interface{}{
		"path": []string{filter.Field},
	}

	// Map operator
	switch strings.ToLower(filter.Operator) {
	case "=", "==", "eq":
		result["operator"] = "Equal"
		result["valueString"] = fmt.Sprint(filter.Value)
	case "!=", "<>", "ne":
		result["operator"] = "NotEqual"
		result["valueString"] = fmt.Sprint(filter.Value)
	case ">", "gt":
		result["operator"] = "GreaterThan"
		result["valueNumber"] = embedding.Float64(filter.Value)
	case ">=", "gte":
		result["operator"] = "GreaterThanEqual"
		result["valueNumber"] = embedding.Float64(filter.Value)
	case "<", "lt":
		result["operator"] = "LessThan"
		result["valueNumber"] = embedding.Float64(filter.Value)
	case "<=", "lte":
		result["operator"] = "LessThanEqual"
		result["valueNumber"] = embedding.Float64(filter.Value)
	case "contains":
		result["operator"] = "Like"
		result["valueString"] = fmt.Sprint(filter.Value)
	case "in":
		// Handle 'in' operator
		result["operator"] = "ContainsAny"
		if values, ok := filter.Value.([]interface{}); ok {
			strValues := make([]string, len(values))
			for i, v := range values {
				strValues[i] = fmt.Sprint(v)
			}
			result["valueString"] = strValues
		} else {
			result["valueString"] = []string{fmt.Sprint(filter.Value)}
		}
	case "not_in":
		// Handle 'not_in' operator
		result["operator"] = "NotContainsAny"
		if values, ok := filter.Value.([]interface{}); ok {
			strValues := make([]string, len(values))
			for i, v := range values {
				strValues[i] = fmt.Sprint(v)
			}
			result["valueString"] = strValues
		} else {
			result["valueString"] = []string{fmt.Sprint(filter.Value)}
		}
	default:
		result["operator"] = "Equal"
		result["valueString"] = fmt.Sprint(filter.Value)
	}

	return result
}

// CreateWeaviateFilter creates a simple Weaviate filter for a single field.
// This is a convenience function for simple filter cases.
// For complex filters, use FilterGroup and FilterToWeaviateFormat.
func CreateWeaviateFilter(field string, operator string, value interface{}) map[string]interface{} {
	// Create a simple filter
	filter := embedding.NewFilter(field, operator, value)

	// Convert to Weaviate format
	return weaviateFilterFromSingle(filter)
}

// CreateWeaviateAndFilter creates a Weaviate filter with AND logic for multiple conditions.
// This is a convenience function for common filter cases.
func CreateWeaviateAndFilter(conditions ...map[string]interface{}) map[string]interface{} {
	if len(conditions) == 0 {
		return nil
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	return map[string]interface{}{
		"operator": "And",
		"operands": conditions,
	}
}

// CreateWeaviateOrFilter creates a Weaviate filter with OR logic for multiple conditions.
// This is a convenience function for common filter cases.
func CreateWeaviateOrFilter(conditions ...map[string]interface{}) map[string]interface{} {
	if len(conditions) == 0 {
		return nil
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	return map[string]interface{}{
		"operator": "Or",
		"operands": conditions,
	}
}
