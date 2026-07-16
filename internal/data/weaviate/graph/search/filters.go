package search

import "github.com/weaviate/weaviate-go-client/v5/weaviate/filters"

// BuildFilter creates a Weaviate filter for search queries.
func BuildFilter(tenant string, entityTypes []string) *filters.WhereBuilder {
	filterList := []*filters.WhereBuilder{}

	// Add tenant filter
	if tenant != "" {
		tenantFilter := filters.Where().
			WithPath([]string{"orgId"}).
			WithOperator(filters.Equal).
			WithValueString(tenant)
		filterList = append(filterList, tenantFilter)
	}

	// Add entity type filter
	if len(entityTypes) > 0 {
		if len(entityTypes) == 1 {
			typeFilter := filters.Where().
				WithPath([]string{"entityType"}).
				WithOperator(filters.Equal).
				WithValueString(entityTypes[0])
			filterList = append(filterList, typeFilter)
		} else {
			typeFilters := make([]*filters.WhereBuilder, len(entityTypes))
			for i, et := range entityTypes {
				typeFilters[i] = filters.Where().
					WithPath([]string{"entityType"}).
					WithOperator(filters.Equal).
					WithValueString(et)
			}
			typeOrFilter := filters.Where().
				WithOperator(filters.Or).
				WithOperands(typeFilters)
			filterList = append(filterList, typeOrFilter)
		}
	}

	if len(filterList) == 0 {
		return nil
	}

	if len(filterList) == 1 {
		return filterList[0]
	}

	return filters.Where().
		WithOperator(filters.And).
		WithOperands(filterList)
}
