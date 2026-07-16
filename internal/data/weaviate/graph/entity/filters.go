package entity

import "github.com/weaviate/weaviate-go-client/v5/weaviate/filters"

// BuildIDFilter creates a Weaviate filter for entity ID and optional tenant.
func BuildIDFilter(entityID, tenant string) *filters.WhereBuilder {
	entityFilter := filters.Where().
		WithPath([]string{"entityId"}).
		WithOperator(filters.Equal).
		WithValueString(entityID)

	if tenant != "" {
		tenantFilter := filters.Where().
			WithPath([]string{"orgId"}).
			WithOperator(filters.Equal).
			WithValueString(tenant)

		return filters.Where().
			WithOperator(filters.And).
			WithOperands([]*filters.WhereBuilder{entityFilter, tenantFilter})
	}

	return entityFilter
}
