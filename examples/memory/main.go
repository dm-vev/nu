package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dm-vev/nu/contracts"
	"github.com/dm-vev/nu/internal/memory/conversation"
	"github.com/dm-vev/nu/internal/multitenancy"
)

func main() {
	ctx := multitenancy.WithOrgID(context.Background(), "acme")
	ctx = conversation.WithConversationID(ctx, "demo")
	buffer := conversation.NewConversationBuffer()

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
