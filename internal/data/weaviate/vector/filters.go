package vector

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate-go-client/v5/weaviate/filters"
)

func (s *Store) buildWhereFilter(filterMap map[string]interface{}) *filters.WhereBuilder {
	if len(filterMap) == 0 {
		return nil
	}

	// Check for operands
	operandsIface, hasOperands := filterMap["operands"]
	if hasOperands {
		operator, hasOperator := filterMap["operator"]
		if !hasOperator {
			s.logger.Info(context.Background(), "Warning: Filter with operands missing operator", map[string]interface{}{"filter": filterMap})
			return nil
		}

		// Convert operands to a slice of filters
		operandsSlice, ok := operandsIface.([]interface{})
		if !ok {
			s.logger.Info(context.Background(), "Warning: Operands is not a slice", map[string]interface{}{"operands": operandsIface})
			return nil
		}

		// Build operands
		var whereOperands []*filters.WhereBuilder
		for _, operand := range operandsSlice {
			if subFilter := s.buildWhereFilter(operand.(map[string]interface{})); subFilter != nil {
				whereOperands = append(whereOperands, subFilter)
			}
		}

		// Create filter with operands
		if len(whereOperands) > 0 {
			switch operator {
			case "And":
				return filters.Where().WithOperator(filters.And).WithOperands(whereOperands)
			case "Or":
				return filters.Where().WithOperator(filters.Or).WithOperands(whereOperands)
			default:
				s.logger.Info(context.Background(), "Warning: Unsupported operator in filter with operands", map[string]interface{}{"operator": operator})
				return nil
			}
		}
		return nil
	}

	// Direct filter
	if len(filterMap) > 0 {
		operator, hasOperator := filterMap["operator"]
		if !hasOperator {
			s.logger.Info(context.Background(), "Warning: Direct filter missing operator", map[string]interface{}{"filter": filterMap})
			return nil
		}

		// Create the filter
		condition := filters.Where()

		// Handle path
		if pathSlice, ok := filterMap["path"].([]string); ok {
			condition = condition.WithPath(pathSlice)
		} else if pathStr, ok := filterMap["path"].(string); ok {
			condition = condition.WithPath([]string{pathStr})
		} else if pathIface, ok := filterMap["path"].([]interface{}); ok {
			pathSlice := make([]string, len(pathIface))
			for i, p := range pathIface {
				pathSlice[i] = fmt.Sprint(p)
			}
			condition = condition.WithPath(pathSlice)
		}

		// Handle operator and value
		switch operator {
		case "Equal":
			if val, ok := filterMap["valueString"]; ok {
				return condition.WithOperator(filters.Equal).WithValueString(fmt.Sprint(val))
			}
		case "NotEqual":
			if val, ok := filterMap["valueString"]; ok {
				return condition.WithOperator(filters.NotEqual).WithValueString(fmt.Sprint(val))
			}
		case "GreaterThan":
			if val, ok := filterMap["valueNumber"]; ok {
				return condition.WithOperator(filters.GreaterThan).WithValueNumber(vectorToFloat64(val))
			}
		case "GreaterThanEqual":
			if val, ok := filterMap["valueNumber"]; ok {
				return condition.WithOperator(filters.GreaterThanEqual).WithValueNumber(vectorToFloat64(val))
			}
		case "LessThan":
			if val, ok := filterMap["valueNumber"]; ok {
				return condition.WithOperator(filters.LessThan).WithValueNumber(vectorToFloat64(val))
			}
		case "LessThanEqual":
			if val, ok := filterMap["valueNumber"]; ok {
				return condition.WithOperator(filters.LessThanEqual).WithValueNumber(vectorToFloat64(val))
			}
		case "Like":
			if val, ok := filterMap["valueString"]; ok {
				return condition.WithOperator(filters.Like).WithValueString(fmt.Sprint(val))
			}
		case "ContainsAny":
			if val, ok := filterMap["valueString"]; ok {
				if strSlice, ok := val.([]string); ok {
					return condition.WithOperator(filters.ContainsAny).WithValueString(strSlice...)
				} else if strIface, ok := val.([]interface{}); ok {
					strSlice := make([]string, len(strIface))
					for i, s := range strIface {
						strSlice[i] = fmt.Sprint(s)
					}
					return condition.WithOperator(filters.ContainsAny).WithValueString(strSlice...)
				} else {
					return condition.WithOperator(filters.ContainsAny).WithValueString(fmt.Sprint(val))
				}
			}
		}

		s.logger.Info(context.Background(), "Warning: Could not build direct filter", map[string]interface{}{"filter": filterMap})
		return nil
	}

	// Check for logical operators (and, or)
	if andConditions, ok := filterMap["and"].([]interface{}); ok {
		// Create conditions for each operand
		var operands []*filters.WhereBuilder

		// Process each condition in the AND array
		for _, condition := range andConditions {
			// Check if this is a direct Weaviate filter
			if condMap, ok := condition.(map[string]interface{}); ok {
				if _, hasPath := condMap["path"]; hasPath {
					// This is a direct Weaviate filter
					if subFilter := s.buildWhereFilter(condMap); subFilter != nil {
						operands = append(operands, subFilter)
					}
					continue
				}

				// Otherwise, process as our custom filter format
				for field, value := range condMap {
					if valueMap, ok := value.(map[string]interface{}); ok {
						// Get operator and value from the map
						operator := valueMap["operator"].(string)
						val := valueMap["value"]

						// Create a condition for this field
						condition := filters.Where().
							WithPath([]string{field})

						// Apply the appropriate operator
						switch operator {
						case "equals":
							condition = condition.WithOperator(filters.Equal).WithValueString(fmt.Sprint(val))
						case "notEquals":
							condition = condition.WithOperator(filters.NotEqual).WithValueString(fmt.Sprint(val))
						case "greaterThan":
							condition = condition.WithOperator(filters.GreaterThan).WithValueNumber(vectorToFloat64(val))
						case "greaterThanEqual":
							condition = condition.WithOperator(filters.GreaterThanEqual).WithValueNumber(vectorToFloat64(val))
						case "lessThan":
							condition = condition.WithOperator(filters.LessThan).WithValueNumber(vectorToFloat64(val))
						case "lessThanEqual":
							condition = condition.WithOperator(filters.LessThanEqual).WithValueNumber(vectorToFloat64(val))
						case "like", "contains":
							condition = condition.WithOperator(filters.Like).WithValueString(fmt.Sprint(val))
						case "in":
							// Handle 'in' operator if supported by your Weaviate version
							if values, ok := val.([]interface{}); ok {
								strValues := make([]string, len(values))
								for i, v := range values {
									strValues[i] = fmt.Sprint(v)
								}
								// Use the correct method for ContainsAny operator
								condition = condition.WithOperator(filters.ContainsAny).WithValueString(strValues...)
							}
						}

						// Add this condition to the operands
						operands = append(operands, condition)
					}
				}
			}
		}

		// Create the AND group with all operands
		if len(operands) > 0 {
			return filters.Where().WithOperator(filters.And).WithOperands(operands)
		}
		return nil
	} else if orConditions, ok := filterMap["or"].([]interface{}); ok {
		// Create conditions for each operand
		var operands []*filters.WhereBuilder

		// Process each condition in the OR array
		for _, condition := range orConditions {
			// Check if this is a direct Weaviate filter
			if condMap, ok := condition.(map[string]interface{}); ok {
				if _, hasPath := condMap["path"]; hasPath {
					// This is a direct Weaviate filter
					if subFilter := s.buildWhereFilter(condMap); subFilter != nil {
						operands = append(operands, subFilter)
					}
					continue
				}

				// Otherwise, process as our custom filter format
				for field, value := range condMap {
					if valueMap, ok := value.(map[string]interface{}); ok {
						// Get operator and value from the map
						operator := valueMap["operator"].(string)
						val := valueMap["value"]

						// Create a condition for this field
						condition := filters.Where().
							WithPath([]string{field})

						// Apply the appropriate operator
						switch operator {
						case "equals":
							condition = condition.WithOperator(filters.Equal).WithValueString(fmt.Sprint(val))
						case "notEquals":
							condition = condition.WithOperator(filters.NotEqual).WithValueString(fmt.Sprint(val))
						case "greaterThan":
							condition = condition.WithOperator(filters.GreaterThan).WithValueNumber(vectorToFloat64(val))
						case "greaterThanEqual":
							condition = condition.WithOperator(filters.GreaterThanEqual).WithValueNumber(vectorToFloat64(val))
						case "lessThan":
							condition = condition.WithOperator(filters.LessThan).WithValueNumber(vectorToFloat64(val))
						case "lessThanEqual":
							condition = condition.WithOperator(filters.LessThanEqual).WithValueNumber(vectorToFloat64(val))
						case "like", "contains":
							condition = condition.WithOperator(filters.Like).WithValueString(fmt.Sprint(val))
						}

						// Add this condition to the operands
						operands = append(operands, condition)
					}
				}
			}
		}

		// Create the OR group with all operands
		if len(operands) > 0 {
			return filters.Where().WithOperator(filters.Or).WithOperands(operands)
		}
		return nil
	} else {
		// Handle simple key-value filters
		for field, value := range filterMap {
			if valueMap, ok := value.(map[string]interface{}); ok {
				operator := valueMap["operator"].(string)
				val := valueMap["value"]

				where := filters.Where().WithPath([]string{field})

				switch operator {
				case "equals":
					return where.WithOperator(filters.Equal).WithValueString(fmt.Sprint(val))
				case "notEquals":
					return where.WithOperator(filters.NotEqual).WithValueString(fmt.Sprint(val))
				case "greaterThan":
					return where.WithOperator(filters.GreaterThan).WithValueNumber(vectorToFloat64(val))
				case "greaterThanEqual":
					return where.WithOperator(filters.GreaterThanEqual).WithValueNumber(vectorToFloat64(val))
				case "lessThan":
					return where.WithOperator(filters.LessThan).WithValueNumber(vectorToFloat64(val))
				case "lessThanEqual":
					return where.WithOperator(filters.LessThanEqual).WithValueNumber(vectorToFloat64(val))
				case "like", "contains":
					return where.WithOperator(filters.Like).WithValueString(fmt.Sprint(val))
				}
			} else {
				// Simple equality
				return filters.Where().
					WithPath([]string{field}).
					WithOperator(filters.Equal).
					WithValueString(fmt.Sprint(value))
			}
		}
	}
	return nil
}
