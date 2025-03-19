package models

import "time"

// Usage represents the usage statistics for an API request
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ResponseMessage represents a message in a response
type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseTool represents a tool that can be used in a response
type ResponseTool struct {
	Type           string               `json:"type"`
	Name           string               `json:"name,omitempty"`
	Description    string               `json:"description,omitempty"`
	Parameters     any                  `json:"parameters,omitempty"`
	Function       *ResponseToolFunction `json:"function,omitempty"`
	VectorStoreIDs []string             `json:"vector_store_ids,omitempty"`
	MaxNumResults  int                  `json:"max_num_results,omitempty"`
}

// ResponseToolFunction represents a function definition for a tool
type ResponseToolFunction struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  any      `json:"parameters"`
	VectorStoreIDs []string `json:"vector_store_ids,omitempty"`
}

// ResponseToolCall represents a tool call in a response
type ResponseToolCall struct {
	ID       string `json:"id"`
	CallID   string `json:"call_id,omitempty"` // Alias for ID, for compatibility
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ResponseChoice represents a choice in a response
type ResponseChoice struct {
	Index        int                `json:"index"`
	Message      ResponseMessage    `json:"message"`
	FinishReason string             `json:"finish_reason"`
	ToolCalls    []ResponseToolCall `json:"tool_calls,omitempty"`
}

// ResponseInputMessage represents a message in the input field
type ResponseInputMessage struct {
	Role     string `json:"role,omitempty"`
	Content  string `json:"content,omitempty"`
	Type     string `json:"type,omitempty"`
	CallID   string `json:"call_id,omitempty"`
	Output   string `json:"output,omitempty"`
}

// ResponseRequest represents a request to the Responses API
type ResponseRequest struct {
	// Model is the model to use for the response
	Model string `json:"model"`
	// Messages is the list of messages to send to the model (deprecated, use Input instead)
	Messages []ResponseMessage `json:"messages,omitempty"`
	// Input is the list of messages to send to the model
	Input []ResponseInputMessage `json:"input"`
	// Tools is the list of tools the model can use
	Tools []ResponseTool `json:"tools,omitempty"`
	// ToolChoice is the tool choice for the model
	ToolChoice any `json:"tool_choice,omitempty"`
	// Temperature is the sampling temperature to use
	Temperature float32 `json:"temperature,omitempty"`
	// TopP is the nucleus sampling parameter
	TopP float32 `json:"top_p,omitempty"`
	// N is the number of responses to generate
	N int `json:"n,omitempty"`
	// Stream indicates whether to stream the response
	Stream bool `json:"stream,omitempty"`
	// MaxTokens is the maximum number of tokens to generate (deprecated, use MaxOutputTokens instead)
	MaxTokens int `json:"max_tokens,omitempty"`
	// MaxOutputTokens is an upper bound for the number of tokens that can be generated for a response
	MaxOutputTokens int `json:"max_output_tokens,omitempty"`
	// PreviousResponseID is the unique ID of the previous response to the model, used for multi-turn conversations
	PreviousResponseID string `json:"previous_response_id,omitempty"`
	// Instructions inserts a system (or developer) message as the first item in the model's context
	Instructions string `json:"instructions,omitempty"`
	// User is the user ID for the request
	User string `json:"user,omitempty"`
	// Store indicates whether to store the response in the system
	Store bool `json:"store,omitempty"`
}

// ResponseResponse represents a response from the Responses API
type ResponseResponse struct {
	ID         string           `json:"id"`
	Object     string           `json:"object"`
	Created    int64            `json:"created"`
	Model      string           `json:"model"`
	Choices    []ResponseChoice `json:"choices"`
	Usage      *Usage           `json:"usage,omitempty"`
	OutputText string           `json:"output_text,omitempty"` // Alias for first choice's content
}

// GetOutputText returns the content of the first choice's message
func (r ResponseResponse) GetOutputText() string {
	if len(r.Choices) == 0 || r.Choices[0].Message.Content == "" {
		return ""
	}
	return r.Choices[0].Message.Content
}

// ResponseStreamChoice represents a choice in a streaming response
type ResponseStreamChoice struct {
	Index        int                 `json:"index"`
	Delta        ResponseStreamDelta `json:"delta"`
	FinishReason string              `json:"finish_reason,omitempty"`
}

// ResponseStreamDelta represents a delta in a streaming response
type ResponseStreamDelta struct {
	Role      string             `json:"role,omitempty"`
	Content   string             `json:"content,omitempty"`
	ToolCalls []ResponseToolCall `json:"tool_calls,omitempty"`
}

// ResponseStreamResponse represents a streaming response from the Responses API
type ResponseStreamResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ResponseStreamChoice `json:"choices"`
	Usage   *Usage                 `json:"usage,omitempty"`
}

// ResponseState represents the state of a response
type ResponseState struct {
	ID        string            `json:"id"`
	Object    string            `json:"object"`
	CreatedAt time.Time         `json:"created_at"`
	Messages  []ResponseMessage `json:"messages"`
}

// ResponseStateRequest represents a request to create a response state
type ResponseStateRequest struct {
	Messages []ResponseMessage `json:"messages"`
}

// ResponseStateResponse represents a response from creating a response state
type ResponseStateResponse struct {
	ID        string            `json:"id"`
	Object    string            `json:"object"`
	CreatedAt time.Time         `json:"created_at"`
	Messages  []ResponseMessage `json:"messages"`
}

// WebSearchTool represents the web search tool
type WebSearchTool struct {
	Type string `json:"type"`
}

// FileSearchTool represents the file search tool
type FileSearchTool struct {
	Type           string   `json:"type"`
	VectorStoreIDs []string `json:"vector_store_ids,omitempty"`
	MaxNumResults  int      `json:"max_num_results,omitempty"`
}

// ComputerUseTool represents the computer use tool
type ComputerUseTool struct {
	Type string `json:"type"`
}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool() ResponseTool {
	return ResponseTool{
		Type: "web_search",
	}
}

// NewFileSearchTool creates a new file search tool
func NewFileSearchTool(vectorStoreIDs []string, maxNumResults int) ResponseTool {
	return ResponseTool{
		Type:           "file_search",
		VectorStoreIDs: vectorStoreIDs,
		MaxNumResults:  maxNumResults,
	}
}

// NewFileSearchToolWithIDs creates a new file search tool with just vector store IDs
func NewFileSearchToolWithIDs(vectorStoreIDs ...string) ResponseTool {
	return ResponseTool{
		Type:           "file_search",
		VectorStoreIDs: vectorStoreIDs,
	}
}

// NewComputerUseTool creates a new computer use tool
func NewComputerUseTool() ResponseTool {
	return ResponseTool{
		Type: "computer_use",
	}
}

// NewFunctionTool creates a new function tool
func NewFunctionTool(name, description string, parameters any) ResponseTool {
	return ResponseTool{
		Type:        "function",
		Name:        name,
		Description: description,
		Parameters:  parameters,
	}
}

// UserMessage creates a new user message
func UserMessage(content string) ResponseMessage {
	return ResponseMessage{
		Role:    "user",
		Content: content,
	}
}

// SystemMessage creates a new system message
func SystemMessage(content string) ResponseMessage {
	return ResponseMessage{
		Role:    "system",
		Content: content,
	}
}

// AssistantMessage creates a new assistant message
func AssistantMessage(content string) ResponseMessage {
	return ResponseMessage{
		Role:    "assistant",
		Content: content,
	}
}

// ToolMessage creates a new tool message
func ToolMessage(content string, toolCallID string) ResponseMessage {
	return ResponseMessage{
		Role:    "tool",
		Content: content,
	}
}

// UserInputMessage creates a new user input message
func UserInputMessage(content string) ResponseInputMessage {
	return ResponseInputMessage{
		Role:    "user",
		Content: content,
	}
}

// DeveloperInputMessage creates a new developer input message
func DeveloperInputMessage(content string) ResponseInputMessage {
	return ResponseInputMessage{
		Role:    "developer",
		Content: content,
	}
}

// SystemInputMessage creates a new system input message
func SystemInputMessage(content string) ResponseInputMessage {
	return ResponseInputMessage{
		Role:    "system",
		Content: content,
	}
}

// FunctionCallOutputMessage creates a new function call output message
func FunctionCallOutputMessage(callID string, output string) ResponseInputMessage {
	return ResponseInputMessage{
		Type:   "function_call_output",
		CallID: callID,
		Output: output,
	}
}

// GetCallID returns the call_id, using ID if CallID is empty
func (tc ResponseToolCall) GetCallID() string {
	if tc.CallID != "" {
		return tc.CallID
	}
	return tc.ID
}
