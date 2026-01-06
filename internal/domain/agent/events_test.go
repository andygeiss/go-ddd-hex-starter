package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
)

func Test_EventTaskStarted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := agent.NewEventTaskStarted()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.task_started")
}

func Test_EventTaskStarted_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := agent.NewEventTaskStarted().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithName("Test Task")

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, agent.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, agent.TaskID("task-1"))
	assert.That(t, "name must match", evt.Name, "Test Task")
}

func Test_EventTaskCompleted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := agent.NewEventTaskCompleted()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.task_completed")
}

func Test_EventTaskCompleted_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := agent.NewEventTaskCompleted().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithName("Test Task").
		WithOutput("result").
		WithIterations(3)

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, agent.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, agent.TaskID("task-1"))
	assert.That(t, "name must match", evt.Name, "Test Task")
	assert.That(t, "output must match", evt.Output, "result")
	assert.That(t, "iterations must match", evt.Iterations, 3)
}

func Test_EventTaskFailed_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := agent.NewEventTaskFailed()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.task_failed")
}

func Test_EventTaskFailed_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := agent.NewEventTaskFailed().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithName("Test Task").
		WithError("something went wrong").
		WithIterations(5)

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, agent.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, agent.TaskID("task-1"))
	assert.That(t, "name must match", evt.Name, "Test Task")
	assert.That(t, "error must match", evt.Error, "something went wrong")
	assert.That(t, "iterations must match", evt.Iterations, 5)
}

func Test_EventToolCallExecuted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := agent.NewEventToolCallExecuted()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.tool_call_executed")
}

func Test_EventToolCallExecuted_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := agent.NewEventToolCallExecuted().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithToolCallID("tc-1").
		WithToolName("search").
		WithSuccess(true)

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, agent.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, agent.TaskID("task-1"))
	assert.That(t, "tool call ID must match", evt.ToolCallID, agent.ToolCallID("tc-1"))
	assert.That(t, "tool name must match", evt.ToolName, "search")
	assert.That(t, "success must match", evt.Success, true)
}

func Test_EventAgentLoopFinished_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := agent.NewEventAgentLoopFinished()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.loop_finished")
}

func Test_EventAgentLoopFinished_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := agent.NewEventAgentLoopFinished().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithTotalIterations(10).
		WithSuccess(true).
		WithReason("task completed")

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, agent.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, agent.TaskID("task-1"))
	assert.That(t, "total iterations must match", evt.TotalIterations, 10)
	assert.That(t, "success must match", evt.Success, true)
	assert.That(t, "reason must match", evt.Reason, "task completed")
}
