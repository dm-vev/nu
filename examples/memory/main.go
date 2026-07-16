package main

import (
	"context"
	"fmt"
	"log"

	"nu/internal/contracts"
	memory "nu/internal/memory/conversation"
	"nu/internal/multitenancy"
)

func main() {
	ctx := multitenancy.WithOrgID(context.Background(), "acme")
	ctx = memory.WithConversationID(ctx, "demo")
	buffer := memory.NewConversationBuffer()

	err := buffer.AddMessage(ctx, contracts.Message{Role: contracts.RoleUser, Content: "hello"})
	var messages []contracts.Message
	if err == nil {
		messages, err = buffer.GetMessages(ctx)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(messages[0].Content)
}
