package relationship

import (
	"github.com/weaviate/weaviate-go-client/v5/weaviate/filters"

	"nu/internal/contracts"
)

// BuildIDFilter creates a Weaviate filter for relationship ID and optional tenant.
func BuildIDFilter(relationshipID, tenant string) *filters.WhereBuilder {
	relFilter := filters.Where().
		WithPath([]string{"relationshipId"}).
		WithOperator(filters.Equal).
		WithValueString(relationshipID)

	if tenant != "" {
		tenantFilter := filters.Where().
			WithPath([]string{"orgId"}).
			WithOperator(filters.Equal).
			WithValueString(tenant)

		return filters.Where().
			WithOperator(filters.And).
			WithOperands([]*filters.WhereBuilder{relFilter, tenantFilter})
	}

	return relFilter
}

// BuildDirectionFilter creates a filter for relationship queries based on direction.
func BuildDirectionFilter(
	entityID string,
	direction contracts.RelationshipDirection,
	tenant string,
	relTypes []string,
) *filters.WhereBuilder {
	var directionFilter *filters.WhereBuilder

	switch direction {
	case contracts.DirectionOutgoing:
		directionFilter = filters.Where().
			WithPath([]string{"sourceId"}).
			WithOperator(filters.Equal).
			WithValueString(entityID)
	case contracts.DirectionIncoming:
		directionFilter = filters.Where().
			WithPath([]string{"targetId"}).
			WithOperator(filters.Equal).
			WithValueString(entityID)
	default: // DirectionBoth
		sourceFilter := filters.Where().
			WithPath([]string{"sourceId"}).
			WithOperator(filters.Equal).
			WithValueString(entityID)
		targetFilter := filters.Where().
			WithPath([]string{"targetId"}).
			WithOperator(filters.Equal).
			WithValueString(entityID)
		directionFilter = filters.Where().
			WithOperator(filters.Or).
			WithOperands([]*filters.WhereBuilder{sourceFilter, targetFilter})
	}

	filterList := []*filters.WhereBuilder{directionFilter}

	// Add tenant filter if specified
	if tenant != "" {
		tenantFilter := filters.Where().
			WithPath([]string{"orgId"}).
			WithOperator(filters.Equal).
			WithValueString(tenant)
		filterList = append(filterList, tenantFilter)
	}

	// Add relationship type filter if specified
	if len(relTypes) > 0 {
		if len(relTypes) == 1 {
			typeFilter := filters.Where().
				WithPath([]string{"relationshipType"}).
				WithOperator(filters.Equal).
				WithValueString(relTypes[0])
			filterList = append(filterList, typeFilter)
		} else {
			typeFilters := make([]*filters.WhereBuilder, len(relTypes))
			for i, rt := range relTypes {
				typeFilters[i] = filters.Where().
					WithPath([]string{"relationshipType"}).
					WithOperator(filters.Equal).
					WithValueString(rt)
			}
			typeOrFilter := filters.Where().
				WithOperator(filters.Or).
				WithOperands(typeFilters)
			filterList = append(filterList, typeOrFilter)
		}
	}

	// Combine all filters with AND
	if len(filterList) == 1 {
		return filterList[0]
	}

	return filters.Where().
		WithOperator(filters.And).
		WithOperands(filterList)
}
