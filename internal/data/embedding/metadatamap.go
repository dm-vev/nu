package embedding

import "strings"

// FilterToMap converts a FilterGroup to a map for use with vector store filters
// Deprecated: This function produces a format that may not be compatible with all vector stores.
// For Weaviate, use FilterToWeaviateFormat instead.
func FilterToMap(group FilterGroup) map[string]interface{} {
	result := make(map[string]interface{})

	// Simple case: single filter
	if len(group.Filters) == 1 && len(group.SubGroups) == 0 {
		filter := group.Filters[0]
		result[filter.Field] = map[string]interface{}{
			"operator": operatorToMapKey(filter.Operator),
			"value":    filter.Value,
		}
		return result
	}

	// Complex case: multiple filters or sub-groups
	var conditions []map[string]interface{}

	// Add filters
	for _, filter := range group.Filters {
		condition := map[string]interface{}{
			filter.Field: map[string]interface{}{
				"operator": operatorToMapKey(filter.Operator),
				"value":    filter.Value,
			},
		}
		conditions = append(conditions, condition)
	}

	// Add sub-groups
	for _, subGroup := range group.SubGroups {
		conditions = append(conditions, FilterToMap(subGroup))
	}

	// Combine with operator
	if len(conditions) > 0 {
		result[strings.ToLower(group.Operator)] = conditions
	}

	return result
}

// operatorToMapKey converts a filter operator to a map key
func operatorToMapKey(operator string) string {
	switch strings.ToLower(operator) {
	case "=", "==", "eq":
		return "equals"
	case "!=", "<>", "ne":
		return "notEquals"
	case ">", "gt":
		return "greaterThan"
	case ">=", "gte":
		return "greaterThanEqual"
	case "<", "lt":
		return "lessThan"
	case "<=", "lte":
		return "lessThanEqual"
	case "contains":
		return "contains"
	case "in":
		return "in"
	case "not_in":
		return "notIn"
	default:
		return "equals"
	}
}
