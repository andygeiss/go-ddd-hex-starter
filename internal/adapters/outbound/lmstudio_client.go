package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
)

// The adapter translates between domain types (agent.Message, agent.ToolCall)
// and the OpenAI chat payload that LM Studio expects.

// LMStudioClient implements the agent.LLMClient interface.
// It communicates with LM Studio using the OpenAI-compatible API.
type LMStudioClient struct {
	httpClient *http.Client
	baseURL    string
	model      string
}

// NewLMStudioClient creates a new LMStudioClient instance.
func NewLMStudioClient(baseURL, model string) *LMStudioClient {
	return &LMStudioClient{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{},
	}
}

// WithHTTPClient sets a custom HTTP client for the LMStudioClient.
func (c *LMStudioClient) WithHTTPClient(httpClient *http.Client) *LMStudioClient {
	c.httpClient = httpClient
	return c
}

// Run sends the conversation messages to LM Studio and returns the response.
// It translates between domain types and the OpenAI-compatible API format.
func (c *LMStudioClient) Run(ctx context.Context, messages []agent.Message) (agent.LLMResponse, error) {
	apiMessages := c.convertToAPIMessages(messages)

	respPayload, err := c.sendRequest(ctx, apiMessages)
	if err != nil {
		return agent.LLMResponse{}, err
	}

	return c.convertToResponse(respPayload)
}

// convertToAPIMessages converts domain messages to API format.
func (c *LMStudioClient) convertToAPIMessages(messages []agent.Message) []chatMessage {
	apiMessages := make([]chatMessage, len(messages))
	for i, msg := range messages {
		apiMessages[i] = chatMessage{
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		if len(msg.ToolCalls) > 0 {
			apiMessages[i].ToolCalls = make([]apiToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				apiMessages[i].ToolCalls[j] = apiToolCall{
					ID:   string(tc.ID),
					Type: "function",
					Function: apiFunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
		}
	}
	return apiMessages
}

// sendRequest sends the chat completion request to LM Studio.
func (c *LMStudioClient) sendRequest(ctx context.Context, apiMessages []chatMessage) (*chatCompletionResponse, error) {
	reqPayload := chatCompletionRequest{
		Model:    c.model,
		Messages: apiMessages,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LM Studio returned status %d: %s", resp.StatusCode, string(body))
	}

	var respPayload chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &respPayload, nil
}

// convertToResponse converts the API response to domain types.
func (c *LMStudioClient) convertToResponse(respPayload *chatCompletionResponse) (agent.LLMResponse, error) {
	if len(respPayload.Choices) == 0 {
		return agent.LLMResponse{}, errors.New("no choices in response")
	}

	choice := respPayload.Choices[0]
	domainMessage := agent.NewMessage(agent.Role(choice.Message.Role), choice.Message.Content)

	var domainToolCalls []agent.ToolCall
	if len(choice.Message.ToolCalls) > 0 {
		domainToolCalls = make([]agent.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			domainToolCalls[i] = agent.NewToolCall(
				agent.ToolCallID(tc.ID),
				tc.Function.Name,
				tc.Function.Arguments,
			)
		}
		domainMessage = domainMessage.WithToolCalls(domainToolCalls)
	}

	return agent.NewLLMResponse(domainMessage, choice.FinishReason).WithToolCalls(domainToolCalls), nil
}

// OpenAI-compatible API types for LM Studio

type apiFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type apiToolCall struct {
	ID       string          `json:"id"`
	Function apiFunctionCall `json:"function"`
	Type     string          `json:"type"`
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatCompletionResponse struct {
	ID      string                 `json:"id"`
	Model   string                 `json:"model"`
	Object  string                 `json:"object"`
	Choices []chatCompletionChoice `json:"choices"`
	Usage   chatCompletionUsage    `json:"usage"`
	Created int64                  `json:"created"`
}

type chatCompletionChoice struct {
	FinishReason string      `json:"finish_reason"`
	Message      chatMessage `json:"message"`
	Index        int         `json:"index"`
}

type chatCompletionUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatMessage struct {
	Content    string        `json:"content"`
	Role       string        `json:"role"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []apiToolCall `json:"tool_calls,omitempty"`
}
