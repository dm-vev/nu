package schema

// SchemaBuilder helps create JSON schemas for tool outputs
type SchemaBuilder struct {
	schema map[string]interface{}
}

func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{schema: make(map[string]interface{})}
}

func (sb *SchemaBuilder) Object() *SchemaBuilder {
	sb.schema["type"] = "object"
	sb.schema["properties"] = make(map[string]interface{})
	return sb
}

func (sb *SchemaBuilder) Array(itemSchema map[string]interface{}) *SchemaBuilder {
	sb.schema["type"] = "array"
	sb.schema["items"] = itemSchema
	return sb
}

func (sb *SchemaBuilder) String() *SchemaBuilder {
	sb.schema["type"] = "string"
	return sb
}

func (sb *SchemaBuilder) Number() *SchemaBuilder {
	sb.schema["type"] = "number"
	return sb
}

func (sb *SchemaBuilder) Boolean() *SchemaBuilder {
	sb.schema["type"] = "boolean"
	return sb
}

func (sb *SchemaBuilder) Property(name string, propertySchema map[string]interface{}) *SchemaBuilder {
	if properties, exists := sb.schema["properties"]; exists {
		if propsMap, ok := properties.(map[string]interface{}); ok {
			propsMap[name] = propertySchema
		}
	}
	return sb
}

func (sb *SchemaBuilder) Required(fields ...string) *SchemaBuilder {
	required := make([]interface{}, len(fields))
	for i, field := range fields {
		required[i] = field
	}
	sb.schema["required"] = required
	return sb
}

func (sb *SchemaBuilder) Description(desc string) *SchemaBuilder {
	sb.schema["description"] = desc
	return sb
}

func (sb *SchemaBuilder) Build() map[string]interface{} { return sb.schema }

func CreateWeatherSchema() map[string]interface{} {
	return NewSchemaBuilder().
		Object().
		Property("temperature", map[string]interface{}{
			"type": "number", "description": "Temperature in celsius",
		}).
		Property("conditions", map[string]interface{}{
			"type": "string", "description": "Weather conditions description",
		}).
		Property("humidity", map[string]interface{}{
			"type": "number", "description": "Humidity percentage", "minimum": 0, "maximum": 100,
		}).
		Required("temperature", "conditions").
		Description("Weather information").
		Build()
}

func CreateFileInfoSchema() map[string]interface{} {
	return NewSchemaBuilder().
		Object().
		Property("name", map[string]interface{}{
			"type": "string", "description": "File name",
		}).
		Property("size", map[string]interface{}{
			"type": "integer", "description": "File size in bytes", "minimum": 0,
		}).
		Property("modified", map[string]interface{}{
			"type": "string", "description": "Last modified timestamp", "format": "date-time",
		}).
		Property("type", map[string]interface{}{
			"type": "string", "description": "File type or extension",
		}).
		Required("name", "size").
		Description("File information").
		Build()
}
