package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openairesponses "github.com/gosticks/openai-responses-api-go"
)

// WeatherParams represents the parameters for the weather function
type WeatherParams struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

// getWeather is a mock function to get the weather
func getWeather(location, unit string) string {
	// In a real application, this would call a weather API
	return fmt.Sprintf("The weather in %s is sunny and 25 degrees %s.", location, unit)
}

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		os.Exit(1)
	}

	// Create a new client
	client := openairesponses.NewClient(apiKey)

	// Define the weather function parameters schema
	weatherParamsSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The city and state, e.g. San Francisco, CA",
			},
			"unit": map[string]interface{}{
				"type": "string",
				"enum": []string{"celsius", "fahrenheit"},
			},
		},
		"required": []string{"location"},
	}

	// Define tools
	tools := []openairesponses.ResponseTool{
		openairesponses.NewFunctionTool(
			"get_weather",
			"Get the current weather in a given location",
			weatherParamsSchema,
		),
	}

	// Create initial messages
	input := []openairesponses.ResponseInputMessage{
		openairesponses.DeveloperInputMessage("You are a helpful assistant with access to weather information."),
		openairesponses.UserInputMessage("What's the weather like in San Francisco and how does it compare to New York?"),
	}

	// Create a new response with a function tool
	fmt.Println("Creating initial response...")
	resp1, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model:  "gpt-4o",
			Input:  input,
			Tools:  tools,
			Store:  true,
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp1.GetOutputText())

	// Check if the model wants to call a function
	if len(resp1.Choices) > 0 && len(resp1.Choices[0].ToolCalls) > 0 {
		fmt.Println("\nProcessing tool calls...")

		// Create a new input array that includes the previous conversation
		newInput := make([]openairesponses.ResponseInputMessage, len(input))
		copy(newInput, input)

		// Process each tool call
		for _, toolCall := range resp1.Choices[0].ToolCalls {
			if toolCall.Function.Name == "get_weather" {
				// Parse the function arguments
				var params WeatherParams
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					fmt.Printf("Error parsing function arguments: %v\n", err)
					continue
				}

				// Call the function
				unit := params.Unit
				if unit == "" {
					unit = "celsius"
				}
				result := getWeather(params.Location, unit)
				fmt.Printf("Function %s returned: %s\n", toolCall.Function.Name, result)

				// Append the model's function call to the input
				newInput = append(newInput, openairesponses.ResponseInputMessage{
					Role:    "assistant",
					Content: fmt.Sprintf("I need to call the %s function to get weather information for %s.", toolCall.Function.Name, params.Location),
				})

				// Append the function call result to the input using the new format
				newInput = append(newInput, openairesponses.FunctionCallOutputMessage(
					toolCall.GetCallID(),
					result,
				))
			}
		}

		// Create a follow-up response with the function results
		fmt.Println("\nCreating follow-up response with function results...")
		resp2, err := client.Responses.Create(
			context.Background(),
			openairesponses.ResponseRequest{
				Model:  "gpt-4o",
				Input:  newInput,
				Tools:  tools,
				Store:  true,
			},
		)
		if err != nil {
			fmt.Printf("Error creating follow-up response: %v\n", err)
			os.Exit(1)
		}

		// Print the follow-up response
		fmt.Printf("\nFollow-up response: %s\n", resp2.GetOutputText())

		// Print usage information if available
		if resp2.Usage != nil {
			fmt.Printf("\nFollow-up usage information:\n")
			fmt.Printf("  Prompt tokens: %d\n", resp2.Usage.PromptTokens)
			fmt.Printf("  Completion tokens: %d\n", resp2.Usage.CompletionTokens)
			fmt.Printf("  Total tokens: %d\n", resp2.Usage.TotalTokens)
		}
	}

	// Print usage information if available
	if resp1.Usage != nil {
		fmt.Printf("\nInitial usage information:\n")
		fmt.Printf("  Prompt tokens: %d\n", resp1.Usage.PromptTokens)
		fmt.Printf("  Completion tokens: %d\n", resp1.Usage.CompletionTokens)
		fmt.Printf("  Total tokens: %d\n", resp1.Usage.TotalTokens)
	}
}