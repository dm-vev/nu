package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dm-vev/nu/agent"
	"github.com/dm-vev/nu/internal/config"
	"github.com/dm-vev/nu/internal/llm/openai"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
	"github.com/dm-vev/nu/internal/tools/registry"
	"github.com/dm-vev/nu/internal/tools/search"
	"github.com/dm-vev/nu/telemetry"
)

func main() {
	cfg := config.Get()
	logger := telemetry.NewLogger()
	client := openai.NewClient(cfg.LLM.OpenAI.APIKey, openai.WithModel(cfg.LLM.OpenAI.Model), openai.WithLogger(logger))
	toolRegistry := registry.NewRegistry()
	if web := cfg.Tools.WebSearch; web.GoogleAPIKey != "" && web.GoogleSearchEngineID != "" {
		toolRegistry.Register(search.NewWebSearch(web.GoogleAPIKey, web.GoogleSearchEngineID))
	}
	memory := conversation.NewConversationBuffer()
	researcher, err := agent.NewAgent(
		agent.WithName("researcher"),
		agent.WithLLM(client),
		agent.WithMemory(memory),
		agent.WithTools(toolRegistry.List()...),
		agent.WithSystemPrompt("Research the question and return a concise answer with sources when available."),
		agent.WithMaxIterations(5),
		agent.WithLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := multitenancy.WithOrgID(context.Background(), cfg.Multitenancy.DefaultOrgID)
	ctx = conversation.WithConversationID(ctx, "example")
	answer, err := researcher.Run(ctx, "What is retrieval-augmented generation?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(answer)
}
