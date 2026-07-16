package generation

import (
	"context"
	"fmt"
	"strings"

	"github.com/dm-vev/nu/internal/llm/openai"
)

// IsAskingAboutRole detects questions about the agent's role or identity.
func (s *Service) IsAskingAboutRole(input string) bool {
	lowerInput := strings.ToLower(input)
	roleQueries := []string{
		"what are you", "who are you", "what is your role", "what do you do",
		"what can you do", "what is your purpose", "what is your function",
		"tell me about yourself", "introduce yourself", "what are your capabilities",
		"what are you designed to do", "what's your job", "what kind of assistant are you",
		"your role", "your expertise", "what are you expert in", "what are you specialized in",
		"your specialty", "what's your specialty",
	}
	for _, query := range roleQueries {
		if strings.Contains(lowerInput, query) {
			return true
		}
	}
	return false
}

// GenerateRoleResponse creates a response based on the agent's system prompt.
func (s *Service) GenerateRoleResponse() string {
	if s.SystemPrompt == "" || s.LLM == nil {
		return "I'm an AI assistant designed to help you with various tasks and answer your questions. How can I assist you today?"
	}

	agentName := "an AI assistant"
	if s.Name != "" {
		agentName = s.Name
	}
	prompt := fmt.Sprintf(`Based on the following system prompt that defines your role and capabilities,
generate a brief, natural-sounding response (3-5 sentences) introducing yourself to a user who asked what you can do.
You are named "%s".
Do not directly quote from the system prompt, but create a conversational first-person response that captures your
purpose, expertise, and how you can help. The response should feel like a natural conversation, not like reading documentation.

System prompt:
%s

Your response should:
1. Introduce yourself using first-person perspective, mentioning your name ("%s")
2. Briefly explain your specialization or purpose
3. Mention 2-3 key areas you can help with
4. End with a friendly question about how you can assist the user

Response:`, agentName, s.SystemPrompt, agentName)

	response, err := s.LLM.Generate(context.Background(), prompt, openai.WithSystemMessage(s.SystemPrompt))
	if err != nil {
		if s.Name != "" {
			return fmt.Sprintf("I'm %s, an AI assistant based on the role defined in my system prompt. How can I help you today?", s.Name)
		}
		return "I'm an AI assistant based on the role defined in my system prompt. How can I help you today?"
	}
	return response
}
