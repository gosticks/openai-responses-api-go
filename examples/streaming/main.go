package main

import (
	"context"
	"fmt"
	"io"
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

	// Create a new streaming response
	stream, err := client.Responses.CreateStream(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.DeveloperInputMessage("You are a helpful assistant."),
				openairesponses.UserInputMessage("Write a short poem about programming."),
			},
		},
	)
	if err != nil {
		fmt.Printf("Error creating streaming response: %v\n", err)
		os.Exit(1)
	}
	defer stream.Close()

	// Create an accumulator to accumulate the streaming responses
	accumulator := &openairesponses.ResponsesStreamAccumulator{}

	// Print the streaming response
	fmt.Println("Streaming response:")
	contentReceived := false
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("\nStream closed")
			break
		}
		if err != nil {
			fmt.Printf("Error receiving chunk: %v\n", err)
			os.Exit(1)
		}

		// Add the chunk to the accumulator
		accumulator.AddChunk(chunk)

		// Print the chunk content if available
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				fmt.Print(choice.Delta.Content)
				contentReceived = true
			}
		}
	}

	if !contentReceived {
		fmt.Println("No content was streamed.")
	}

	// Convert the accumulator to a response
	resp := accumulator.ToResponse()

	// Print the accumulated response
	if len(resp.Choices) > 0 && resp.Choices[0].Message.Content != "" {
		fmt.Printf("\nAccumulated response: %s\n", resp.Choices[0].Message.Content)
	} else {
		fmt.Println("\nNo content received in the response.")
	}

	// Print the usage information if available
	if resp.Usage != nil {
		fmt.Printf("\nUsage information:\n")
		fmt.Printf("  Prompt tokens: %d\n", resp.Usage.PromptTokens)
		fmt.Printf("  Completion tokens: %d\n", resp.Usage.CompletionTokens)
		fmt.Printf("  Total tokens: %d\n", resp.Usage.TotalTokens)
	}
}