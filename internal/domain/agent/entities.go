package agent

import "time"

// Task represents a task to be executed by the agent.
// It is an entity within the Agent aggregate.
type Task struct {
	ID          TaskID     `json:"id"`
	Description string     `json:"description"`
	Error       string     `json:"error,omitempty"`
	Input       string     `json:"input"`
	Name        string     `json:"name"`
	Output      string     `json:"output"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewTask creates a new task with the given ID, name, and input.
func NewTask(id TaskID, name string, input string) Task {
	now := time.Now()
	return Task{
		ID:        id,
		Input:     input,
		Name:      name,
		Status:    TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// WithDescription sets the description for the task.
func (t *Task) WithDescription(description string) *Task {
	t.Description = description
	t.UpdatedAt = time.Now()
	return t
}

// Start marks the task as in progress.
func (t *Task) Start() {
	t.Status = TaskStatusInProgress
	t.UpdatedAt = time.Now()
}

// Complete marks the task as completed with the given output.
func (t *Task) Complete(output string) {
	t.Status = TaskStatusCompleted
	t.Output = output
	t.UpdatedAt = time.Now()
}

// Fail marks the task as failed with the given error.
func (t *Task) Fail(err string) {
	t.Status = TaskStatusFailed
	t.Error = err
	t.UpdatedAt = time.Now()
}

// IsCompleted returns true if the task is completed.
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// IsFailed returns true if the task has failed.
func (t *Task) IsFailed() bool {
	return t.Status == TaskStatusFailed
}

// IsTerminal returns true if the task is in a terminal state (completed or failed).
func (t *Task) IsTerminal() bool {
	return t.IsCompleted() || t.IsFailed()
}

// ToolCall represents a tool call requested by the LLM.
// It is an entity within the Agent aggregate.
type ToolCall struct {
	ID        ToolCallID     `json:"id"`
	Arguments string         `json:"arguments"`
	Error     string         `json:"error,omitempty"`
	Name      string         `json:"name"`
	Result    string         `json:"result"`
	Status    ToolCallStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// NewToolCall creates a new tool call with the given ID, name, and arguments.
func NewToolCall(id ToolCallID, name string, arguments string) ToolCall {
	now := time.Now()
	return ToolCall{
		ID:        id,
		Arguments: arguments,
		Name:      name,
		Status:    ToolCallStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Complete marks the tool call as completed with the given result.
func (tc *ToolCall) Complete(result string) {
	tc.Status = ToolCallStatusCompleted
	tc.Result = result
	tc.UpdatedAt = time.Now()
}

// Execute marks the tool call as executing.
func (tc *ToolCall) Execute() {
	tc.Status = ToolCallStatusExecuting
	tc.UpdatedAt = time.Now()
}

// Fail marks the tool call as failed with the given error.
func (tc *ToolCall) Fail(err string) {
	tc.Status = ToolCallStatusFailed
	tc.Error = err
	tc.UpdatedAt = time.Now()
}

// IsCompleted returns true if the tool call is completed.
func (tc *ToolCall) IsCompleted() bool {
	return tc.Status == ToolCallStatusCompleted
}

// IsFailed returns true if the tool call has failed.
func (tc *ToolCall) IsFailed() bool {
	return tc.Status == ToolCallStatusFailed
}

// ToMessage converts the tool call result to a message for the LLM.
func (tc *ToolCall) ToMessage() Message {
	content := tc.Result
	if tc.IsFailed() {
		content = "Error: " + tc.Error
	}
	return NewMessage(RoleTool, content).WithToolCallID(string(tc.ID))
}
