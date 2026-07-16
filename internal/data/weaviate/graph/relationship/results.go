package relationship

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

	id, _ := additional["id"].(string)
	return id
}

// ParseResults parses GraphQL response into Relationship slice.
func ParseResults(result *models.GraphQLResponse, className string) []contracts.Relationship {
	relationships := []contracts.Relationship{}

	if result.Data == nil {
		return relationships
	}

	getData, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		return relationships
	}

	classData, ok := getData[className].([]interface{})
	if !ok {
		return relationships
	}

	for _, item := range classData {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		rel := contracts.Relationship{}

		if v, ok := itemMap["relationshipId"].(string); ok {
			rel.ID = v
		}
		if v, ok := itemMap["sourceId"].(string); ok {
			rel.SourceID = v
		}
		if v, ok := itemMap["targetId"].(string); ok {
			rel.TargetID = v
		}
		if v, ok := itemMap["relationshipType"].(string); ok {
			rel.Type = v
		}
		if v, ok := itemMap["description"].(string); ok {
			rel.Description = v
		}
		if v, ok := itemMap["strength"].(float64); ok {
			rel.Strength = float32(v)
		}
		if v, ok := itemMap["orgId"].(string); ok {
			rel.OrgID = v
		}

		// Parse properties from JSON
		if propsStr, ok := itemMap["properties"].(string); ok && propsStr != "" {
			var props map[string]interface{}
			if err := json.Unmarshal([]byte(propsStr), &props); err == nil {
				rel.Properties = props
			}
		}

		// Parse timestamp
		if v, ok := itemMap["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				rel.CreatedAt = t
			}
		}

		relationships = append(relationships, rel)
	}

	return relationships
}
