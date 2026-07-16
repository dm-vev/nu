package mcp

import (
	"fmt"

	"github.com/dm-vev/nu/internal/mcp/builder"
)

// ConfigsFromURLs parses MCP server URLs into lazy configurations.
func ConfigsFromURLs(urls ...string) ([]LazyMCPConfig, error) {
	return buildConfigs(func(b *builder.Builder) {
		for _, url := range urls {
			b.AddServer(url)
		}
	})
}

// ConfigsFromPresets resolves named MCP presets into lazy configurations.
func ConfigsFromPresets(presetNames ...string) ([]LazyMCPConfig, error) {
	return buildConfigs(func(b *builder.Builder) {
		for _, name := range presetNames {
			b.AddPreset(name)
		}
	})
}

func buildConfigs(configure func(*builder.Builder)) ([]LazyMCPConfig, error) {
	b := builder.NewBuilder()
	configure(b)
	configs, err := b.BuildLazy()
	if err != nil {
		return nil, fmt.Errorf("build MCP configurations: %w", err)
	}

	result := make([]LazyMCPConfig, 0, len(configs))
	for _, config := range configs {
		result = append(result, LazyMCPConfig{
			Name:              config.Name,
			Type:              config.Type,
			Command:           config.Command,
			Args:              config.Args,
			Env:               config.Env,
			URL:               config.URL,
			Token:             config.Token,
			HttpTransportMode: config.HttpTransportMode,
			AllowedTools:      config.AllowedTools,
		})
	}
	return result, nil
}
