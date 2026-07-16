package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// debugPrintConfig prints the agent config as YAML for debugging
func debugPrintConfig(config *AgentConfig, label string) {
	if config == nil {
		fmt.Printf("\n=== DEBUG: %s ===\nnil\n", label)
		return
	}

	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("\n=== DEBUG: %s (YAML marshal error: %v) ===\n", label, err)
		return
	}

	fmt.Printf("\n=== DEBUG: %s ===\n%s\n", label, string(yamlBytes))
}
