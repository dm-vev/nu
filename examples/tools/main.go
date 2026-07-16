package main

import (
	"context"
	"fmt"
	"log"

	"nu/internal/tools/calculator"
	"nu/internal/tools/registry"
)

func main() {
	toolRegistry := registry.NewRegistry()
	toolRegistry.Register(calculator.NewCalculator())

	tool, _ := toolRegistry.Get("calculator")
	result, err := tool.Execute(context.Background(), `{"expression":"6 * 7"}`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
