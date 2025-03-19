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

	// Example 1: Using max_output_tokens to limit response length
	fmt.Println("Example 1: Using max_output_tokens to limit response length")
	resp1, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.UserInputMessage("Write a detailed essay about artificial intelligence."),
			},
			// Limit the response to 100 tokens
			MaxOutputTokens: 100,
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

	// Example 2: Using instructions to set system message
	fmt.Println("Example 2: Using instructions to set system message")
	resp2, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.UserInputMessage("Tell me a joke."),
			},
			// Set custom instructions
			Instructions: "You are a comedian assistant that specializes in dad jokes.",
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp2.Choices[0].Message.Content)
	printUsage(resp2)

	fmt.Println("\n---\n")

	// Example 3: Multi-turn conversation with previous_response_id
	fmt.Println("Example 3: Multi-turn conversation with previous_response_id")

	// First turn
	fmt.Println("First turn:")
	resp3, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.UserInputMessage("What's the capital of France?"),
			},
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp3.Choices[0].Message.Content)
	printUsage(resp3)

	// Second turn using previous_response_id
	fmt.Println("\nSecond turn (using previous_response_id):")
	resp4, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.UserInputMessage("What's the population of that city?"),
			},
			// Use the previous response ID to continue the conversation
			PreviousResponseID: resp3.ID,
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp4.Choices[0].Message.Content)
	printUsage(resp4)

	// Third turn with new instructions
	fmt.Println("\nThird turn (with new instructions):")
	resp5, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.UserInputMessage("Tell me more interesting facts about this city."),
			},
			// Use the previous response ID to continue the conversation
			PreviousResponseID: resp4.ID,
			// Change the instructions for this turn
			Instructions: "You are a travel guide that provides interesting and unusual facts about cities.",
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp5.Choices[0].Message.Content)
	printUsage(resp5)
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