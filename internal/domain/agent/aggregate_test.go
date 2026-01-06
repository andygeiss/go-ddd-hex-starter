package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
)

func Test_Agent_NewAgent_With_ValidParams_Should_Return_Agent(t *testing.T) {
	// Arrange
	id := agent.AgentID("test-agent")
	systemPrompt := "You are a helpful assistant."

	// Act
	a := agent.NewAgent(id, systemPrompt)

	// Assert
	assert.That(t, "agent ID must match", a.ID, id)
	assert.That(t, "system prompt must match", a.SystemPrompt, systemPrompt)
	assert.That(t, "tasks must be empty", len(a.Tasks), 0)
	assert.That(t, "messages must be empty", len(a.Messages), 0)
	assert.That(t, "max iterations must be default", a.MaxIterations, 10)
}

func Test_Agent_AddTask_With_OneTask_Should_HaveOneTask(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	task := agent.NewTask("task-1", "Test Task", "Do something")

	// Act
	a.AddTask(task)

	// Assert
	assert.That(t, "agent must have one task", len(a.Tasks), 1)
	assert.That(t, "task ID must match", a.Tasks[0].ID, task.ID)
}

func Test_Agent_GetCurrentTask_With_PendingTask_Should_ReturnTask(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	task := agent.NewTask("task-1", "Test Task", "Do something")
	a.AddTask(task)

	// Act
	current := a.GetCurrentTask()

	// Assert
	assert.That(t, "current task must not be nil", current != nil, true)
	assert.That(t, "current task ID must match", current.ID, task.ID)
}

func Test_Agent_GetCurrentTask_With_CompletedTask_Should_ReturnNil(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	task := agent.NewTask("task-1", "Test Task", "Do something")
	a.AddTask(task)
	a.Tasks[0].Complete("done")

	// Act
	current := a.GetCurrentTask()

	// Assert
	assert.That(t, "current task must be nil", current == nil, true)
}

func Test_Agent_AddMessage_With_OneMessage_Should_HaveOneMessage(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	msg := agent.NewMessage(agent.RoleUser, "Hello")

	// Act
	a.AddMessage(msg)

	// Assert
	assert.That(t, "agent must have one message", len(a.Messages), 1)
	assert.That(t, "message content must match", a.Messages[0].Content, "Hello")
}

func Test_Agent_CanContinue_With_AtMaxIterations_Should_ReturnFalse(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	a.WithMaxIterations(2)
	a.IncrementIteration()
	a.IncrementIteration()

	// Act
	canContinue := a.CanContinue()

	// Assert
	assert.That(t, "agent cannot continue", canContinue, false)
}

func Test_Agent_CanContinue_With_BelowMaxIterations_Should_ReturnTrue(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	a.WithMaxIterations(5)
	a.IncrementIteration()
	a.IncrementIteration()

	// Act
	canContinue := a.CanContinue()

	// Assert
	assert.That(t, "agent can continue", canContinue, true)
}

func Test_Agent_ClearMessages_With_Messages_Should_ClearAll(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	a.AddMessage(agent.NewMessage(agent.RoleUser, "msg1"))
	a.AddMessage(agent.NewMessage(agent.RoleUser, "msg2"))

	// Act
	a.ClearMessages()

	// Assert
	assert.That(t, "messages must be empty", len(a.Messages), 0)
}

func Test_Agent_GetMessages_With_SystemPrompt_Should_IncludeSystemMessage(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "You are helpful")
	a.AddMessage(agent.NewMessage(agent.RoleUser, "Hi"))

	// Act
	messages := a.GetMessages()

	// Assert
	assert.That(t, "messages must include system prompt", len(messages), 2)
	assert.That(t, "first message must be system", messages[0].Role, agent.RoleSystem)
	assert.That(t, "second message must be user", messages[1].Role, agent.RoleUser)
}

func Test_Agent_HasPendingTasks_With_PendingTask_Should_ReturnTrue(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	a.AddTask(agent.NewTask("task-1", "Task", "input"))

	// Act
	hasPending := a.HasPendingTasks()

	// Assert
	assert.That(t, "agent has pending tasks", hasPending, true)
}

func Test_Agent_HasPendingTasks_With_NoTasks_Should_ReturnFalse(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")

	// Act
	hasPending := a.HasPendingTasks()

	// Assert
	assert.That(t, "agent has no pending tasks", hasPending, false)
}

func Test_Agent_ResetIteration_With_Iterations_Should_ResetToZero(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")
	a.IncrementIteration()
	a.IncrementIteration()

	// Act
	a.ResetIteration()

	// Assert
	assert.That(t, "iteration must be zero", a.CurrentIteration, 0)
}

func Test_Agent_WithMaxIterations_With_CustomValue_Should_Update(t *testing.T) {
	// Arrange
	a := agent.NewAgent("test-agent", "prompt")

	// Act
	a.WithMaxIterations(5)

	// Assert
	assert.That(t, "max iterations must be updated", a.MaxIterations, 5)
}
