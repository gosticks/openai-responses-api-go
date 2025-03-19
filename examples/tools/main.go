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

	// Create a new response with a function tool
	resp, err := client.Responses.Create(
		context.Background(),
		openairesponses.ResponseRequest{
			Model: "gpt-4o",
			Input: []openairesponses.ResponseInputMessage{
				openairesponses.DeveloperInputMessage("You are a helpful assistant."),
				openairesponses.UserInputMessage("What's the weather like in San Francisco?"),
			},
			Tools: []openairesponses.ResponseTool{
				openairesponses.NewFunctionTool(
					"get_weather",
					"Get the current weather in a given location",
					weatherParamsSchema,
				),
			},
		},
	)
	if err != nil {
		fmt.Printf("Error creating response: %v\n", err)
		os.Exit(1)
	}

	// Check if the model wants to call a function
	if len(resp.Choices) > 0 && len(resp.Choices[0].ToolCalls) > 0 {
		// Get the function call
		toolCall := resp.Choices[0].ToolCalls[0]
		if toolCall.Function.Name == "get_weather" {
			// Parse the function arguments
			var params WeatherParams
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
				fmt.Printf("Error parsing function arguments: %v\n", err)
				os.Exit(1)
			}

			// Call the function
			unit := params.Unit
			if unit == "" {
				unit = "celsius"
			}
			result := getWeather(params.Location, unit)

			// Create a new response with the function result
			resp, err = client.Responses.Create(
				context.Background(),
				openairesponses.ResponseRequest{
					Model: "gpt-4o",
					Input: []openairesponses.ResponseInputMessage{
						openairesponses.DeveloperInputMessage("You are a helpful assistant."),
						openairesponses.UserInputMessage("What's the weather like in San Francisco?"),
					},
					// In a real application, you would include the tool calls in the messages field
				},
			)
			if err != nil {
				fmt.Printf("Error creating response: %v\n", err)
				os.Exit(1)
			}

			// Print the function result
			fmt.Printf("Function result: %s\n", result)
		}
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