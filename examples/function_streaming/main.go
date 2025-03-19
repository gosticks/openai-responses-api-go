package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gosticks/openai-responses-api-go/client"
	"github.com/gosticks/openai-responses-api-go/models"
)

func main() {
	// Create a new client
	c := client.NewClient(client.WithAPIKey(os.Getenv("OPENAI_API_KEY")))

	// Create a responses client
	responsesClient := client.NewResponses(c)

	// Create a weather function parameter schema
	weatherParamSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The city and state, e.g. San Francisco, CA",
			},
		},
		"required": []string{"location"},
	}

	// Create a weather function tool
	weatherTool := models.ResponseTool{
		Type:        "function",
		Name:        "get_weather",
		Description: "Get the current weather in a given location",
		Parameters:  weatherParamSchema,
	}

	// Create a file search tool
	fileSearchTool := models.ResponseTool{
		Type:          "file_search",
		Description:   "Search through files to find relevant information",
		VectorStoreIDs: []string{"default_store"},
		MaxNumResults: 3,
	}

	// Define the query prompt - specifically designed to trigger both tool types
	userPrompt := "What's the weather like in New York, and can you find information about weather patterns in my documents?"

	// Print what we're doing
	fmt.Println("Testing OpenAI Responses API with streaming function calls and file search")
	fmt.Printf("Prompt: %s\n\n", userPrompt)

	// Create a request
	req := models.ResponseRequest{
		Model: "gpt-4o",
		Input: []models.ResponseInputMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Tools:  []models.ResponseTool{weatherTool, fileSearchTool},
		Stream: true,
	}

	// Print the request as JSON for debugging
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	fmt.Printf("Request: %s\n\n", reqJSON)

	// Create a stream
	ctx := context.Background()
	stream, err := responsesClient.CreateStream(ctx, req)
	if err != nil {
		fmt.Printf("Error creating stream: %v\n", err)
		return
	}
	defer stream.Close()

	// Create a new accumulator
	accumulator := &client.ResponsesStreamAccumulator{}

	// Read from the stream
	fmt.Println("Streaming response:")
	for {
		resp, err := stream.Recv()
		if err != nil {
			fmt.Printf("\nStream closed with error: %v\n", err)
			break
		}

		// Add the chunk to our accumulator
		accumulator.AddChunk(resp)

		// Print the streaming update
		printStreamingUpdate(resp)
	}

	// Get the final result
	finalResponse := accumulator.ToResponse()

	// Pretty print the final response
	fmt.Println("\n\n--- Final Accumulated Response ---")
	finalJSON, _ := json.MarshalIndent(finalResponse, "", "  ")
	fmt.Printf("%s\n", finalJSON)

	// Check if the response contains tool calls
	if len(finalResponse.Choices) > 0 && len(finalResponse.Choices[0].ToolCalls) > 0 {
		fmt.Println("\n--- Tool Call Details ---")
		for i, toolCall := range finalResponse.Choices[0].ToolCalls {
			fmt.Printf("Tool Call #%d:\n", i+1)
			fmt.Printf("  ID: %s\n", toolCall.ID)
			fmt.Printf("  Type: %s\n", toolCall.Type)

			if toolCall.Type == "function" {
				fmt.Printf("  Function: %s\n", toolCall.Function.Name)
				fmt.Printf("  Arguments: %s\n", toolCall.Function.Arguments)

				// Parse the arguments for display
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err == nil {
					fmt.Printf("  Parsed Arguments: %v\n", args)
				}
			}
		}
	}
}

// printStreamingUpdate prints relevant information from a streaming chunk
func printStreamingUpdate(chunk *models.ResponseStreamResponse) {
	// Print ID and model information when available
	if chunk.ID != "" {
		fmt.Printf("\n[Response ID: %s, Model: %s]", chunk.ID, chunk.Model)
	}

	// Print usage information when available
	if chunk.Usage != nil {
		fmt.Printf("\n[Usage - Prompt: %d, Completion: %d, Total: %d]",
			chunk.Usage.PromptTokens,
			chunk.Usage.CompletionTokens,
			chunk.Usage.TotalTokens)
	}

	for _, choice := range chunk.Choices {
		// Print content delta
		if choice.Delta.Content != "" {
			fmt.Print(choice.Delta.Content)
		}

		// Print tool call information
		for _, toolCall := range choice.Delta.ToolCalls {
			// Print basic tool call info
			if toolCall.ID != "" {
				fmt.Printf("\n[Tool Call ID: %s]", toolCall.ID)
			}

			if toolCall.Type != "" {
				// Handle different tool call types
				switch toolCall.Type {
				case "function":
					fmt.Printf("\n[Function Call]")
				case "file_search":
					fmt.Printf("\n[File Search Call]")
				default:
					fmt.Printf("\n[Tool Call Type: %s]", toolCall.Type)
				}
			}

			// Print function details
			if toolCall.Function.Name != "" {
				fmt.Printf("\n[Function Name: %s]", toolCall.Function.Name)
			}

			if toolCall.Function.Arguments != "" {
				// Check if it's a complete JSON or a delta
				if isValidJSON(toolCall.Function.Arguments) {
					fmt.Printf("\n[Complete Arguments: %s]", toolCall.Function.Arguments)
				} else {
					fmt.Printf("\n[Argument Delta: %s]", toolCall.Function.Arguments)
				}
			}
		}

		// Print finish reason
		if choice.FinishReason != "" {
			fmt.Printf("\n[Finished: %s]", choice.FinishReason)
		}
	}
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}