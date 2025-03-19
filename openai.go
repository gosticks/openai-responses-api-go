package openairesponses

import (
	"net/http"

	"github.com/gosticks/openai-responses-api-go/client"
	"github.com/gosticks/openai-responses-api-go/models"
)

// Client is the client for the OpenAI Responses API
type Client struct {
	// Responses is the client for the Responses API
	Responses *client.Responses
}

// NewClient creates a new OpenAI Responses API client
func NewClient(apiKey string, options ...client.ClientOption) *Client {
	// Create the base client
	baseClient := client.NewClient(append([]client.ClientOption{
		client.WithAPIKey(apiKey),
	}, options...)...)

	// Create the responses client
	responsesClient := client.NewResponses(baseClient)

	return &Client{
		Responses: responsesClient,
	}
}

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) client.ClientOption {
	return client.WithBaseURL(baseURL)
}

// WithHTTPClient sets the HTTP client for the client
func WithHTTPClient(httpClient *http.Client) client.ClientOption {
	return client.WithHTTPClient(httpClient)
}

// WithUserAgent sets the user agent for the client
func WithUserAgent(userAgent string) client.ClientOption {
	return client.WithUserAgent(userAgent)
}

// WithOrganization sets the organization ID for the client
func WithOrganization(organization string) client.ClientOption {
	return client.WithOrganization(organization)
}

// Export models
type (
	// ResponseMessage represents a message in a response
	ResponseMessage = models.ResponseMessage
	// ResponseTool represents a tool that can be used in a response
	ResponseTool = models.ResponseTool
	// ResponseToolFunction represents a function definition for a tool
	ResponseToolFunction = models.ResponseToolFunction
	// ResponseToolCall represents a tool call in a response
	ResponseToolCall = models.ResponseToolCall
	// ResponseChoice represents a choice in a response
	ResponseChoice = models.ResponseChoice
	// ResponseRequest represents a request to the Responses API
	ResponseRequest = models.ResponseRequest
	// ResponseResponse represents a response from the Responses API
	ResponseResponse = models.ResponseResponse
	// ResponseStreamResponse represents a streaming response from the Responses API
	ResponseStreamResponse = models.ResponseStreamResponse
	// ResponseState represents the state of a response
	ResponseState = models.ResponseState
	// ResponseStateRequest represents a request to create a response state
	ResponseStateRequest = models.ResponseStateRequest
	// ResponseStateResponse represents a response from creating a response state
	ResponseStateResponse = models.ResponseStateResponse
	// Usage represents the usage statistics for an API request
	Usage = models.Usage
	// ResponsesStream is a stream of responses
	ResponsesStream = client.ResponsesStream
	// ResponsesStreamAccumulator accumulates streaming responses
	ResponsesStreamAccumulator = client.ResponsesStreamAccumulator
	// ResponseInputMessage represents a message in the input field
	ResponseInputMessage = models.ResponseInputMessage
)

// Export helper functions
var (
	// UserMessage creates a new user message
	UserMessage = models.UserMessage
	// SystemMessage creates a new system message
	SystemMessage = models.SystemMessage
	// AssistantMessage creates a new assistant message
	AssistantMessage = models.AssistantMessage
	// ToolMessage creates a new tool message
	ToolMessage = models.ToolMessage
	// NewWebSearchTool creates a new web search tool
	NewWebSearchTool = models.NewWebSearchTool
	// NewFileSearchTool creates a new file search tool with vector store IDs and max results
	NewFileSearchTool = models.NewFileSearchTool
	// NewFileSearchToolWithIDs creates a new file search tool with just vector store IDs
	NewFileSearchToolWithIDs = models.NewFileSearchToolWithIDs
	// NewComputerUseTool creates a new computer use tool
	NewComputerUseTool = models.NewComputerUseTool
	// NewFunctionTool creates a new function tool
	NewFunctionTool = models.NewFunctionTool
	// UserInputMessage creates a new user input message
	UserInputMessage = models.UserInputMessage
	// DeveloperInputMessage creates a new developer input message
	DeveloperInputMessage = models.DeveloperInputMessage
	// SystemInputMessage creates a new system input message
	SystemInputMessage = models.SystemInputMessage
	// FunctionCallOutputMessage creates a new function call output message
	FunctionCallOutputMessage = models.FunctionCallOutputMessage
)