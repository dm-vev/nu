package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func extractTextFromMCPContent(content interface{}) string {
	switch c := content.(type) {
	case string:
		return c
	case []byte:
		return string(c)
	case []mcp.Content:
		var result string
		for i, item := range c {
			if i > 0 {
				result += "\n"
			}
			switch contentItem := item.(type) {
			case *mcp.TextContent:
				if contentItem.Text != "" {
					result += contentItem.Text
				} else {
					result += fmt.Sprintf("%+v", contentItem)
				}
			default:
				if bytes, err := json.Marshal(item); err == nil {
					result += string(bytes)
				} else {
					result += fmt.Sprintf("%v", item)
				}
			}
		}
		return result
	case []interface{}:
		var result string
		for i, item := range c {
			if i > 0 {
				result += "\n"
			}
			result += extractTextFromMCPContent(item)
		}
		return result
	case map[string]interface{}:
		if text, ok := c["text"].(string); ok {
			return text
		}
		if content, ok := c["content"].(string); ok {
			return content
		}
		if message, ok := c["message"].(string); ok {
			return message
		}
		if bytes, err := json.Marshal(c); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", c)
	default:
		if bytes, err := json.Marshal(content); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", content)
	}
}
