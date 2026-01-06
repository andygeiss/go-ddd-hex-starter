package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
)

func Test_Message_NewMessage_With_ValidParams_Should_ReturnMessage(t *testing.T) {
	// Arrange
	role := agent.RoleUser
	content := "Hello, assistant!"

	// Act
	msg := agent.NewMessage(role, content)

	// Assert
	assert.That(t, "message role must match", msg.Role, role)
	assert.That(t, "message content must match", msg.Content, content)
}

func Test_Message_WithToolCalls_With_ToolCalls_Should_HaveToolCalls(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleAssistant, "content")
	toolCalls := []agent.ToolCall{
		agent.NewToolCall("tc-1", "search", `{}`),
	}

	// Act
	msg = msg.WithToolCalls(toolCalls)

	// Assert
	assert.That(t, "message must have tool calls", len(msg.ToolCalls), 1)
	assert.That(t, "tool call name must match", msg.ToolCalls[0].Name, "search")
}

func Test_Message_WithToolCallID_With_ID_Should_HaveToolCallID(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleTool, "result")

	// Act
	msg = msg.WithToolCallID("tc-1")

	// Assert
	assert.That(t, "message tool call ID must match", msg.ToolCallID, "tc-1")
}

func Test_LLMResponse_NewLLMResponse_With_ValidParams_Should_ReturnResponse(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleAssistant, "Hello!")
	finishReason := "stop"

	// Act
	resp := agent.NewLLMResponse(msg, finishReason)

	// Assert
	assert.That(t, "response message content must match", resp.Message.Content, "Hello!")
	assert.That(t, "response finish reason must match", resp.FinishReason, "stop")
}

func Test_LLMResponse_HasToolCalls_With_NoToolCalls_Should_ReturnFalse(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleAssistant, "Hello!")
	resp := agent.NewLLMResponse(msg, "stop")

	// Act
	hasToolCalls := resp.HasToolCalls()

	// Assert
	assert.That(t, "response must not have tool calls", hasToolCalls, false)
}

func Test_LLMResponse_HasToolCalls_With_ToolCalls_Should_ReturnTrue(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleAssistant, "")
	toolCalls := []agent.ToolCall{
		agent.NewToolCall("tc-1", "search", `{}`),
	}
	resp := agent.NewLLMResponse(msg, "tool_calls").WithToolCalls(toolCalls)

	// Act
	hasToolCalls := resp.HasToolCalls()

	// Assert
	assert.That(t, "response must have tool calls", hasToolCalls, true)
}

func Test_Result_NewResult_With_Success_Should_ReturnSuccessResult(t *testing.T) {
	// Arrange
	taskID := agent.TaskID("task-1")
	output := "Task completed successfully"

	// Act
	result := agent.NewResult(taskID, true, output)

	// Assert
	assert.That(t, "result task ID must match", result.TaskID, taskID)
	assert.That(t, "result success must be true", result.Success, true)
	assert.That(t, "result output must match", result.Output, output)
}

func Test_Result_WithError_With_ErrorMessage_Should_HaveError(t *testing.T) {
	// Arrange
	result := agent.NewResult("task-1", false, "")

	// Act
	result = result.WithError("something failed")

	// Assert
	assert.That(t, "result error must match", result.Error, "something failed")
}

func Test_ToolDefinition_NewToolDefinition_With_ValidParams_Should_ReturnDefinition(t *testing.T) {
	// Arrange
	name := "search"
	description := "Search the web"

	// Act
	td := agent.NewToolDefinition(name, description)

	// Assert
	assert.That(t, "tool definition name must match", td.Name, name)
	assert.That(t, "tool definition description must match", td.Description, description)
	assert.That(t, "tool definition parameters must be empty", len(td.Parameters), 0)
}

func Test_ToolDefinition_WithParameter_With_Params_Should_HaveParameters(t *testing.T) {
	// Arrange
	td := agent.NewToolDefinition("search", "Search the web")

	// Act
	td = td.WithParameter("query", "The search query")

	// Assert
	assert.That(t, "tool definition must have one parameter", len(td.Parameters), 1)
	assert.That(t, "parameter description must match", td.Parameters["query"], "The search query")
}
