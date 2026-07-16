package generation

import (
	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/telemetry"
)

// Config contains the agent state required by response generation.
type Config struct {
	LLM                 contracts.LLM
	Memory              contracts.Memory
	Guardrails          contracts.Guardrails
	Logger              telemetry.Logger
	SystemPrompt        string
	Name                string
	ResponseFormat      *contracts.ResponseFormat
	LLMConfig           *contracts.LLMConfig
	MaxIterations       int
	DisableFinalSummary bool
	CacheConfig         *contracts.CacheConfig
	StreamConfig        *contracts.StreamConfig
}

// Service owns local LLM response and stream generation.
type Service struct {
	Config
}

// NewService creates a generation service from agent state.
func NewService(config Config) *Service {
	return &Service{Config: config}
}
