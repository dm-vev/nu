package embedding

import "strings"

// Filter represents a filter condition for document metadata.
// It defines a single condition to be applied on a specific field.
// Example: field="word_count", operator=">", value=10
type Filter struct {
	// Field is the metadata field to filter on
	Field string

	// Operator is the comparison operator
	// Supported operators: "=", "!=", ">", ">=", "<", "<=", "contains", "in", "not_in"
	Operator string

	// Value is the value to compare against
	Value interface{}
}

// FilterGroup represents a group of filters with a logical operator.
// It allows for complex nested conditions with AND/OR logic.
// Example: (word_count > 10 AND type = "article") OR (category IN ["news", "blog"])
type FilterGroup struct {
	// Filters is the list of filters in this group
	Filters []Filter

	// SubGroups is the list of sub-groups in this group
	SubGroups []FilterGroup

	// Operator is the logical operator to apply between filters
	// Supported operators: "and", "or"
	Operator string
}

// NewFilter creates a new metadata filter
func NewFilter(field, operator string, value interface{}) Filter {
	return Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// NewFilterGroup creates a new metadata filter group
func NewFilterGroup(operator string, filters ...Filter) FilterGroup {
	return FilterGroup{
		Filters:  filters,
		Operator: strings.ToLower(operator),
	}
}

// AddFilter adds a filter to the group
func (g *FilterGroup) AddFilter(filter Filter) {
	g.Filters = append(g.Filters, filter)
}

// AddSubGroup adds a sub-group to the group
func (g *FilterGroup) AddSubGroup(subGroup FilterGroup) {
	g.SubGroups = append(g.SubGroups, subGroup)
}
