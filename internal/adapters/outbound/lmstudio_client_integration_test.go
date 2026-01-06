//go:build integration

package outbound_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
)

// Integration tests for LMStudioClient against a real LM Studio server.
// These tests require LM Studio to be running locally.
//
// Run with: go test -tags=integration ./internal/adapters/outbound/...
// Or use: just test-integration
//
// Environment variables:
// - LM_STUDIO_URL: Base URL for LM Studio (default: http://localhost:1234)
// - LM_STUDIO_MODEL: Model to use (default: default)

func getIntegrationEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Test_Integration_LMStudioClient_Run_With_SimplePrompt_Should_ReturnResponse(t *testing.T) {
	// Arrange
	baseURL := getIntegrationEnvOrDefault("LM_STUDIO_URL", "http://localhost:1234")
	model := getIntegrationEnvOrDefault("LM_STUDIO_MODEL", "default")

	sut := outbound.NewLMStudioClient(baseURL, model)
	messages := []agent.Message{
		agent.NewMessage(agent.RoleSystem, "You are a helpful assistant. Respond briefly."),
		agent.NewMessage(agent.RoleUser, "Say 'Hello, World!' and nothing else."),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Act
	response, err := sut.Run(ctx, messages)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "response message role must be assistant", response.Message.Role, agent.RoleAssistant)
	assert.That(t, "response message content must not be empty", len(response.Message.Content) > 0, true)
	t.Logf("LLM Response: %s", response.Message.Content)
}

func Test_Integration_LMStudioClient_Run_With_ConversationHistory_Should_MaintainContext(t *testing.T) {
	// Arrange
	baseURL := getIntegrationEnvOrDefault("LM_STUDIO_URL", "http://localhost:1234")
	model := getIntegrationEnvOrDefault("LM_STUDIO_MODEL", "default")

	sut := outbound.NewLMStudioClient(baseURL, model)
	messages := []agent.Message{
		agent.NewMessage(agent.RoleSystem, "You are a helpful assistant. Remember the context of our conversation."),
		agent.NewMessage(agent.RoleUser, "My name is Alice."),
		agent.NewMessage(agent.RoleAssistant, "Nice to meet you, Alice! How can I help you today?"),
		agent.NewMessage(agent.RoleUser, "What is my name?"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Act
	response, err := sut.Run(ctx, messages)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "response message role must be assistant", response.Message.Role, agent.RoleAssistant)
	assert.That(t, "response message content must not be empty", len(response.Message.Content) > 0, true)
	t.Logf("LLM Response: %s", response.Message.Content)
}

func Test_Integration_LMStudioClient_Run_With_CodeQuestion_Should_ReturnCodeResponse(t *testing.T) {
	// Arrange
	baseURL := getIntegrationEnvOrDefault("LM_STUDIO_URL", "http://localhost:1234")
	model := getIntegrationEnvOrDefault("LM_STUDIO_MODEL", "default")

	sut := outbound.NewLMStudioClient(baseURL, model)
	messages := []agent.Message{
		agent.NewMessage(agent.RoleSystem, "You are a Go programming expert. Provide concise code examples."),
		agent.NewMessage(agent.RoleUser, "Write a Go function that adds two integers and returns the result. Only show the function, no explanation."),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Act
	response, err := sut.Run(ctx, messages)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "response message role must be assistant", response.Message.Role, agent.RoleAssistant)
	assert.That(t, "response message content must not be empty", len(response.Message.Content) > 0, true)
	t.Logf("LLM Response:\n%s", response.Message.Content)
}

func Test_Integration_LMStudioClient_Run_With_LongPrompt_Should_HandleGracefully(t *testing.T) {
	// Arrange
	baseURL := getIntegrationEnvOrDefault("LM_STUDIO_URL", "http://localhost:1234")
	model := getIntegrationEnvOrDefault("LM_STUDIO_MODEL", "default")

	sut := outbound.NewLMStudioClient(baseURL, model)

	// Create a moderately long prompt
	longContext := "You are analyzing a Go project structure. Here are the files: " +
		"internal/domain/agent/aggregate.go (Agent aggregate root), " +
		"internal/domain/agent/entities.go (Task and ToolCall entities), " +
		"internal/domain/agent/value_objects.go (Value objects for the agent domain), " +
		"internal/domain/agent/service.go (TaskService with agent loop), " +
		"internal/adapters/outbound/lmstudio_client.go (LM Studio adapter)."

	messages := []agent.Message{
		agent.NewMessage(agent.RoleSystem, longContext),
		agent.NewMessage(agent.RoleUser, "Summarize this project structure in one sentence."),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Act
	response, err := sut.Run(ctx, messages)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "response message role must be assistant", response.Message.Role, agent.RoleAssistant)
	assert.That(t, "response message content must not be empty", len(response.Message.Content) > 0, true)
	t.Logf("LLM Response: %s", response.Message.Content)
}
