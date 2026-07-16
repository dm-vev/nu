package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDeploymentAgentConfig_LocalOnlyNeedsNoDeploymentID(t *testing.T) {
	t.Setenv("AGENT_DEPLOYMENT_ID", "")
	path := filepath.Join(t.TempDir(), "agent.yaml")
	if err := os.WriteFile(path, []byte("worker:\n  role: local role\n  goal: local goal\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	config, err := LoadDeploymentAgentConfig(context.Background(), "worker", "test",
		WithDeploymentConfigLocalOnly(),
		WithDeploymentConfigLocalFallback(path),
		WithoutDeploymentConfigCache(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if config.Role != "local role" || config.ConfigSource.Type != "local" {
		t.Fatalf("unexpected local config: %#v", config)
	}
}

func TestMergeAgentConfig_RemotePriority(t *testing.T) {
	tests := []struct {
		name     string
		remote   *AgentConfig
		local    *AgentConfig
		expected *AgentConfig
	}{
		{
			name: "remote overrides non-empty local fields",
			remote: &AgentConfig{
				Role: "Senior Software Engineer",
				Goal: "Build scalable systems",
			},
			local: &AgentConfig{
				Role:      "Junior Developer",
				Goal:      "Learn basics",
				Backstory: "New to the team",
			},
			expected: &AgentConfig{
				Role:      "Senior Software Engineer",
				Goal:      "Build scalable systems",
				Backstory: "New to the team", // From local since remote is empty
			},
		},
		{
			name: "local fills gaps when remote has empty fields",
			remote: &AgentConfig{
				Role: "Data Analyst",
				// Goal is empty
			},
			local: &AgentConfig{
				Role:      "Should be overridden",
				Goal:      "Analyze data patterns",
				Backstory: "Expert in analytics",
			},
			expected: &AgentConfig{
				Role:      "Data Analyst",
				Goal:      "Analyze data patterns", // From local
				Backstory: "Expert in analytics",   // From local
			},
		},
		{
			name: "remote nil pointer fields use local values",
			remote: &AgentConfig{
				Role: "DevOps Engineer",
				// MaxIterations is nil
			},
			local: &AgentConfig{
				Role:          "Should be overridden",
				MaxIterations: intPtr(10),
			},
			expected: &AgentConfig{
				Role:          "DevOps Engineer",
				MaxIterations: intPtr(10), // From local
			},
		},
		{
			name: "merge tools - remote tools kept, local tools appended if not present",
			remote: &AgentConfig{
				Role: "Research Agent",
				Tools: []ToolConfigYAML{
					{Name: "web_search", Type: "search"},
				},
			},
			local: &AgentConfig{
				Role: "Should be overridden",
				Tools: []ToolConfigYAML{
					{Name: "web_search", Type: "search"}, // Duplicate - should not be added
					{Name: "calculator", Type: "math"},   // New tool - should be added
				},
			},
			expected: &AgentConfig{
				Role: "Research Agent",
				Tools: []ToolConfigYAML{
					{Name: "web_search", Type: "search"},
					{Name: "calculator", Type: "math"}, // Added from local
				},
			},
		},
		{
			name: "deep merge LLMProvider",
			remote: &AgentConfig{
				Role: "AI Assistant",
				LLMProvider: &LLMProviderYAML{
					Provider: "anthropic",
					// Model is empty
				},
			},
			local: &AgentConfig{
				Role: "Should be overridden",
				LLMProvider: &LLMProviderYAML{
					Provider: "openai", // Should be overridden
					Model:    "gpt-4",  // Should be used
				},
			},
			expected: &AgentConfig{
				Role: "AI Assistant",
				LLMProvider: &LLMProviderYAML{
					Provider: "anthropic",
					Model:    "gpt-4", // From local
				},
			},
		},
		{
			name: "recursive merge SubAgents",
			remote: &AgentConfig{
				Role: "Manager Agent",
				SubAgents: map[string]AgentConfig{
					"worker1": {
						Role: "Worker from remote",
						Goal: "Process tasks",
					},
				},
			},
			local: &AgentConfig{
				Role: "Should be overridden",
				SubAgents: map[string]AgentConfig{
					"worker1": {
						Role:      "Should be overridden",
						Goal:      "Should be overridden",
						Backstory: "Experienced worker", // Should be kept
					},
					"worker2": {
						Role: "Additional worker",
						Goal: "Handle overflow",
					},
				},
			},
			expected: &AgentConfig{
				Role: "Manager Agent",
				SubAgents: map[string]AgentConfig{
					"worker1": {
						Role:      "Worker from remote",
						Goal:      "Process tasks",
						Backstory: "Experienced worker", // From local
					},
					"worker2": {
						Role: "Additional worker", // Added from local
						Goal: "Handle overflow",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeDeploymentAgentConfig(tt.remote, tt.local, DeploymentConfigMergeRemotePriority)
			require.NotNil(t, result)

			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.Goal, result.Goal)
			assert.Equal(t, tt.expected.Backstory, result.Backstory)

			if tt.expected.MaxIterations != nil {
				require.NotNil(t, result.MaxIterations)
				assert.Equal(t, *tt.expected.MaxIterations, *result.MaxIterations)
			}

			if tt.expected.Tools != nil {
				require.NotNil(t, result.Tools)
				assert.Equal(t, len(tt.expected.Tools), len(result.Tools))
				for i, expectedTool := range tt.expected.Tools {
					assert.Equal(t, expectedTool.Name, result.Tools[i].Name)
					assert.Equal(t, expectedTool.Type, result.Tools[i].Type)
				}
			}

			if tt.expected.LLMProvider != nil {
				require.NotNil(t, result.LLMProvider)
				assert.Equal(t, tt.expected.LLMProvider.Provider, result.LLMProvider.Provider)
				assert.Equal(t, tt.expected.LLMProvider.Model, result.LLMProvider.Model)
			}

			if tt.expected.SubAgents != nil {
				require.NotNil(t, result.SubAgents)
				assert.Equal(t, len(tt.expected.SubAgents), len(result.SubAgents))
				for name, expectedSub := range tt.expected.SubAgents {
					resultSub, exists := result.SubAgents[name]
					require.True(t, exists, "SubAgent %s should exist", name)
					assert.Equal(t, expectedSub.Role, resultSub.Role)
					assert.Equal(t, expectedSub.Goal, resultSub.Goal)
					assert.Equal(t, expectedSub.Backstory, resultSub.Backstory)
				}
			}
		})
	}
}

func TestMergeAgentConfig_LocalPriority(t *testing.T) {
	local := &AgentConfig{
		Role:      "Local Role",
		Goal:      "Local Goal",
		Backstory: "",
	}

	remote := &AgentConfig{
		Role:      "Remote Role",
		Goal:      "",
		Backstory: "Remote Backstory",
	}

	result := MergeDeploymentAgentConfig(local, remote, DeploymentConfigMergeLocalPriority)
	require.NotNil(t, result)

	// Local values should take priority
	assert.Equal(t, "Local Role", result.Role)
	assert.Equal(t, "Local Goal", result.Goal)
	// Remote value used when local is empty
	assert.Equal(t, "Remote Backstory", result.Backstory)
}

func TestMergeAgentConfig_NilHandling(t *testing.T) {
	config := &AgentConfig{
		Role: "Test Role",
	}

	t.Run("nil remote returns local", func(t *testing.T) {
		result := MergeDeploymentAgentConfig(nil, config, DeploymentConfigMergeRemotePriority)
		assert.Equal(t, config, result)
	})

	t.Run("nil local returns remote", func(t *testing.T) {
		result := MergeDeploymentAgentConfig(config, nil, DeploymentConfigMergeRemotePriority)
		assert.Equal(t, config, result)
	})

	t.Run("both nil returns nil", func(t *testing.T) {
		result := MergeDeploymentAgentConfig(nil, nil, DeploymentConfigMergeRemotePriority)
		assert.Nil(t, result)
	})
}

func TestMergeAgentConfig_ConfigSourceMetadata(t *testing.T) {
	remote := &AgentConfig{
		Role: "Remote Role",
		ConfigSource: &ConfigSourceMetadata{
			Type:   "remote",
			Source: "config-server:8080",
			Variables: map[string]string{
				"API_KEY": "remote-key",
			},
		},
	}

	local := &AgentConfig{
		Role: "Local Role",
		ConfigSource: &ConfigSourceMetadata{
			Type:   "local",
			Source: "/path/to/local.yaml",
			Variables: map[string]string{
				"DB_HOST": "localhost",
			},
		},
	}

	result := MergeDeploymentAgentConfig(remote, local, DeploymentConfigMergeRemotePriority)
	require.NotNil(t, result)
	require.NotNil(t, result.ConfigSource)

	// Should be marked as merged
	assert.Equal(t, "merged", result.ConfigSource.Type)
	assert.Contains(t, result.ConfigSource.Source, "merged")
	assert.Contains(t, result.ConfigSource.Source, "config-server:8080")
	assert.Contains(t, result.ConfigSource.Source, "/path/to/local.yaml")

	// Variables should be merged
	assert.Equal(t, "remote-key", result.ConfigSource.Variables["API_KEY"])
	assert.Equal(t, "localhost", result.ConfigSource.Variables["DB_HOST"])
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

// TestMergeDoesNotMutateOriginal verifies that merging configs doesn't mutate the originals
func TestMergeDoesNotMutateOriginal(t *testing.T) {
	primary := &AgentConfig{
		Role: "Primary Role",
		Tools: []ToolConfigYAML{
			{Name: "tool1", Type: "type1"},
		},
		SubAgents: map[string]AgentConfig{
			"sub1": {Role: "Sub Role"},
		},
		ConfigSource: &ConfigSourceMetadata{
			Variables: map[string]string{
				"KEY1": "value1",
			},
		},
	}

	base := &AgentConfig{
		Role: "Base Role",
		Tools: []ToolConfigYAML{
			{Name: "tool2", Type: "type2"},
		},
		ConfigSource: &ConfigSourceMetadata{
			Variables: map[string]string{
				"KEY2": "value2",
			},
		},
	}

	// Keep original values
	originalPrimaryToolCount := len(primary.Tools)
	originalBaseToolCount := len(base.Tools)
	originalPrimaryRole := primary.Role
	originalBaseRole := base.Role
	originalPrimarySubAgentCount := len(primary.SubAgents)
	originalPrimaryVarCount := len(primary.ConfigSource.Variables)

	// Perform merge
	merged := MergeDeploymentAgentConfig(primary, base, DeploymentConfigMergeRemotePriority)

	// Modify merged config to test for shared state
	merged.Tools = append(merged.Tools, ToolConfigYAML{Name: "tool3", Type: "type3"})
	merged.SubAgents["sub2"] = AgentConfig{Role: "New Sub"}
	merged.ConfigSource.Variables["KEY3"] = "value3"
	merged.Role = "Modified Role"

	// Verify primary was not mutated
	assert.Equal(t, originalPrimaryToolCount, len(primary.Tools), "primary.Tools was mutated")
	assert.Equal(t, originalPrimaryRole, primary.Role, "primary.Role was mutated")
	assert.Equal(t, originalPrimarySubAgentCount, len(primary.SubAgents), "primary.SubAgents was mutated")
	assert.NotContains(t, primary.SubAgents, "sub2", "primary.SubAgents was mutated")
	assert.Equal(t, originalPrimaryVarCount, len(primary.ConfigSource.Variables), "primary.ConfigSource.Variables was mutated")
	assert.NotContains(t, primary.ConfigSource.Variables, "KEY3", "primary.ConfigSource.Variables was mutated")

	// Verify base was not mutated
	assert.Equal(t, originalBaseToolCount, len(base.Tools), "base.Tools was mutated")
	assert.Equal(t, originalBaseRole, base.Role, "base.Role was mutated")
}

// TestDeepCopyToolConfigs verifies that tool configs are deeply copied
func TestDeepCopyToolConfigs(t *testing.T) {
	primary := &AgentConfig{
		Role: "Primary",
		Tools: []ToolConfigYAML{
			{
				Name:   "tool1",
				Type:   "type1",
				Config: map[string]interface{}{"key": "value"},
			},
		},
	}

	base := &AgentConfig{
		Role: "Base",
		Tools: []ToolConfigYAML{
			{
				Name:   "tool2",
				Type:   "type2",
				Config: map[string]interface{}{"key2": "value2"},
			},
		},
	}

	merged := MergeDeploymentAgentConfig(primary, base, DeploymentConfigMergeRemotePriority)

	// Modify merged tool config
	merged.Tools[0].Config["new_key"] = "new_value"
	merged.Tools[1].Config["another_key"] = "another_value"

	// Verify originals not affected
	assert.NotContains(t, primary.Tools[0].Config, "new_key", "primary tool config was mutated")
	assert.NotContains(t, base.Tools[0].Config, "another_key", "base tool config was mutated")
}

// TestDeepCopyNestedMaps verifies that nested maps in configs are deeply copied
func TestDeepCopyNestedMaps(t *testing.T) {
	primary := &AgentConfig{
		Role: "Primary",
		LLMProvider: &LLMProviderYAML{
			Provider: "anthropic",
			Config: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	base := &AgentConfig{
		Role: "Base",
	}

	merged := MergeDeploymentAgentConfig(primary, base, DeploymentConfigMergeRemotePriority)

	// Modify nested map in merged config
	nestedMap := merged.LLMProvider.Config["nested"].(map[string]interface{})
	nestedMap["new_key"] = "new_value"

	// Verify original not affected
	originalNested := primary.LLMProvider.Config["nested"].(map[string]interface{})
	assert.NotContains(t, originalNested, "new_key", "primary nested config was mutated")
}

// TestDeepCopySubAgentsRecursive verifies recursive deep copying of sub-agents
func TestDeepCopySubAgentsRecursive(t *testing.T) {
	primary := &AgentConfig{
		Role: "Manager",
		SubAgents: map[string]AgentConfig{
			"worker1": {
				Role: "Worker",
				Tools: []ToolConfigYAML{
					{Name: "sub_tool", Type: "type1"},
				},
			},
		},
	}

	base := &AgentConfig{
		Role: "Base Manager",
	}

	merged := MergeDeploymentAgentConfig(primary, base, DeploymentConfigMergeRemotePriority)

	// Modify sub-agent in merged config
	merged.SubAgents["worker1"] = AgentConfig{Role: "Modified Worker"}

	// Verify original not affected
	assert.Equal(t, "Worker", primary.SubAgents["worker1"].Role, "primary sub-agent was mutated")
	assert.Len(t, primary.SubAgents["worker1"].Tools, 1, "primary sub-agent tools were mutated")
}

// TestCacheReturnsDeepCopy verifies that cache returns deep copies
func TestCacheReturnsDeepCopy(t *testing.T) {
	ClearDeploymentConfigCache()

	config := &AgentConfig{
		Role: "Test Role",
		Tools: []ToolConfigYAML{
			{Name: "tool1", Type: "type1"},
		},
		ConfigSource: &ConfigSourceMetadata{
			Variables: map[string]string{"KEY": "value"},
		},
	}

	// Cache the config
	cacheConfig("test-key", config, time.Hour)

	// Get from cache
	cached1 := getFromCache("test-key")
	cached2 := getFromCache("test-key")

	// Verify they're different instances
	require.NotNil(t, cached1)
	require.NotNil(t, cached2)

	// Modify one cached copy
	cached1.Tools = append(cached1.Tools, ToolConfigYAML{Name: "new_tool"})
	cached1.ConfigSource.Variables["NEW_KEY"] = "new_value"
	cached1.Role = "Modified"

	// Verify the other copy is unaffected
	assert.Len(t, cached2.Tools, 1, "cached2 tools were affected by cached1 modification")
	assert.Equal(t, "Test Role", cached2.Role, "cached2 role was affected")
	assert.NotContains(t, cached2.ConfigSource.Variables, "NEW_KEY", "cached2 variables were affected")

	// Verify original config is unaffected
	assert.Len(t, config.Tools, 1, "original config tools were affected")
	assert.Equal(t, "Test Role", config.Role, "original config role was affected")
}

// TestNilConfigMerge verifies proper handling of nil configs
func TestNilConfigMerge(t *testing.T) {
	config := &AgentConfig{
		Role: "Test Role",
		Tools: []ToolConfigYAML{
			{Name: "tool1"},
		},
	}

	t.Run("nil primary returns deep copy of base", func(t *testing.T) {
		result := MergeDeploymentAgentConfig(nil, config, DeploymentConfigMergeRemotePriority)
		require.NotNil(t, result)

		// Modify result
		result.Tools = append(result.Tools, ToolConfigYAML{Name: "new_tool"})

		// Verify original unaffected
		assert.Len(t, config.Tools, 1, "original config was mutated")
	})

	t.Run("nil base returns deep copy of primary", func(t *testing.T) {
		result := MergeDeploymentAgentConfig(config, nil, DeploymentConfigMergeRemotePriority)
		require.NotNil(t, result)

		// Modify result
		result.Tools = append(result.Tools, ToolConfigYAML{Name: "new_tool"})

		// Verify original unaffected
		assert.Len(t, config.Tools, 1, "original config was mutated")
	})
}

// TestDeepCopyComplexPointers verifies deep copying of complex pointer fields
func TestDeepCopyComplexPointers(t *testing.T) {
	primary := &AgentConfig{
		Role:                "Primary",
		MaxIterations:       intPtr(5),
		RequirePlanApproval: boolPtr(true),
		StreamConfig: &StreamConfigYAML{
			BufferSize:          intPtr(100),
			IncludeToolProgress: boolPtr(true),
		},
		LLMConfig: &LLMConfigYAML{
			Temperature: func() *float64 { v := 0.7; return &v }(),
			TopP:        func() *float64 { v := 0.9; return &v }(),
		},
	}

	base := &AgentConfig{
		Role: "Base",
	}

	merged := MergeDeploymentAgentConfig(primary, base, DeploymentConfigMergeRemotePriority)

	// Modify merged pointers
	*merged.MaxIterations = 10
	*merged.StreamConfig.BufferSize = 200
	*merged.LLMConfig.Temperature = 1.0

	// Verify originals unaffected
	assert.Equal(t, 5, *primary.MaxIterations, "primary.MaxIterations was mutated")
	assert.Equal(t, 100, *primary.StreamConfig.BufferSize, "primary.StreamConfig.BufferSize was mutated")
	assert.InDelta(t, 0.7, *primary.LLMConfig.Temperature, 0.001, "primary.LLMConfig.Temperature was mutated")
}
