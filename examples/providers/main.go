package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"nu/internal/config"
	"nu/internal/llm/anthropic"
	"nu/internal/llm/azureopenai"
	"nu/internal/llm/deepseek"
	"nu/internal/llm/gemini"
	"nu/internal/llm/ollama"
	"nu/internal/llm/openai"
	"nu/internal/llm/vllm"
)

func main() {
	cfg := config.Get()
	fmt.Println(openai.NewClient(cfg.LLM.OpenAI.APIKey, openai.WithModel(cfg.LLM.OpenAI.Model)).Name())
	fmt.Println(anthropic.NewClient(cfg.LLM.Anthropic.APIKey, anthropic.WithModel(cfg.LLM.Anthropic.Model)).Name())

	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		client, err := gemini.NewClient(context.Background(), gemini.WithAPIKey(key))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(client.Name())
	}
	if azure := cfg.LLM.AzureOpenAI; azure.APIKey != "" && azure.BaseURL != "" && azure.Deployment != "" {
		fmt.Println(azureopenai.NewClient(azure.APIKey, azure.BaseURL, azure.Deployment).Name())
	}

	fmt.Println(deepseek.NewClient(os.Getenv("DEEPSEEK_API_KEY")).Name())
	fmt.Println(ollama.NewClient().Name())
	fmt.Println(vllm.NewClient().Name())
}
