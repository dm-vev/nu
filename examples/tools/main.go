package main

import (
	"context"
	"fmt"
	"log"

	"nu/internal/tools"
)

func main() {
	registry := tools.NewRegistry()
	registry.Register(tools.NewCalculator())

	tool, _ := registry.Get("calculator")
	result, err := tool.Execute(context.Background(), `{"expression":"6 * 7"}`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
