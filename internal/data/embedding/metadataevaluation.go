package embedding

import (
	"strings"

	"github.com/dm-vev/nu/contracts"
)

// ApplyFilters filters a list of documents based on metadata filters
func ApplyFilters(docs []contracts.Document, filterGroup FilterGroup) []contracts.Document {
	if len(filterGroup.Filters) == 0 && len(filterGroup.SubGroups) == 0 {
		return docs
	}

	var filtered []contracts.Document
	for _, doc := range docs {
		if evaluateFilterGroup(doc.Metadata, filterGroup) {
			filtered = append(filtered, doc)
		}
	}
	return filtered
}

// evaluateFilterGroup evaluates a filter group against document metadata
func evaluateFilterGroup(metadata map[string]interface{}, group FilterGroup) bool {
	if len(group.Filters) == 0 && len(group.SubGroups) == 0 {
		return true
	}

	// Evaluate individual filters
	filterResults := make([]bool, len(group.Filters))
	for i, filter := range group.Filters {
		filterResults[i] = evaluateFilter(metadata, filter)
	}

	// Evaluate sub-groups
	subGroupResults := make([]bool, len(group.SubGroups))
	for i, subGroup := range group.SubGroups {
		subGroupResults[i] = evaluateFilterGroup(metadata, subGroup)
	}

	// Combine results based on operator
	switch strings.ToLower(group.Operator) {
	case "or":
		// OR logic: return true if any filter or sub-group is true
		for _, result := range filterResults {
			if result {
				return true
			}
		}
		for _, result := range subGroupResults {
			if result {
				return true
			}
		}
		return false
	case "and", "": // Default to AND if not specified
		// AND logic: return false if any filter or sub-group is false
		for _, result := range filterResults {
			if !result {
				return false
			}
		}
		for _, result := range subGroupResults {
			if !result {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// evaluateFilter evaluates a single filter against document metadata
func evaluateFilter(metadata map[string]interface{}, filter Filter) bool {
	// Handle nested fields with dot notation (e.g., "user.name")
	if strings.Contains(filter.Field, ".") {
		return evaluateNestedFilter(metadata, filter)
	}

	value, exists := metadata[filter.Field]
	if !exists {
		// If the field doesn't exist, the filter doesn't match
		return false
	}

	switch strings.ToLower(filter.Operator) {
	case "=", "==", "eq":
		return equals(value, filter.Value)
	case "!=", "<>", "ne":
		return !equals(value, filter.Value)
	case ">", "gt":
		return compare(value, filter.Value) > 0
	case ">=", "gte":
		return compare(value, filter.Value) >= 0
	case "<", "lt":
		return compare(value, filter.Value) < 0
	case "<=", "lte":
		return compare(value, filter.Value) <= 0
	case "contains":
		return contains(value, filter.Value)
	case "in":
		return valueIn(value, filter.Value)
	case "not_in":
		return !valueIn(value, filter.Value)
	default:
		return false
	}
}

// evaluateNestedFilter handles filters with dot notation for nested fields
func evaluateNestedFilter(metadata map[string]interface{}, filter Filter) bool {
	parts := strings.Split(filter.Field, ".")
	current := metadata

	// Navigate through nested maps until the last part
	for i := 0; i < len(parts)-1; i++ {
		next, ok := current[parts[i]]
		if !ok {
			return false // Field path doesn't exist
		}

		// Check if the next level is a map
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			// Try to convert from map[interface{}]interface{} which might come from YAML/JSON
			if ifaceMap, isIfaceMap := next.(map[interface{}]interface{}); isIfaceMap {
				nextMap = make(map[string]interface{})
				for k, v := range ifaceMap {
					if kStr, ok := k.(string); ok {
						nextMap[kStr] = v
					}
				}
				current = nextMap
				continue
			}
			return false // Not a map, can't continue
		}

		current = nextMap
	}

	// Create a new filter for the final field
	lastField := parts[len(parts)-1]
	newFilter := Filter{
		Field:    lastField,
		Operator: filter.Operator,
		Value:    filter.Value,
	}

	// Evaluate the filter on the final nested map
	return evaluateFilter(current, newFilter)
}
