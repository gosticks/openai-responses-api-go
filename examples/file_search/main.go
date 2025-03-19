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

	// Example 1: Using NewFileSearchTool with vector store IDs and max results
	fmt.Println("Example 1: Using NewFileSearchTool with vector store IDs and max results")
	resp1, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.UserInputMessage("What are the attributes of an ancient brown dragon?"),
			},
			Tools: []openairesponses.ResponseTool{
				// Specify vector store IDs and max results for file search
				openairesponses.NewFileSearchTool([]string{"vs_1234567890"}, 20),
			},
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp1.Choices[0].Message.Content)
	printUsage(resp1)

	fmt.Println("\n---\n")

	// Example 2: Using NewFileSearchToolWithIDs for simpler cases
	fmt.Println("Example 2: Using NewFileSearchToolWithIDs for simpler cases")
	resp2, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.UserInputMessage("Find information about climate change in my documents."),
			},
			Tools: []openairesponses.ResponseTool{
				// Just specify vector store IDs
				openairesponses.NewFileSearchToolWithIDs("vs_1234567890", "vs_0987654321"),
			},
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp2.Choices[0].Message.Content)
	printUsage(resp2)
}

// Helper function to print usage information
func printUsage(resp *openairesponses.ResponseResponse) {
	if resp.Usage != nil {
		fmt.Printf("\nUsage information:\n")
		fmt.Printf("  Prompt tokens: %d\n", resp.Usage.PromptTokens)
		fmt.Printf("  Completion tokens: %d\n", resp.Usage.CompletionTokens)
		fmt.Printf("  Total tokens: %d\n", resp.Usage.TotalTokens)
	}
}