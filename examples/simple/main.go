package main

import (
	"context"
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

	// Create a new response
	resp, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.DeveloperInputMessage("You are a helpful assistant."),
				openairesponses.UserInputMessage("Hello, how are you today?"),
			},
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)

	// Print the usage information if available
	if resp.Usage != nil {
		fmt.Printf("\nUsage information:\n")
		fmt.Printf("  Prompt tokens: %d\n", resp.Usage.PromptTokens)
		fmt.Printf("  Completion tokens: %d\n", resp.Usage.CompletionTokens)
		fmt.Printf("  Total tokens: %d\n", resp.Usage.TotalTokens)
	}
}