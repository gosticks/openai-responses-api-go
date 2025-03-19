# OpenAI Responses API Go Client

A Go client for the OpenAI Responses API, which combines the simplicity of Chat Completions with the tool use and state management of the Assistants API.

## ⚠️ Temporary Solution Warning

**IMPORTANT**: This library is intended as a temporary solution until the official OpenAI library includes full support for the Responses API. Once the official OpenAI library releases this functionality, it's recommended to migrate to that implementation for better maintenance and official support.

This implementation aims to bridge the gap between the current capabilities of the official libraries and the new Responses API features. It may not be maintained long-term once official support is available.

## Installation

```bash
go get github.com/gosticks/openai-responses-api-go
```

## Usage

### Important Note

The OpenAI Responses API uses an `input` field that is an array of message objects, which is different from the Chat Completions API. Each message has a `role` (like "user" or "developer") and `content`.

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"os"

	openairesponses "github.com/yourusername/openai-responses"
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
}
```

### Streaming Responses

```go
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	openairesponses "github.com/yourusername/openai-responses"
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
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error receiving chunk: %v\n", err)
			os.Exit(1)
		}

		// Add the chunk to the accumulator
		accumulator.AddChunk(chunk)

		// Print the chunk
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				fmt.Print(choice.Delta.Content)
			}
		}
	}
	fmt.Println()

	// Convert the accumulator to a response
	resp := accumulator.ToResponse()

	// Print the accumulated response
	fmt.Printf("\nAccumulated response: %s\n", resp.Choices[0].Message.Content)
}
```

### Using Tools

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openairesponses "github.com/yourusername/openai-responses"
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

			// Print the function result
			fmt.Printf("Function result: %s\n", result)

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
		}
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
}
```

### Using Built-in Tools

The Responses API supports built-in tools like web search, file search, and computer use. Here's how to use them:

```go
// Create a new response with built-in tools
resp, err := client.Responses.Create(
	context.Background(),
	openairesponses.ResponseRequest{
		Model: "gpt-4o",
		Input: []openairesponses.ResponseInputMessage{
			openairesponses.DeveloperInputMessage("You are a helpful assistant."),
			openairesponses.UserInputMessage("What's the latest news about OpenAI?"),
		},
		Tools: []openairesponses.ResponseTool{
			openairesponses.NewWebSearchTool(),
		},
	},
)
```

## Response State Management

The Responses API allows you to manage the state of a conversation:

```go
// Create a new response state
stateResp, err := client.Responses.CreateState(
	context.Background(),
	openairesponses.ResponseStateRequest{
		Messages: []openairesponses.ResponseMessage{
			openairesponses.SystemMessage("You are a helpful assistant."),
			openairesponses.UserMessage("Hello, how are you today?"),
			openairesponses.AssistantMessage("I'm doing well, thank you for asking! How can I help you today?"),
		},
	},
)
if err != nil {
	fmt.Printf("Error creating response state: %v\n", err)
	os.Exit(1)
}

// Get a response state
stateResp, err = client.Responses.GetState(
	context.Background(),
	stateResp.ID,
)
if err != nil {
	fmt.Printf("Error getting response state: %v\n", err)
	os.Exit(1)
}

// Delete a response state
err = client.Responses.DeleteState(
	context.Background(),
	stateResp.ID,
)
if err != nil {
	fmt.Printf("Error deleting response state: %v\n", err)
	os.Exit(1)
}
```

## License

This library is licensed under the MIT License. See the LICENSE file for details.