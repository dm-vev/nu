package main

import (
	"context"
	"fmt"
	"log"

	"nu/internal/agent"
	"nu/internal/config"
	"nu/internal/llm/openai"
	memory "nu/internal/memory/conversation"
	"nu/internal/multitenancy"
	"nu/internal/telemetry"
	"nu/internal/tools/registry"
	"nu/internal/tools/search"
)

func main() {
	cfg := config.Get()
	logger := telemetry.NewLogger()
	client := openai.NewClient(cfg.LLM.OpenAI.APIKey, openai.WithModel(cfg.LLM.OpenAI.Model), openai.WithLogger(logger))
	toolRegistry := registry.NewRegistry()
	if web := cfg.Tools.WebSearch; web.GoogleAPIKey != "" && web.GoogleSearchEngineID != "" {
		toolRegistry.Register(search.NewWebSearch(web.GoogleAPIKey, web.GoogleSearchEngineID))
	}
	conversation := memory.NewConversationBuffer()
	researcher, err := agent.NewAgent(
		agent.WithName("researcher"),
		agent.WithLLM(client),
		agent.WithMemory(conversation),
		agent.WithTools(toolRegistry.List()...),
		agent.WithSystemPrompt("Research the question and return a concise answer with sources when available."),
		agent.WithMaxIterations(5),
		agent.WithLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := multitenancy.WithOrgID(context.Background(), cfg.Multitenancy.DefaultOrgID)
	ctx = memory.WithConversationID(ctx, "example")
	answer, err := researcher.Run(ctx, "What is retrieval-augmented generation?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(answer)
}
