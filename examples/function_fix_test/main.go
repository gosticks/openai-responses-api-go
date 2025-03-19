package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openairesponses "github.com/gosticks/openai-responses-api-go"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		os.Exit(1)
	}

	// Create a new client
	client := openairesponses.NewClient(apiKey)

	// Define a simple function with parameters
	weatherParamsSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The city and state, e.g. San Francisco, CA",
			},
		},
		"required": []string{"location"},
	}

	// Create a request with a function tool
	req := openairesponses.ResponseRequest{
		Model: "gpt-4o",
		Input: []openairesponses.ResponseInputMessage{
			openairesponses.UserInputMessage("What's the weather in New York?"),
		},
		Tools: []openairesponses.ResponseTool{
			openairesponses.NewFunctionTool(
				"get_weather",
				"Get the current weather in a given location",
				weatherParamsSchema,
			),
		},
	}

	// Print the request as JSON for debugging
	reqJson, _ := json.MarshalIndent(req, "", "  ")
	fmt.Println("Request JSON:")
	fmt.Println(string(reqJson))

	// Try to send the request
	fmt.Println("\nSending request to OpenAI API...")
	resp, err := client.Responses.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Response received successfully!")
	fmt.Printf("Response ID: %s\n", resp.ID)
	if len(resp.Choices) > 0 && len(resp.Choices[0].ToolCalls) > 0 {
		fmt.Println("Tool calls:")
		for _, tc := range resp.Choices[0].ToolCalls {
			fmt.Printf("  Function: %s\n", tc.Function.Name)
			fmt.Printf("  Arguments: %s\n", tc.Function.Arguments)
		}
	}
}