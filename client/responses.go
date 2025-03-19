package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gosticks/openai-responses-api-go/models"
)

const (
	responsesEndpoint      = "/responses"
	responsesStateEndpoint = "/responses/state"
)

// Responses is the client for the OpenAI Responses API
type Responses struct {
	client *Client
}

// NewResponses creates a new Responses client
func NewResponses(client *Client) *Responses {
	return &Responses{
		client: client,
	}
}

// Create creates a new response
func (r *Responses) Create(ctx context.Context, request models.ResponseRequest) (*models.ResponseResponse, error) {
	var response models.ResponseResponse
	err := r.client.post(ctx, responsesEndpoint, request, &response)
	if err != nil {
		return nil, err
	}

	// Set the OutputText field based on the first choice's content
	if len(response.Choices) > 0 && response.Choices[0].Message.Content != "" {
		response.OutputText = response.Choices[0].Message.Content
	}

	return &response, nil
}

// CreateStream creates a new streaming response
func (r *Responses) CreateStream(ctx context.Context, request models.ResponseRequest) (*ResponsesStream, error) {
	// Ensure streaming is enabled
	request.Stream = true

	// Create the request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// Construct the URL
	u := r.client.BaseURL + responsesEndpoint

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", r.client.UserAgent)
	req.Header.Set("Authorization", "Bearer "+r.client.APIKey)
	if r.client.Organization != "" {
		req.Header.Set("OpenAI-Organization", r.client.Organization)
	}

	// Make the request
	resp, err := r.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("error decoding error response: %w", err)
		}
		if errResp.Error != nil {
			errResp.Error.StatusCode = resp.StatusCode
			return nil, errResp.Error
		}
		return nil, fmt.Errorf("unknown error, status code: %d", resp.StatusCode)
	}

	return &ResponsesStream{
		reader:   bufio.NewReader(resp.Body),
		response: resp,
	}, nil
}

// CreateState creates a new response state
func (r *Responses) CreateState(ctx context.Context, request models.ResponseStateRequest) (*models.ResponseStateResponse, error) {
	var response models.ResponseStateResponse
	err := r.client.post(ctx, responsesStateEndpoint, request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetState gets a response state
func (r *Responses) GetState(ctx context.Context, id string) (*models.ResponseStateResponse, error) {
	var response models.ResponseStateResponse
	err := r.client.get(ctx, fmt.Sprintf("%s/%s", responsesStateEndpoint, id), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteState deletes a response state
func (r *Responses) DeleteState(ctx context.Context, id string) error {
	return r.client.delete(ctx, fmt.Sprintf("%s/%s", responsesStateEndpoint, id), nil)
}

// ResponsesStream is a stream of responses
type ResponsesStream struct {
	reader   *bufio.Reader
	response *http.Response
	err      error
}

// Recv receives the next response from the stream
func (s *ResponsesStream) Recv() (*models.ResponseStreamResponse, error) {
	// Check if there was a previous error
	if s.err != nil {
		return nil, s.err
	}

	// Read the next line
	line, err := s.reader.ReadString('\n')
	if err != nil {
		s.err = err
		return nil, err
	}

	// Skip empty lines
	line = strings.TrimSpace(line)
	if line == "" {
		return s.Recv()
	}

	// Check for data prefix
	const prefix = "data: "
	if !strings.HasPrefix(line, prefix) {
		return s.Recv()
	}

	// Extract the data
	data := strings.TrimPrefix(line, prefix)

	// Check for the end of the stream
	if data == "[DONE]" {
		s.err = io.EOF
		return nil, io.EOF
	}

	// Parse the new response format
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &eventData); err != nil {
		s.err = err
		return nil, err
	}

	// Create a response object
	response := &models.ResponseStreamResponse{}

	// Extract the event type
	eventType, _ := eventData["type"].(string)

	// Handle different event types
	switch eventType {
	case "response.created", "response.in_progress":
		// Extract response data
		if respData, ok := eventData["response"].(map[string]interface{}); ok {
			response.ID, _ = respData["id"].(string)
			response.Object, _ = respData["object"].(string)
			if createdAt, ok := respData["created_at"].(float64); ok {
				response.Created = int64(createdAt)
			}
			response.Model, _ = respData["model"].(string)
		}
	case "response.output_text.delta":
		// Extract delta text
		delta, _ := eventData["delta"].(string)

		// Create a choice with the delta content
		response.Choices = []models.ResponseStreamChoice{
			{
				Index: 0,
				Delta: models.ResponseStreamDelta{
					Content: delta,
				},
			},
		}

	// Output item events
	case "response.output_item.added":
		// A new output item (e.g., function call) is added
		if item, ok := eventData["item"].(map[string]interface{}); ok {
			if index, ok := eventData["output_index"].(float64); ok {
				itemType, _ := item["type"].(string)

				if itemType == "function_call" {
					id, _ := item["id"].(string)
					callID, _ := item["call_id"].(string)
					name, _ := item["name"].(string)

					toolCall := models.ResponseToolCall{
						ID:     id,
						CallID: callID,
						Type:   "function",
					}
					toolCall.Function.Name = name

					response.Choices = []models.ResponseStreamChoice{
						{
							Index: int(index),
							Delta: models.ResponseStreamDelta{
								ToolCalls: []models.ResponseToolCall{toolCall},
							},
						},
					}
				}
			}
		}

	case "response.output_item.done":
		// An output item has completed
		if item, ok := eventData["item"].(map[string]interface{}); ok {
			if index, ok := eventData["output_index"].(float64); ok {
				itemType, _ := item["type"].(string)

				if itemType == "function_call" {
					id, _ := item["id"].(string)
					callID, _ := item["call_id"].(string)
					name, _ := item["name"].(string)
					arguments, _ := item["arguments"].(string)

					toolCall := models.ResponseToolCall{
						ID:     id,
						CallID: callID,
						Type:   "function",
					}
					toolCall.Function.Name = name
					toolCall.Function.Arguments = arguments

					response.Choices = []models.ResponseStreamChoice{
						{
							Index: int(index),
							Delta: models.ResponseStreamDelta{
								ToolCalls: []models.ResponseToolCall{toolCall},
							},
						},
					}
				}
			}
		}

	// File search related events
	case "response.file_search_call.in_progress",
		"response.file_search_call.searching",
		"response.file_search_call.completed":
		// A file search is in progress or completed
		if index, ok := eventData["output_index"].(float64); ok {
			if itemID, ok := eventData["item_id"].(string); ok {
				response.Choices = []models.ResponseStreamChoice{
					{
						Index: int(index),
						Delta: models.ResponseStreamDelta{
							// Create a tool call for file search
							ToolCalls: []models.ResponseToolCall{
								{
									ID:   itemID,
									Type: "file_search", // Use file_search as type
								},
							},
						},
					},
				}
			}
		}

	// Function call events
	case "response.tool_call.created", "response.tool_call.in_progress":
		// A tool call is being created
		if index, ok := eventData["output_index"].(float64); ok {
			response.Choices = []models.ResponseStreamChoice{
				{
					Index: int(index),
					Delta: models.ResponseStreamDelta{
						// An empty delta to indicate a tool call is being created
						ToolCalls: []models.ResponseToolCall{
							{
								// Empty tool call to be populated with subsequent events
							},
						},
					},
				},
			}
		}
	case "response.tool_call.id":
		// Get the tool call ID
		if toolCallID, ok := eventData["id"].(string); ok {
			if index, ok := eventData["output_index"].(float64); ok {
				response.Choices = []models.ResponseStreamChoice{
					{
						Index: int(index),
						Delta: models.ResponseStreamDelta{
							ToolCalls: []models.ResponseToolCall{
								{
									ID: toolCallID,
								},
							},
						},
					},
				}
			}
		}
	// Function call argument events
	case "response.function_call_arguments.delta":
		// Get the function arguments delta
		if delta, ok := eventData["delta"].(string); ok {
			if index, ok := eventData["output_index"].(float64); ok {
				if itemID, ok := eventData["item_id"].(string); ok {
					toolCall := models.ResponseToolCall{
						ID: itemID,
					}
					toolCall.Function.Arguments = delta

					response.Choices = []models.ResponseStreamChoice{
						{
							Index: int(index),
							Delta: models.ResponseStreamDelta{
								ToolCalls: []models.ResponseToolCall{toolCall},
							},
						},
					}
				}
			}
		}
	case "response.function_call_arguments.done":
		// Get the complete function arguments
		if arguments, ok := eventData["arguments"].(string); ok {
			if index, ok := eventData["output_index"].(float64); ok {
				if itemID, ok := eventData["item_id"].(string); ok {
					toolCall := models.ResponseToolCall{
						ID: itemID,
					}
					toolCall.Function.Arguments = arguments

					response.Choices = []models.ResponseStreamChoice{
						{
							Index: int(index),
							Delta: models.ResponseStreamDelta{
								ToolCalls: []models.ResponseToolCall{toolCall},
							},
						},
					}
				}
			}
		}

	case "response.completed", "response.incomplete":
		// Extract usage data if available
		if respData, ok := eventData["response"].(map[string]interface{}); ok {
			response.ID, _ = respData["id"].(string)
			response.Object, _ = respData["object"].(string)
			if createdAt, ok := respData["created_at"].(float64); ok {
				response.Created = int64(createdAt)
			}
			response.Model, _ = respData["model"].(string)

			if usageData, ok := respData["usage"].(map[string]interface{}); ok {
				promptTokens, _ := usageData["prompt_tokens"].(float64)
				completionTokens, _ := usageData["completion_tokens"].(float64)
				totalTokens, _ := usageData["total_tokens"].(float64)

				response.Usage = &models.Usage{
					PromptTokens:     int(promptTokens),
					CompletionTokens: int(completionTokens),
					TotalTokens:      int(totalTokens),
				}
			}
		}

		// Signal that this is the end of the stream
		s.err = io.EOF
	}

	// Skip events that don't contain useful data for our client
	if len(response.Choices) == 0 && response.ID == "" && response.Usage == nil {
		return s.Recv()
	}

	return response, nil
}

// Close closes the stream
func (s *ResponsesStream) Close() error {
	if s.response != nil && s.response.Body != nil {
		return s.response.Body.Close()
	}
	return nil
}

// Err returns the last error that occurred while reading from the stream
func (s *ResponsesStream) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// ResponsesStreamAccumulator accumulates streaming responses
type ResponsesStreamAccumulator struct {
	ID      string
	Object  string
	Created int64
	Model   string
	Choices []models.ResponseChoice
	Usage   *models.Usage
}

// AddChunk adds a chunk to the accumulator
func (a *ResponsesStreamAccumulator) AddChunk(chunk *models.ResponseStreamResponse) {
	// Initialize the accumulator if this is the first chunk with an ID
	if a.ID == "" && chunk.ID != "" {
		a.ID = chunk.ID
		a.Object = chunk.Object
		a.Created = chunk.Created
		a.Model = chunk.Model

		// Initialize choices if there are any in the chunk
		if len(chunk.Choices) > 0 {
			a.Choices = make([]models.ResponseChoice, len(chunk.Choices))
			for i := range chunk.Choices {
				a.Choices[i] = models.ResponseChoice{
					Index:        chunk.Choices[i].Index,
					Message:      models.ResponseMessage{Role: "assistant"},
					FinishReason: "",
				}
			}
		} else {
			// Initialize with at least one choice for content
			a.Choices = []models.ResponseChoice{
				{
					Index:        0,
					Message:      models.ResponseMessage{Role: "assistant"},
					FinishReason: "",
				},
			}
		}
	}

	// If this chunk has usage data, store it
	if chunk.Usage != nil {
		a.Usage = chunk.Usage
	}

	// Ensure we have at least one choice for content
	if len(a.Choices) == 0 && len(chunk.Choices) > 0 {
		a.Choices = []models.ResponseChoice{
			{
				Index:        0,
				Message:      models.ResponseMessage{Role: "assistant"},
				FinishReason: "",
			},
		}
	}

	// Update the accumulator with the chunk data
	for _, choice := range chunk.Choices {
		// Ensure we have enough choices
		for len(a.Choices) <= choice.Index {
			a.Choices = append(a.Choices, models.ResponseChoice{
				Index:        len(a.Choices),
				Message:      models.ResponseMessage{Role: "assistant"},
				FinishReason: "",
			})
		}

		// Update the message
		if choice.Delta.Role != "" {
			a.Choices[choice.Index].Message.Role = choice.Delta.Role
		}
		if choice.Delta.Content != "" {
			a.Choices[choice.Index].Message.Content += choice.Delta.Content
		}

		// Update the tool calls
		if len(choice.Delta.ToolCalls) > 0 {
			for _, toolCallDelta := range choice.Delta.ToolCalls {
				// If our choice doesn't have tool calls yet, initialize the slice
				if a.Choices[choice.Index].ToolCalls == nil {
					a.Choices[choice.Index].ToolCalls = []models.ResponseToolCall{}
				}

				// Find if we already have this tool call
				toolCallIndex := -1
				for i, existingToolCall := range a.Choices[choice.Index].ToolCalls {
					if existingToolCall.ID == toolCallDelta.ID {
						toolCallIndex = i
						break
					}
				}

				// If we don't have this tool call yet, add it
				if toolCallIndex == -1 {
					a.Choices[choice.Index].ToolCalls = append(a.Choices[choice.Index].ToolCalls, models.ResponseToolCall{
						ID:     toolCallDelta.ID,
						Type:   toolCallDelta.Type,
						CallID: toolCallDelta.CallID,
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{
							Name:      toolCallDelta.Function.Name,
							Arguments: toolCallDelta.Function.Arguments,
						},
					})
				} else {
					// Update existing tool call
					if toolCallDelta.Type != "" {
						a.Choices[choice.Index].ToolCalls[toolCallIndex].Type = toolCallDelta.Type
					}
					if toolCallDelta.Function.Name != "" {
						a.Choices[choice.Index].ToolCalls[toolCallIndex].Function.Name = toolCallDelta.Function.Name
					}
					if toolCallDelta.Function.Arguments != "" {
						a.Choices[choice.Index].ToolCalls[toolCallIndex].Function.Arguments = toolCallDelta.Function.Arguments
					}
				}
			}
		}

		// Update the finish reason
		if choice.FinishReason != "" {
			a.Choices[choice.Index].FinishReason = choice.FinishReason
		}
	}
}

// ToResponse converts the accumulator to a response
func (a *ResponsesStreamAccumulator) ToResponse() *models.ResponseResponse {
	choices := make([]models.ResponseChoice, len(a.Choices))
	for i, choice := range a.Choices {
		choices[i] = models.ResponseChoice{
			Index:        choice.Index,
			Message:      choice.Message,
			FinishReason: choice.FinishReason,
		}

		// Include tool calls if present
		if len(choice.ToolCalls) > 0 {
			choices[i].ToolCalls = choice.ToolCalls
		}
	}

	return &models.ResponseResponse{
		ID:      a.ID,
		Object:  a.Object,
		Created: a.Created,
		Model:   a.Model,
		Choices: choices,
		Usage:   a.Usage,
	}
}
