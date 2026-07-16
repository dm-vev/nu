package entity

import (
	"encoding/json"
	"time"

	"github.com/weaviate/weaviate/entities/models"

	"nu/internal/contracts"
)

// ExtractUUID extracts the Weaviate UUID from a GraphQL response.
func ExtractUUID(result *models.GraphQLResponse, className string) string {
	if result.Data == nil {
		return ""
	}

	getData, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		return ""
	}

	classData, ok := getData[className].([]interface{})
	if !ok || len(classData) == 0 {
		return ""
	}

	firstItem, ok := classData[0].(map[string]interface{})
	if !ok {
		return ""
	}

	additional, ok := firstItem["_additional"].(map[string]interface{})
	if !ok {
		return ""
	}

	id, ok := additional["id"].(string)
	if !ok {
		return ""
	}

	return id
}

// ParseResults parses GraphQL response into Entity slice.
func ParseResults(result *models.GraphQLResponse, className string) []contracts.Entity {
	entities := []contracts.Entity{}

	if result.Data == nil {
		return entities
	}

	getData, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		return entities
	}

	classData, ok := getData[className].([]interface{})
	if !ok {
		return entities
	}

	for _, item := range classData {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		entity := contracts.Entity{}

		if v, ok := itemMap["entityId"].(string); ok {
			entity.ID = v
		}
		if v, ok := itemMap["name"].(string); ok {
			entity.Name = v
		}
		if v, ok := itemMap["entityType"].(string); ok {
			entity.Type = v
		}
		if v, ok := itemMap["description"].(string); ok {
			entity.Description = v
		}
		if v, ok := itemMap["orgId"].(string); ok {
			entity.OrgID = v
		}

		// Parse properties from JSON
		if propsStr, ok := itemMap["properties"].(string); ok && propsStr != "" {
			var props map[string]interface{}
			if err := json.Unmarshal([]byte(propsStr), &props); err == nil {
				entity.Properties = props
			}
		}

		// Parse timestamps
		if v, ok := itemMap["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				entity.CreatedAt = t
			}
		}
		if v, ok := itemMap["updatedAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				entity.UpdatedAt = t
			}
		}

		// Extract vector from _additional
		if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
			if vector, ok := additional["vector"].([]interface{}); ok {
				entity.Embedding = make([]float32, len(vector))
				for i, v := range vector {
					if f, ok := v.(float64); ok {
						entity.Embedding[i] = float32(f)
					}
				}
			}
		}

		entities = append(entities, entity)
	}

	return entities
}

// ParseMap parses an entity from a GraphQL result item.
func ParseMap(itemMap map[string]interface{}) contracts.Entity {
	entity := contracts.Entity{}

	if value, ok := itemMap["entityId"].(string); ok {
		entity.ID = value
	}
	if value, ok := itemMap["name"].(string); ok {
		entity.Name = value
	}
	if value, ok := itemMap["entityType"].(string); ok {
		entity.Type = value
	}
	if value, ok := itemMap["description"].(string); ok {
		entity.Description = value
	}
	if value, ok := itemMap["orgId"].(string); ok {
		entity.OrgID = value
	}

	return entity
}
