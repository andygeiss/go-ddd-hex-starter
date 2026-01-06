package agent

// AgentID represents a unique identifier for an agent.
type AgentID string

// TaskID represents a unique identifier for a task.
type TaskID string

// ToolCallID represents a unique identifier for a tool call.
type ToolCallID string

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// ToolCallStatus represents the status of a tool call.
type ToolCallStatus string

const (
	ToolCallStatusPending   ToolCallStatus = "pending"
	ToolCallStatusExecuting ToolCallStatus = "executing"
	ToolCallStatusCompleted ToolCallStatus = "completed"
	ToolCallStatusFailed    ToolCallStatus = "failed"
)

// Role represents the role of a message in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a message in a conversation with the LLM.
type Message struct {
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Role       Role       `json:"role"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// NewMessage creates a new message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
	}
}

// WithToolCalls sets the tool calls for the message.
func (m Message) WithToolCalls(toolCalls []ToolCall) Message {
	m.ToolCalls = toolCalls
	return m
}

// WithToolCallID sets the tool call ID for the message.
func (m Message) WithToolCallID(id string) Message {
	m.ToolCallID = id
	return m
}

// LLMResponse represents a response from the LLM.
type LLMResponse struct {
	Message      Message    `json:"message"`
	FinishReason string     `json:"finish_reason"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
}

// NewLLMResponse creates a new LLM response with the given message.
func NewLLMResponse(message Message, finishReason string) LLMResponse {
	return LLMResponse{
		Message:      message,
		FinishReason: finishReason,
	}
}

// WithToolCalls sets the tool calls for the response.
func (r LLMResponse) WithToolCalls(toolCalls []ToolCall) LLMResponse {
	r.ToolCalls = toolCalls
	return r
}

// HasToolCalls returns true if the response contains tool calls.
func (r LLMResponse) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}

// Result represents the final result of a task execution.
type Result struct {
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
	TaskID  TaskID `json:"task_id"`
	Success bool   `json:"success"`
}

// NewResult creates a new result for the given task.
func NewResult(taskID TaskID, success bool, output string) Result {
	return Result{
		TaskID:  taskID,
		Success: success,
		Output:  output,
	}
}

// WithError sets the error message for the result.
func (r Result) WithError(err string) Result {
	r.Error = err
	return r
}
