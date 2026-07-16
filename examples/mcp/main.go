package main

import (
	"fmt"
	"log"

	"nu/internal/mcp/builder"
)

func main() {
	configs, err := builder.NewBuilder().
		AddStdioServer("local-tools", "go", "run", "./cmd/mcp-server").
		AddHTTPServer("remote-tools", "https://mcp.example.test/v1").
		BuildLazy()
	if err != nil {
		log.Fatal(err)
	}
	for _, config := range configs {
		fmt.Println(config.Name, config.Type)
	}
}
