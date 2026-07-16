package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SaveAgentConfigsToFile saves agent configurations to a YAML file
func SaveAgentConfigsToFile(configs AgentConfigs, file *os.File) error {
	data, err := yaml.Marshal(configs)
	if err != nil {
		return fmt.Errorf("failed to marshal agent configs: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write agent configs to file: %w", err)
	}

	return nil
}

// SaveTaskConfigsToFile saves task configurations to a YAML file
func SaveTaskConfigsToFile(configs TaskConfigs, file *os.File) error {
	data, err := yaml.Marshal(configs)
	if err != nil {
		return fmt.Errorf("failed to marshal task configs: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write task configs to file: %w", err)
	}

	return nil
}
